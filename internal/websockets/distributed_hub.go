package websockets

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// ClientConnection represents a single connected WebSocket client.
type ClientConnection struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Conn     *websocket.Conn
	SendChan chan []byte
	Hub      *DistributedHub
}

// WSMessage represents outgoing and incoming JSON payloads over the socket.
type WSMessage struct {
	Action         string          `json:"action,omitempty"`
	Event          string          `json:"event,omitempty"`
	ConversationID *uuid.UUID      `json:"conversation_id,omitempty"`
	RecipientID    *uuid.UUID      `json:"recipient_id,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	Timestamp      int64           `json:"timestamp"`
}

// DistributedHub manages active WebSockets across multi-node servers using Redis Pub/Sub.
type DistributedHub struct {
	redisClient *redis.Client
	clients     map[uuid.UUID]map[*ClientConnection]bool // UserID -> set of connections
	register    chan *ClientConnection
	unregister  chan *ClientConnection
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewDistributedHub instantiates a thread-safe distributed hub.
func NewDistributedHub(rdb *redis.Client) *DistributedHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &DistributedHub{
		redisClient: rdb,
		clients:     make(map[uuid.UUID]map[*ClientConnection]bool),
		register:    make(chan *ClientConnection, 256),
		unregister:  make(chan *ClientConnection, 256),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Run begins the hub event loops and Redis subscriber listener.
func (h *DistributedHub) Run() {
	go h.listenRedisPubSub()

	for {
		select {
		case <-h.ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*ClientConnection]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("[WS Hub] Registered connection for user %s (total sockets for user: %d)",
				client.UserID, len(h.clients[client.UserID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.UserID]; ok {
				if _, found := conns[client]; found {
					delete(conns, client)
					close(client.SendChan)
					if len(conns) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS Hub] Unregistered connection for user %s", client.UserID)
		}
	}
}

// listenRedisPubSub subscribes to global cross-server WebSocket broadcast channel.
func (h *DistributedHub) listenRedisPubSub() {
	if h.redisClient == nil {
		return
	}
	pubsub := h.redisClient.Subscribe(h.ctx, "ws:broadcast:users")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-h.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var wsMsg WSMessage
			if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
				continue
			}

			if wsMsg.RecipientID != nil {
				h.DeliverDirectly(*wsMsg.RecipientID, []byte(msg.Payload))
			}
		}
	}
}

// BroadcastToUser sends an event to a user across instances via Redis Pub/Sub.
func (h *DistributedHub) BroadcastToUser(ctx context.Context, recipientID uuid.UUID, event string, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	wsMsg := WSMessage{
		Event:       event,
		RecipientID: &recipientID,
		Data:        raw,
		Timestamp:   time.Now().UnixMilli(),
	}

	payload, err := json.Marshal(wsMsg)
	if err != nil {
		return err
	}

	// Deliver to local node connections first
	h.DeliverDirectly(recipientID, payload)

	// Publish to Redis channel so other server instances deliver if user is connected elsewhere
	if h.redisClient != nil {
		return h.redisClient.Publish(ctx, "ws:broadcast:users", payload).Err()
	}
	return nil
}

// DeliverDirectly writes the raw frame directly into any active local client socket buffers.
func (h *DistributedHub) DeliverDirectly(userID uuid.UUID, msg []byte) {
	h.mu.RLock()
	conns, exists := h.clients[userID]
	if !exists || len(conns) == 0 {
		h.mu.RUnlock()
		return
	}

	for client := range conns {
		select {
		case client.SendChan <- msg:
		default:
			// Client channel full/lagging, drop message or mark for disconnect
			log.Printf("[WS Hub] Slow client dropped message: user %s", userID)
		}
	}
	h.mu.RUnlock()
}

// RegisterClient queues client registration.
func (h *DistributedHub) RegisterClient(c *ClientConnection) {
	h.register <- c
}

// UnregisterClient queues client removal.
func (h *DistributedHub) UnregisterClient(c *ClientConnection) {
	h.unregister <- c
}

// Shutdown stops the hub and terminates all active connections.
func (h *DistributedHub) Shutdown() {
	h.cancel()
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, conns := range h.clients {
		for c := range conns {
			close(c.SendChan)
			_ = c.Conn.Close()
		}
	}
	h.clients = make(map[uuid.UUID]map[*ClientConnection]bool)
}
