package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
)

type (
	MessageType      string
	MessageStatus    string
	FriendshipStatus string
	ConversationType string
)

const (
	MessageTypeText  MessageType = "TEXT"
	MessageTypeImage MessageType = "IMAGE"
	MessageTypeFile  MessageType = "FILE"
)

const (
	MessageStatusDelivered MessageStatus = "DELIVERED"
	MessageStatusRead      MessageStatus = "READ"
)

const (
	FriendshipStatusPending  FriendshipStatus = "PENDING"
	FriendshipStatusAccepted FriendshipStatus = "ACCEPTED"
	FriendshipStatusBlocked  FriendshipStatus = "BLOCKED"
)

const (
	ConversationTypeDirect ConversationType = "DIRECT"
	ConversationTypeGroup  ConversationType = "GROUP"
)

type (
	Friendship struct {
		Base
		PersonnelID uuid.UUID        `db:"personnel_id" json:"personnel_id"`
		FriendID    uuid.UUID        `db:"friend_id" json:"friend_id"`
		Status      FriendshipStatus `db:"status" json:"status"`
		BaseAudit
	}

	Conversation struct {
		Base
		Name             *string          `db:"name" json:"name,omitempty"`
		ConversationType ConversationType `db:"conversation_type" json:"conversation_type"`
		BaseAudit
	}

	ConversationParticipant struct {
		Base
		ConversationID uuid.UUID  `db:"conversation_id" json:"conversation_id"`
		PersonnelID    uuid.UUID  `db:"personnel_id" json:"personnel_id"`
		JoinedAt       time.Time  `db:"joined_at" json:"joined_at"`
		LeftAt         *time.Time `db:"left_at" json:"left_at,omitempty"`
		BaseAudit
	}

	Message struct {
		Base
		ConversationID uuid.UUID   `db:"conversation_id" json:"conversation_id"`
		SenderID       uuid.UUID   `db:"sender_id" json:"sender_id"`
		MessageType    MessageType `db:"message_type" json:"message_type"`
		Content        string      `db:"content" json:"content"`
		CreatedAt      time.Time   `db:"created_at" json:"created_at"`
		BaseAudit
	}

	MessageStatusEntity struct {
		Base
		MessageID   uuid.UUID     `db:"message_id" json:"message_id"`
		PersonnelID uuid.UUID     `db:"personnel_id" json:"personnel_id"`
		Status      MessageStatus `db:"status" json:"status"`
		BaseAudit
	}
)

type (
	CreateFriendshipInput struct {
		FriendUsername string `json:"friend_username"`
	}

	CreateMessageInput struct {
		ConversationID uuid.UUID   `json:"conversation_id"`
		MessageType    MessageType `json:"message_type"`
		Content        string      `json:"content"`
	}

	CreateConversationInput struct {
		Name             string           `json:"name,omitempty"`
		ConversationType ConversationType `json:"conversation_type"`
		ParticipantIDs   []uuid.UUID      `json:"participant_ids"`
	}

	UpdateMessageStatusInput struct {
		Status MessageStatus `json:"status"`
	}
)

type (
	ChatRepository interface {
		// Friendship methods
		CreateFriendship(ctx context.Context, friendship *Friendship) error
		UpdateFriendshipStatus(ctx context.Context, id uuid.UUID, status FriendshipStatus) error
		FindFriendshipByID(ctx context.Context, id uuid.UUID) (Friendship, error)
		FindFriendships(ctx context.Context, personnelID uuid.UUID) ([]Friendship, error)

		// Conversation methods
		CreateConversation(ctx context.Context, conversation *Conversation) error
		FindConversationByID(ctx context.Context, id uuid.UUID) (Conversation, error)
		FindDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*Conversation, error)
		FindUserConversations(ctx context.Context, personnelID uuid.UUID) ([]Conversation, error)
		FindConversationParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
		AddParticipant(ctx context.Context, participant *ConversationParticipant) error
		RemoveParticipant(ctx context.Context, conversationID, personnelID uuid.UUID) error

		// Message methods
		CreateMessage(ctx context.Context, message *Message) error
		FindMessagesByConversation(ctx context.Context, conversationID uuid.UUID, options QueryOptions) ([]Message, int64, error)
		UpdateMessageStatus(ctx context.Context, status *MessageStatusEntity) error
		FindMessageStatuses(ctx context.Context, messageID uuid.UUID) ([]MessageStatusEntity, error)
	}

	ChatService interface {
		// Friendship methods
		SendFriendRequest(personnelID uuid.UUID, input CreateFriendshipInput) (Friendship, error)
		AcceptFriendRequest(personnelID uuid.UUID, friendshipID uuid.UUID) error
		GetFriendships(personnelID uuid.UUID) ([]Friendship, error)

		// Conversation methods
		CreateConversation(personnelID uuid.UUID, input CreateConversationInput) (Conversation, error)
		GetConversation(personnelID uuid.UUID, id uuid.UUID) (Conversation, error)
		GetUserConversations(personnelID uuid.UUID) ([]Conversation, error)

		// Message methods
		SendMessage(personnelID uuid.UUID, input CreateMessageInput) (Message, error)
		GetMessages(personnelID uuid.UUID, conversationID uuid.UUID, options QueryOptions) ([]Message, int64, error)
		UpdateMessageStatus(personnelID uuid.UUID, messageID uuid.UUID, input UpdateMessageStatusInput) error
	}

	WebSocketService interface {
		RegisterConnection(personnelID uuid.UUID, conn *websocket.Conn)
		UnregisterConnection(personnelID uuid.UUID)
		SendMessageToUser(personnelID uuid.UUID, message Message) error
		BroadcastToConversation(conversationID uuid.UUID, message Message) error
	}

	MessageEvent struct {
		ConversationID uuid.UUID   `json:"conversation_id"`
		Recipients     []uuid.UUID `json:"recipients"`
		Message        Message     `json:"message"`
	}

	MessageBroker interface {
		PublishMessage(ctx context.Context, event MessageEvent) error
		SubscribeMessages(ctx context.Context) (<-chan MessageEvent, error)
	}
)
