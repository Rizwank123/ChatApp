package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/websockets"
)

type ChatService struct {
	cr         domain.ChatRepository
	pr         domain.PersonnelRepository
	ur         domain.UserRepository
	socialRepo domain.SocialGraphRepository

	broker domain.MessageBroker
	wss    *websockets.Hub
}

func NewChatService(
	cr domain.ChatRepository,
	pr domain.PersonnelRepository,
	ur domain.UserRepository,
	socialRepo domain.SocialGraphRepository,
	broker domain.MessageBroker,
	wss *websockets.Hub,
) domain.ChatService {
	if wss != nil && socialRepo != nil {
		wss.SetAuthorizer(socialRepo)
	}
	return &ChatService{
		cr:         cr,
		pr:         pr,
		ur:         ur,
		socialRepo: socialRepo,
		broker:     broker,
		wss:        wss,
	}
}

// Friendship methods

func (s *ChatService) SendFriendRequest(personnelID uuid.UUID, input domain.CreateFriendshipInput) (domain.Friendship, error) {
	ctx := context.Background()

	user, err := s.ur.FindByUserName(ctx, input.FriendUsername)
	if err != nil {
		return domain.Friendship{}, fmt.Errorf("friend not found: %w", err)
	}

	friend, err := s.pr.FindByUserID(ctx, user.ID)
	if err != nil {
		return domain.Friendship{}, fmt.Errorf("friend not found: %w", err)
	}

	existing, err := s.cr.FindFriendships(ctx, personnelID)
	if err != nil {
		return domain.Friendship{}, fmt.Errorf("error checking existing friendships: %w", err)
	}

	for _, f := range existing {
		if (f.PersonnelID == personnelID && f.FriendID == friend.ID) ||
			(f.PersonnelID == friend.ID && f.FriendID == personnelID) {
			return domain.Friendship{}, errors.New("friendship already exists")
		}
	}

	friendship := &domain.Friendship{
		PersonnelID: personnelID,
		FriendID:    friend.ID,
		Status:      domain.FriendshipStatusPending,
	}

	if err := s.cr.CreateFriendship(ctx, friendship); err != nil {
		return domain.Friendship{}, fmt.Errorf("error creating friendship: %w", err)
	}

	return *friendship, nil
}

func (s *ChatService) AcceptFriendRequest(personnelID, friendshipID uuid.UUID) error {
	ctx := context.Background()

	friendship, err := s.cr.FindFriendshipByID(ctx, friendshipID)
	if err != nil {
		return fmt.Errorf("error finding friendship: %w", err)
	}

	if friendship.FriendID != personnelID {
		return errors.New("unauthorized to accept this friend request")
	}

	if friendship.Status != domain.FriendshipStatusPending {
		return errors.New("friendship is not pending")
	}

	if err := s.cr.UpdateFriendshipStatus(ctx, friendshipID, domain.FriendshipStatusAccepted); err != nil {
		return fmt.Errorf("error updating friendship status: %w", err)
	}

	return nil
}

func (s *ChatService) GetFriendships(personnelID uuid.UUID) ([]domain.Friendship, error) {
	ctx := context.Background()

	friendships, err := s.cr.FindFriendships(ctx, personnelID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving friendships: %w", err)
	}

	return friendships, nil
}

// Conversation methods

func (s *ChatService) CreateConversation(personnelID uuid.UUID, input domain.CreateConversationInput) (domain.Conversation, error) {
	ctx := context.Background()

	if len(input.ParticipantIDs) == 0 {
		return domain.Conversation{}, errors.New("conversation must have at least one participant")
	}

	hasCreator := false
	for _, id := range input.ParticipantIDs {
		if id == personnelID {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		input.ParticipantIDs = append([]uuid.UUID{personnelID}, input.ParticipantIDs...)
	}

	if input.ConversationType == domain.ConversationTypeDirect {
		if len(input.ParticipantIDs) != 2 {
			return domain.Conversation{}, errors.New("direct conversation must have exactly two participants including the creator")
		}

		var otherParticipantID uuid.UUID
		for _, id := range input.ParticipantIDs {
			if id != personnelID {
				otherParticipantID = id
				break
			}
		}

		// Check if direct conversation already exists
		existing, err := s.cr.FindDirectConversation(ctx, personnelID, otherParticipantID)
		if err != nil {
			return domain.Conversation{}, fmt.Errorf("error checking existing direct conversation: %w", err)
		}
		if existing != nil {
			return *existing, nil
		}

		if s.socialRepo != nil {
			canChat, err := s.socialRepo.CanChat(ctx, personnelID, otherParticipantID)
			if err != nil {
				return domain.Conversation{}, fmt.Errorf("error verifying chat permission: %w", err)
			}
			if !canChat {
				return domain.Conversation{}, errors.New("cannot create conversation: messaging requires either user to follow the other and active accounts")
			}
		}
	}

	conversation := domain.Conversation{
		Name:             &input.Name,
		ConversationType: input.ConversationType,
	}

	if err := s.cr.CreateConversation(ctx, &conversation); err != nil {
		if input.ConversationType == domain.ConversationTypeDirect {
			for _, id := range input.ParticipantIDs {
				if id != personnelID {
					if existing, findErr := s.cr.FindDirectConversation(ctx, personnelID, id); findErr == nil && existing != nil {
						return *existing, nil
					}
					break
				}
			}
		}
		return domain.Conversation{}, fmt.Errorf("error creating conversation: %w", err)
	}

	for _, id := range input.ParticipantIDs {
		participant := &domain.ConversationParticipant{
			ConversationID: conversation.ID,
			PersonnelID:    id,
			JoinedAt:       time.Now(),
		}
		if err := s.cr.AddParticipant(ctx, participant); err != nil {
			return domain.Conversation{}, fmt.Errorf("error adding participant: %w", err)
		}
	}

	return conversation, nil
}

func (s *ChatService) GetConversation(personnelID uuid.UUID, id uuid.UUID) (domain.Conversation, error) {
	ctx := context.Background()

	participants, err := s.cr.FindConversationParticipantIDs(ctx, id)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("error finding conversation participants: %w", err)
	}

	isParticipant := false
	for _, p := range participants {
		if p == personnelID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return domain.Conversation{}, domain.ForbiddenAccessError{
			Code:    "FORBIDDEN_ACCESS",
			Message: "you are not a participant in this conversation",
		}
	}

	conversation, err := s.cr.FindConversationByID(ctx, id)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("error finding conversation: %w", err)
	}

	return conversation, nil
}

func (s *ChatService) GetUserConversations(personnelID uuid.UUID) ([]domain.Conversation, error) {
	ctx := context.Background()

	conversations, err := s.cr.FindUserConversations(ctx, personnelID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving conversations: %w", err)
	}

	return conversations, nil
}

// Message methods

func (s *ChatService) SendMessage(personnelID uuid.UUID, input domain.CreateMessageInput) (domain.Message, error) {
	ctx := context.Background()

	if _, err := s.cr.FindConversationByID(ctx, input.ConversationID); err != nil {
		return domain.Message{}, fmt.Errorf("conversation not found: %w", err)
	}

	recipients, err := s.cr.FindConversationParticipantIDs(ctx, input.ConversationID)
	if err != nil {
		return domain.Message{}, fmt.Errorf("error finding message recipients: %w", err)
	}

	isParticipant := false
	for _, p := range recipients {
		if p == personnelID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return domain.Message{}, domain.ForbiddenAccessError{
			Code:    "FORBIDDEN_ACCESS",
			Message: "sender is not a participant in this conversation",
		}
	}

	if s.socialRepo != nil {
		for _, recipientID := range recipients {
			if recipientID == personnelID {
				continue
			}
			canChat, err := s.socialRepo.CanChat(ctx, personnelID, recipientID)
			if err != nil {
				return domain.Message{}, fmt.Errorf("error checking chat permissions: %w", err)
			}
			if !canChat {
				return domain.Message{}, errors.New("cannot send message: messaging requires either user to follow the other and active accounts")
			}
		}
	}

	message := &domain.Message{
		Base:           domain.Base{ID: uuid.Must(uuid.NewV4())},
		ConversationID: input.ConversationID,
		SenderID:       personnelID,
		MessageType:    input.MessageType,
		Content:        input.Content,
		CreatedAt:      time.Now(),
		BaseAudit:      domain.BaseAudit{},
	}

	if err := s.cr.CreateMessage(ctx, message); err != nil {
		return domain.Message{}, fmt.Errorf("error saving message: %w", err)
	}

	status := &domain.MessageStatusEntity{
		Base:        domain.Base{ID: uuid.Must(uuid.NewV4())},
		MessageID:   message.ID,
		PersonnelID: personnelID,
		Status:      domain.MessageStatusRead,
		BaseAudit:   domain.BaseAudit{},
	}

	if err := s.cr.UpdateMessageStatus(ctx, status); err != nil {
		return domain.Message{}, fmt.Errorf("error saving message status: %w", err)
	}

	if err := s.broker.PublishMessage(ctx, domain.MessageEvent{
		ConversationID: input.ConversationID,
		Recipients:     recipients,
		Message:        *message,
	}); err != nil {
		return domain.Message{}, fmt.Errorf("error publishing message event: %w", err)
	}

	return *message, nil
}

func DeliverBrokerMessages(ctx context.Context, broker domain.MessageBroker, hub *websockets.Hub) error {
	events, err := broker.SubscribeMessages(ctx)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return nil
			}
			payload, err := json.Marshal(event.Message)
			if err != nil {
				continue
			}
			for _, recipientID := range event.Recipients {
				hub.Broadcast <- websockets.MessageEnvelope{
					SenderID:   event.Message.SenderID.String(),
					ReceiverID: recipientID.String(),
					Message:    payload,
				}
			}
		}
	}
}

func (s *ChatService) GetMessages(personnelID uuid.UUID, conversationID uuid.UUID, options domain.QueryOptions) ([]domain.Message, int64, error) {
	ctx := context.Background()

	participants, err := s.cr.FindConversationParticipantIDs(ctx, conversationID)
	if err != nil {
		return nil, 0, fmt.Errorf("error checking conversation membership: %w", err)
	}

	isParticipant := false
	for _, p := range participants {
		if p == personnelID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, 0, domain.ForbiddenAccessError{
			Code:    "FORBIDDEN_ACCESS",
			Message: "you are not a participant in this conversation",
		}
	}

	messages, total, err := s.cr.FindMessagesByConversation(ctx, conversationID, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error retrieving messages: %w", err)
	}

	return messages, total, nil
}

func (s *ChatService) UpdateMessageStatus(personnelID, messageID uuid.UUID, input domain.UpdateMessageStatusInput) error {
	ctx := context.Background()

	status := &domain.MessageStatusEntity{
		Base:        domain.Base{ID: uuid.Must(uuid.NewV4())},
		MessageID:   messageID,
		PersonnelID: personnelID,
		Status:      input.Status,
		BaseAudit:   domain.BaseAudit{},
	}

	if err := s.cr.UpdateMessageStatus(ctx, status); err != nil {
		return fmt.Errorf("error updating message status: %w", err)
	}

	return nil
}
