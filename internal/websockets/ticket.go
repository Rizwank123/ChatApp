package websockets

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
)

const webSocketTicketTTL = 30 * time.Second

// TicketManager issues short-lived, single-use credentials for WebSocket
// handshakes. Access tokens are never placed in WebSocket URLs.
type TicketManager struct {
	redis *redis.Client
	mu    sync.Mutex
	local map[string]string
}

func NewTicketManager(client *redis.Client) *TicketManager {
	return &TicketManager{redis: client, local: make(map[string]string)}
}

func (m *TicketManager) Issue(ctx context.Context, personnelID uuid.UUID) (string, time.Time, error) {
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(tokenBytes[:])
	expiresAt := time.Now().Add(webSocketTicketTTL)
	if m.redis != nil {
		if err := m.redis.Set(ctx, "ws:ticket:"+token, personnelID.String(), webSocketTicketTTL).Err(); err != nil {
			return "", time.Time{}, fmt.Errorf("store WebSocket ticket: %w", err)
		}
	} else {
		m.mu.Lock()
		m.local[token] = personnelID.String()
		m.mu.Unlock()
	}
	return token, expiresAt, nil
}

func (m *TicketManager) Consume(ctx context.Context, token string) (uuid.UUID, bool) {
	var personnelID string
	var err error
	if m.redis != nil {
		personnelID, err = m.redis.GetDel(ctx, "ws:ticket:"+token).Result()
	} else {
		m.mu.Lock()
		personnelID = m.local[token]
		delete(m.local, token)
		m.mu.Unlock()
	}
	if err != nil || personnelID == "" {
		return uuid.Nil, false
	}
	parsed, err := uuid.FromString(personnelID)
	return parsed, err == nil
}
