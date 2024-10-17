package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	Mstatus string // @name Mstatus define type  Mstatus
)

type (
	// Message defines the model for Message
	Message struct {
		Base
		SenderID   uuid.UUID `db:"sender_id" json:"sender_id"  validate:"required" example:""`
		ReceiverID uuid.UUID `db:"receiver_id" json:"receiver_id"  validate:"required" example:""`
		Content    string    `db:"content" json:"content,omitempty" example:"hi how are you"`
		BaseAudit
	} // @name  Message
	// MessageStatus defines the model for MessageStatus
	MessageStatus struct {
		Base
		MessageID uuid.UUID `db:"message_id" json:"message_id"  validate:"required"`
		Mstatus   Mstatus   `db:"status" json:"m_status"  validate:"required oneof=Sent Delivered, Read"`
		UpdatedAt time.Time `db:"updated_at" json:"updated_at" example:""`
	} // @name  MessageStatus
)

type (
	// CreateMessageInput  defines the input for CreateMessageInput
	CreateMessageInput struct {
		SenderID   uuid.UUID `json:"sender_id" validate:"required"`
		ReceiverID uuid.UUID `json:"receiver_id" validate:"required"`
		Content    string    `json:"content" validate:"required"`
	} // @name  CreateMessageInput

	// UpdateMessageInput  defines the input for UpdateMessageInput
	UpdateMessageInput struct {
		Content string `json:"content" validate:"required"`
	} // @name  UpdateMessageInput

	// UpdateMessageStatusInput defines the model for  UpdateMessageStatusInput
	UpdateMessageStatusInput struct {
		Mstatus Mstatus `json:"m_status" validate:"required oneof=Sent Delivered Read"`
	} // @name  UpdateMessageStatusInput

)

const (
	MstatusSent      Mstatus = "Sent"
	MstatusDelivered Mstatus = "Delivered"
	MstatusRead      Mstatus = "Read"
)

type (
	// MessageRepository defines the methods that any message Repository should implement
	MessageRepository interface {
		// FindByID  returns a message by its ID
		FindByID(ctx context.Context, id uuid.UUID) (result Message, err error)
		// FindAll returns all messages
		FindAll(ctx context.Context) (result []Message, err error)
		// FindBySenderID returns all messages sent by a
		FindBySenderID(ctx context.Context, senderID uuid.UUID) (result []Message, err error)
		// FindByReceiverID returns all messages received by a user
		FindByReceiverID(ctx context.Context, receiverID uuid.UUID) (result []Message, err error)
		// Create creates a new message
		Create(ctx context.Context, entity *Message) (err error)
		// CreateMultiple creates multiple messages
		CreateMultiple(ctx context.Context, entities []*Message) (err error)
		// UpdateMessage updates a message
		Update(ctx context.Context, entity *Message) (err error)
		// UpdateMultiple updates multiple messages
		UpdateMultiple(ctx context.Context, entities []*Message) (err error)
		// CreateMessageStatus creates a new message status
		CreateMessageStatus(ctx context.Context, entity *MessageStatus) (err error)
		// UpdateMessageStatus updates the status of a message
		UpdateMessageStatus(ctx context.Context, entity *MessageStatus) (err error)
		// Delete Message deletes a message
		Delete(ctx context.Context, id uuid.UUID) (err error)
	} // @name  MessageRepository

	//  MessageService defines the methods that any message Service should implement
	MessageService interface {
		// FindByID  returns a message by its ID
		FindByID(id uuid.UUID) (result Message, err error)
		// FindAll returns all messages
		FindAll() (result []Message, err error)
		// FindBySenderID returns all messages sent by a
		FindBySenderID(senderID uuid.UUID) (result []Message, err error)
		// FindByReceiverID returns all messages received by a user
		FindByReceiverID(receiverID uuid.UUID) (result []Message, err error)
		// Create creates a new message
		Create(in CreateMessageInput) (result Message, err error)
		// Update Message updates a message
		Update(in UpdateMessageInput) (result Message, err error)
		// Delete Message deletes a message
		Delete(id uuid.UUID) (err error)
	}
)
