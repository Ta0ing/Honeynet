package httpapi

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/gorm"
)

const userContextKey = "auth.user"

type AuthUser struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	TokenVersion uint64 `json:"-"`
}

type Claims struct {
	AuthUser
	Version uint64 `json:"ver"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret  []byte
	expires time.Duration
	db      *gorm.DB
}

func NewTokenManager(secret string, expires time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), expires: expires}
}

// WithUserStore makes every authenticated request validate the account's live
// state. This is deliberately opt-in so TokenManager remains usable by small
// signing tests, while production always wires the business database.
func (m *TokenManager) WithUserStore(db *gorm.DB) *TokenManager {
	m.db = db
	return m
}

func (m *TokenManager) Issue(user AuthUser) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.expires)
	claims := Claims{AuthUser: user, Version: user.TokenVersion, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: "honeynet-server", Subject: user.ID,
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expiresAt),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, expiresAt, err
}

func (m *TokenManager) Parse(raw string) (AuthUser, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer("honeynet-server"))
	if err != nil || !token.Valid {
		return AuthUser{}, errors.New("invalid token")
	}
	if claims.AuthUser.ID == "" || claims.Subject != claims.AuthUser.ID {
		return AuthUser{}, errors.New("invalid token subject")
	}
	user := claims.AuthUser
	user.TokenVersion = claims.Version
	if m.db == nil {
		return user, nil
	}
	var stored store.User
	if err := m.db.Select("id", "username", "role", "enabled", "token_version").First(&stored, "id = ?", user.ID).Error; err != nil {
		return AuthUser{}, errors.New("account no longer exists")
	}
	if !stored.Enabled || stored.TokenVersion != claims.Version || stored.Username != user.Username || stored.Role != user.Role {
		return AuthUser{}, errors.New("token has been revoked")
	}
	return AuthUser{ID: stored.ID, Username: stored.Username, Role: stored.Role, TokenVersion: stored.TokenVersion}, nil
}

func (m *TokenManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			fail(c, 401, "UNAUTHORIZED", "请先登录")
			c.Abort()
			return
		}
		user, err := m.Parse(raw)
		if err != nil {
			fail(c, 401, "TOKEN_INVALID", "登录状态已失效")
			c.Abort()
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	scheme, raw, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	raw = strings.TrimSpace(raw)
	return raw, raw != "" && !strings.ContainsAny(raw, " \t\r\n")
}

const (
	loginIPFailureLimit = 12
	loginFailureWindow  = 15 * time.Minute
	loginIPLockDuration = 15 * time.Minute
	accountFailureLimit = 6
	accountLockDuration = 15 * time.Minute
)

type loginAttempt struct {
	Failures    int
	WindowStart time.Time
	LockedUntil time.Time
}

// loginAttemptLimiter bounds password attempts even for unknown usernames.
// Per-account locks are persisted on store.User; this in-memory IP guard is an
// additional DoS-resistant first line and does not contain credentials.
type loginAttemptLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
	now      func() time.Time
}

func newLoginAttemptLimiter() *loginAttemptLimiter {
	return &loginAttemptLimiter{attempts: map[string]loginAttempt{}, now: time.Now}
}

func (l *loginAttemptLimiter) allow(key string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	attempt, exists := l.attempts[key]
	if !exists {
		return true
	}
	if !attempt.LockedUntil.IsZero() && now.Before(attempt.LockedUntil) {
		return false
	}
	if now.Sub(attempt.WindowStart) >= loginFailureWindow {
		delete(l.attempts, key)
	}
	return true
}

func (l *loginAttemptLimiter) failure(key string) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	attempt := l.attempts[key]
	if attempt.WindowStart.IsZero() || now.Sub(attempt.WindowStart) >= loginFailureWindow {
		attempt = loginAttempt{WindowStart: now}
	}
	attempt.Failures++
	if attempt.Failures >= loginIPFailureLimit {
		attempt.LockedUntil = now.Add(loginIPLockDuration)
	}
	l.attempts[key] = attempt
	if len(l.attempts) > 4096 {
		for candidate, value := range l.attempts {
			if now.Sub(value.WindowStart) >= loginFailureWindow && now.After(value.LockedUntil) {
				delete(l.attempts, candidate)
			}
		}
	}
	return !attempt.LockedUntil.IsZero() && now.Before(attempt.LockedUntil)
}

func (l *loginAttemptLimiter) success(key string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	delete(l.attempts, key)
	l.mu.Unlock()
}

func currentUser(c *gin.Context) AuthUser {
	value, _ := c.Get(userContextKey)
	user, _ := value.(AuthUser)
	return user
}

func requireRoles(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		if !allowed[currentUser(c).Role] {
			fail(c, 403, "FORBIDDEN", "当前账号无权执行此操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
