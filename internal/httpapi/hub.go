package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
	TS      int64  `json:"ts"`
}

type wsClient struct {
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
}

func NewHub() *Hub { return &Hub{clients: make(map[*wsClient]struct{})} }

func (h *Hub) Broadcast(messageType string, payload any) {
	data, err := json.Marshal(WSMessage{Type: messageType, Payload: payload, TS: time.Now().Unix()})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

// RevokeUser terminates already-upgraded WebSocket sessions after logout or
// an authentication-relevant account change. Without this, a revoked bearer
// token could continue receiving live threat data until its socket happened
// to disconnect.
func (h *Hub) RevokeUser(userID string) {
	if h == nil || userID == "" {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.userID == userID {
			_ = client.conn.Close()
		}
	}
}

func (h *Hub) Serve(c *gin.Context, tokens *TokenManager) {
	var raw string
	for _, protocol := range strings.Split(c.GetHeader("Sec-WebSocket-Protocol"), ",") {
		protocol = strings.TrimSpace(protocol)
		if strings.Count(protocol, ".") == 2 {
			raw = protocol
			break
		}
	}
	user, err := tokens.Parse(raw)
	if err != nil {
		fail(c, 401, "TOKEN_INVALID", "登录状态已失效")
		return
	}
	upgrader := websocket.Upgrader{Subprotocols: []string{raw}, CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || strings.Contains(origin, r.Host)
	}}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &wsClient{conn: conn, send: make(chan []byte, 64), userID: user.ID}
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	go client.writeLoop()
	client.readLoop()
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()
	close(client.send)
	_ = conn.Close()
}

func (c *wsClient) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if c.conn.WriteMessage(websocket.TextMessage, message) != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if c.conn.WriteMessage(websocket.PingMessage, nil) != nil {
				return
			}
		}
	}
}

func (c *wsClient) readLoop() {
	c.conn.SetReadLimit(1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(70 * time.Second)) })
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
