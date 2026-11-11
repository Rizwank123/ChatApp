package websockets

import (
	"context"
	"encoding/json"
	"log"

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

type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan MessageEnvelope
	Authorizer ChatAuthorizer
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan MessageEnvelope),
	}
}

// SetAuthorizer configures chat authorization checks on the hub.
func (h *Hub) SetAuthorizer(auth ChatAuthorizer) {
	h.Authorizer = auth
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("New client registered. ID: %s, ReceiverID: %s", client.ID, client.ReceiverID)
			h.Clients[client.ID] = client

		case client := <-h.Unregister:
			log.Printf("Client unregistered. ID: %s", client.ID)
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}

		case envelope := <-h.Broadcast:
			// Server-side enforcement of mutual follow authorization
			if h.Authorizer != nil && envelope.SenderID != "" && envelope.ReceiverID != "" && envelope.SenderID != envelope.ReceiverID {
				senderUUID, err1 := uuid.FromString(envelope.SenderID)
				receiverUUID, err2 := uuid.FromString(envelope.ReceiverID)
				if err1 == nil && err2 == nil {
					canChat, err := h.Authorizer.CanChat(context.Background(), senderUUID, receiverUUID)
					if err != nil || !canChat {
						log.Printf("WebSocket transmission blocked between %s and %s: mutual follow required", envelope.SenderID, envelope.ReceiverID)
						if sender, ok := h.Clients[envelope.SenderID]; ok {
							errPayload, _ := json.Marshal(map[string]interface{}{
								"type":       "error",
								"error_code": "MUTUAL_FOLLOW_REQUIRED",
								"message":    "Messaging is restricted: both users must follow each other to exchange messages.",
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

			log.Printf("Broadcasting message from %s to %s", envelope.SenderID, envelope.ReceiverID)
			// Find the receiver client
			if receiver, ok := h.Clients[envelope.ReceiverID]; ok {
				log.Printf("Found receiver client: %s", receiver.ID)
				select {
				case receiver.Send <- envelope.Message:
					log.Printf("Message sent to receiver: %s", receiver.ID)
				default:
					log.Printf("Failed to send message to receiver: %s", receiver.ID)
					close(receiver.Send)
					delete(h.Clients, receiver.ID)
				}
			} else {
				log.Printf("Receiver not found: %s", envelope.ReceiverID)
			}
		}
	}
}
