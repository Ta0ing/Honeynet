package ai

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultSettingsID = "00000000-0000-4000-8000-000000000001"

var ErrInvalidSettings = errors.New("invalid AI settings")

type StoredSettings struct {
	Config   Config
	Revision int64
}

type SettingsView struct {
	Enabled        bool   `json:"enabled"`
	Configured     bool   `json:"configured"`
	HasAPIKey      bool   `json:"has_api_key"`
	Provider       string `json:"provider"`
	BaseURL        string `json:"base_url"`
	Model          string `json:"model"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	SendRawPacket  bool   `json:"send_raw_packet"`
	Revision       int64  `json:"revision"`
}

type SettingsUpdate struct {
	Enabled        *bool
	Provider       *string
	BaseURL        *string
	APIKey         *string
	Model          *string
	TimeoutSeconds *int
	SendRawPacket  *bool
	ClearAPIKey    bool
}

// SettingsStore persists the online configuration. The process-local mutex
// prevents two concurrent API updates from merging against stale secrets;
// the database lock preserves the same invariant if transaction behavior is
// later shared by multiple Server instances.
type SettingsStore struct {
	db  *gorm.DB
	key [sha256.Size]byte
	mu  sync.Mutex
}

func NewSettingsStore(db *gorm.DB, masterSecret string) (*SettingsStore, error) {
	if db == nil {
		return nil, errors.New("AI settings database is required")
	}
	if strings.TrimSpace(masterSecret) == "" {
		return nil, errors.New("AI settings encryption secret is required")
	}
	key := sha256.Sum256([]byte("honeynet/ai-settings/v1\x00" + masterSecret))
	return &SettingsStore{db: db, key: key}, nil
}

func (s *SettingsStore) LoadOrCreate(ctx context.Context, bootstrap Config) (StoredSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var item store.AISetting
	result := s.db.WithContext(ctx).Where("id = ?", defaultSettingsID).Limit(1).Find(&item)
	if result.Error != nil {
		return StoredSettings{}, result.Error
	}
	if result.RowsAffected > 0 {
		return s.decode(item)
	}
	normalized, err := NormalizeConfig(bootstrap)
	if err != nil {
		return StoredSettings{}, err
	}
	ciphertext, err := sealAPIKey(s.key, normalized.APIKey)
	if err != nil {
		return StoredSettings{}, err
	}
	item = store.AISetting{
		Base: store.Base{ID: defaultSettingsID}, Enabled: normalized.Enabled, Provider: normalized.Provider,
		BaseURL: normalized.BaseURL, Model: normalized.Model, TimeoutSeconds: int(normalized.Timeout / time.Second),
		SendRawPacket: normalized.SendRawPacket, APIKeyCiphertext: ciphertext, Revision: 1,
	}
	if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
		return StoredSettings{}, fmt.Errorf("create AI settings: %w", err)
	}
	return StoredSettings{Config: normalized, Revision: item.Revision}, nil
}

func (s *SettingsStore) Load(ctx context.Context) (StoredSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var item store.AISetting
	result := s.db.WithContext(ctx).Where("id = ?", defaultSettingsID).Limit(1).Find(&item)
	if result.Error != nil {
		return StoredSettings{}, result.Error
	}
	if result.RowsAffected == 0 {
		return StoredSettings{}, gorm.ErrRecordNotFound
	}
	return s.decode(item)
}

func (s *SettingsStore) Update(ctx context.Context, update SettingsUpdate) (StoredSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return StoredSettings{}, tx.Error
	}
	rollback := func(err error) (StoredSettings, error) {
		_ = tx.Rollback().Error
		return StoredSettings{}, err
	}
	var item store.AISetting
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", defaultSettingsID).Limit(1).Find(&item)
	if result.Error != nil {
		return rollback(result.Error)
	}
	if result.RowsAffected == 0 {
		return rollback(gorm.ErrRecordNotFound)
	}
	stored, err := s.decode(item)
	if err != nil {
		return rollback(err)
	}
	candidate, err := applySettingsUpdate(stored.Config, update)
	if err != nil {
		return rollback(err)
	}
	ciphertext, err := sealAPIKey(s.key, candidate.APIKey)
	if err != nil {
		return rollback(err)
	}
	revision := item.Revision + 1
	updates := map[string]any{
		"enabled": candidate.Enabled, "provider": candidate.Provider, "base_url": candidate.BaseURL,
		"model": candidate.Model, "timeout_seconds": int(candidate.Timeout / time.Second),
		"send_raw_packet": candidate.SendRawPacket, "api_key_ciphertext": ciphertext, "revision": revision,
	}
	if err := tx.Model(&item).Updates(updates).Error; err != nil {
		return rollback(err)
	}
	if err := tx.Commit().Error; err != nil {
		return StoredSettings{}, err
	}
	return StoredSettings{Config: candidate, Revision: revision}, nil
}

func applySettingsUpdate(current Config, update SettingsUpdate) (Config, error) {
	if update.ClearAPIKey && update.APIKey != nil {
		value := strings.TrimSpace(*update.APIKey)
		if value != "" && value != SecretMask {
			return Config{}, fmt.Errorf("%w: 不能同时设置并清除 API Key", ErrInvalidSettings)
		}
	}
	candidate := current
	if update.Enabled != nil {
		candidate.Enabled = *update.Enabled
	}
	if update.Provider != nil {
		candidate.Provider = *update.Provider
	}
	if update.BaseURL != nil {
		candidate.BaseURL = *update.BaseURL
	}
	if update.Model != nil {
		candidate.Model = *update.Model
	}
	if update.TimeoutSeconds != nil {
		candidate.Timeout = time.Duration(*update.TimeoutSeconds) * time.Second
	}
	if update.SendRawPacket != nil {
		candidate.SendRawPacket = *update.SendRawPacket
	}
	if update.APIKey != nil {
		value := strings.TrimSpace(*update.APIKey)
		if value != "" && value != SecretMask {
			candidate.APIKey = value
		}
	}
	if update.ClearAPIKey {
		candidate.APIKey = ""
	}
	normalized, err := NormalizeConfig(candidate)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrInvalidSettings, err)
	}
	return normalized, nil
}

func (settings StoredSettings) View() SettingsView {
	status := statusFromConfig(settings.Config)
	return SettingsView{
		Enabled: status.Enabled, Configured: status.Configured, HasAPIKey: status.HasAPIKey,
		Provider: status.Provider, BaseURL: status.BaseURL, Model: status.Model,
		TimeoutSeconds: status.TimeoutSeconds, SendRawPacket: status.SendRawPacket, Revision: settings.Revision,
	}
}

func (s *SettingsStore) decode(item store.AISetting) (StoredSettings, error) {
	apiKey, err := openAPIKey(s.key, item.APIKeyCiphertext)
	if err != nil {
		return StoredSettings{}, fmt.Errorf("decrypt AI API Key: %w", err)
	}
	config, err := NormalizeConfig(Config{
		Enabled: item.Enabled, Provider: item.Provider, BaseURL: item.BaseURL, APIKey: apiKey, Model: item.Model,
		Timeout: time.Duration(item.TimeoutSeconds) * time.Second, SendRawPacket: item.SendRawPacket,
	})
	if err != nil {
		return StoredSettings{}, fmt.Errorf("stored AI settings are invalid: %w", err)
	}
	return StoredSettings{Config: config, Revision: item.Revision}, nil
}

func sealAPIKey(key [sha256.Size]byte, value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, []byte(value), []byte("honeynet-ai-api-key-v1"))
	result := make([]byte, 1, 1+len(nonce)+len(sealed))
	result[0] = 1
	result = append(result, nonce...)
	return append(result, sealed...), nil
}

func openAPIKey(key [sha256.Size]byte, value []byte) (string, error) {
	if len(value) == 0 {
		return "", nil
	}
	if value[0] != 1 {
		return "", errors.New("unsupported ciphertext version")
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(value) <= 1+gcm.NonceSize() {
		return "", errors.New("invalid ciphertext")
	}
	nonce := value[1 : 1+gcm.NonceSize()]
	plaintext, err := gcm.Open(nil, nonce, value[1+gcm.NonceSize():], []byte("honeynet-ai-api-key-v1"))
	if err != nil {
		return "", errors.New("invalid ciphertext authentication")
	}
	return string(plaintext), nil
}
