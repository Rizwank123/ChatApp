package websockets

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gofrs/uuid/v5"
)

type MessageEnvelope struct {
	SenderID   string
	ReceiverID string
	Message    []byte
}

// ChatAuthorizer validates whether two users have active mutual follow permission to exchange messages.
type ChatAuthorizer interface {
	CanChat(ctx context.Context, userA, userB uuid.UUID) (bool, error)
}

type UserEventBroadcaster interface {
	BroadcastToUser(ctx context.Context, recipientID uuid.UUID, event string, data any) error
}

type Hub struct {
	Clients    map[string]map[*Client]struct{}
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan MessageEnvelope
	Authorizer ChatAuthorizer
	done       chan struct{}
	stopOnce   sync.Once
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]map[*Client]struct{}),
		Register:   make(chan *Client, 128),
		Unregister: make(chan *Client, 128),
		Broadcast:  make(chan MessageEnvelope, 256),
		done:       make(chan struct{}),
	}
}

// SetAuthorizer configures chat authorization checks on the hub.
func (h *Hub) SetAuthorizer(auth ChatAuthorizer) {
	h.Authorizer = auth
}

func (h *Hub) RegisterClient(client *Client) error {
	select {
	case h.Register <- client:
		return nil
	case <-h.done:
		return context.Canceled
	}
}

func (h *Hub) UnregisterClient(client *Client) {
	select {
	case h.Unregister <- client:
	case <-h.done:
	}
}

func (h *Hub) Publish(ctx context.Context, envelope MessageEnvelope) error {
	select {
	case h.Broadcast <- envelope:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-h.done:
		return context.Canceled
	}
}

func (h *Hub) BroadcastToUser(ctx context.Context, recipientID uuid.UUID, event string, data any) error {
	payload, err := json.Marshal(map[string]any{"type": event, "content": data})
	if err != nil {
		return err
	}
	return h.Publish(ctx, MessageEnvelope{ReceiverID: recipientID.String(), Message: payload})
}

func (h *Hub) Run(ctx context.Context) {
	defer h.stopOnce.Do(func() {
		for _, clients := range h.Clients {
			for client := range clients {
				close(client.Send)
				_ = client.Conn.Close()
			}
		}
		clear(h.Clients)
		close(h.done)
	})

	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.Register:
			log.Printf("New WebSocket client registered. ID: %s", client.ID)
			if h.Clients[client.ID] == nil {
				h.Clients[client.ID] = make(map[*Client]struct{})
			}
			h.Clients[client.ID][client] = struct{}{}

		case client := <-h.Unregister:
			log.Printf("Client unregistered. ID: %s", client.ID)
			if clients, ok := h.Clients[client.ID]; ok {
				if _, registered := clients[client]; registered {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.Clients, client.ID)
				}
			}

		case envelope := <-h.Broadcast:
			if h.Authorizer != nil && envelope.SenderID != "" && envelope.ReceiverID != "" && envelope.SenderID != envelope.ReceiverID {
				senderUUID, err1 := uuid.FromString(envelope.SenderID)
				receiverUUID, err2 := uuid.FromString(envelope.ReceiverID)
				if err1 == nil && err2 == nil {
					canChat, err := h.Authorizer.CanChat(ctx, senderUUID, receiverUUID)
					if err != nil || !canChat {
						log.Printf("WebSocket transmission blocked between %s and %s: chat authorization failed", envelope.SenderID, envelope.ReceiverID)
						for sender := range h.Clients[envelope.SenderID] {
							errPayload, _ := json.Marshal(map[string]string{
								"type": "error", "error_code": "CHAT_NOT_ALLOWED",
								"message": "Messaging is not allowed for this relationship.",
							})
							select {
							case sender.Send <- errPayload:
							default:
							}
						}
						continue
					}
				}
			}

			if clients, ok := h.Clients[envelope.ReceiverID]; ok {
				for receiver := range clients {
					select {
					case receiver.Send <- envelope.Message:
					default:
						delete(clients, receiver)
						close(receiver.Send)
						_ = receiver.Conn.Close()
					}
				}
				if len(clients) == 0 {
					delete(h.Clients, envelope.ReceiverID)
				}
			}
		}
	}
}
