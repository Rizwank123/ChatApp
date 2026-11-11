package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/websockets"
)

type mockChatRepo struct {
	conversations  map[uuid.UUID]domain.Conversation
	participants   map[uuid.UUID][]uuid.UUID
	messages       map[uuid.UUID][]domain.Message
	messageStatus  map[uuid.UUID]domain.MessageStatusEntity
}

func newMockChatRepo() *mockChatRepo {
	return &mockChatRepo{
		conversations: make(map[uuid.UUID]domain.Conversation),
		participants:  make(map[uuid.UUID][]uuid.UUID),
		messages:      make(map[uuid.UUID][]domain.Message),
		messageStatus: make(map[uuid.UUID]domain.MessageStatusEntity),
	}
}

func (m *mockChatRepo) CreateFriendship(ctx context.Context, f *domain.Friendship) error { return nil }
func (m *mockChatRepo) UpdateFriendshipStatus(ctx context.Context, id uuid.UUID, s domain.FriendshipStatus) error {
	return nil
}
func (m *mockChatRepo) FindFriendshipByID(ctx context.Context, id uuid.UUID) (domain.Friendship, error) {
	return domain.Friendship{}, nil
}
func (m *mockChatRepo) FindFriendships(ctx context.Context, personnelID uuid.UUID) ([]domain.Friendship, error) {
	return nil, nil
}
func (m *mockChatRepo) CreateConversation(ctx context.Context, c *domain.Conversation) error {
	c.ID = uuid.Must(uuid.NewV4())
	m.conversations[c.ID] = *c
	return nil
}
func (m *mockChatRepo) FindConversationByID(ctx context.Context, id uuid.UUID) (domain.Conversation, error) {
	c, ok := m.conversations[id]
	if !ok {
		return domain.Conversation{}, errors.New("not found")
	}
	return c, nil
}
func (m *mockChatRepo) FindDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*domain.Conversation, error) {
	for cid, participants := range m.participants {
		conv := m.conversations[cid]
		if conv.ConversationType == domain.ConversationTypeDirect {
			hasA, hasB := false, false
			for _, p := range participants {
				if p == userA {
					hasA = true
				}
				if p == userB {
					hasB = true
				}
			}
			if hasA && hasB {
				return &conv, nil
			}
		}
	}
	return nil, nil
}
func (m *mockChatRepo) FindUserConversations(ctx context.Context, personnelID uuid.UUID) ([]domain.Conversation, error) {
	return nil, nil
}
func (m *mockChatRepo) FindConversationParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	return m.participants[conversationID], nil
}
func (m *mockChatRepo) AddParticipant(ctx context.Context, p *domain.ConversationParticipant) error {
	m.participants[p.ConversationID] = append(m.participants[p.ConversationID], p.PersonnelID)
	return nil
}
func (m *mockChatRepo) RemoveParticipant(ctx context.Context, cid, pid uuid.UUID) error { return nil }
func (m *mockChatRepo) CreateMessage(ctx context.Context, msg *domain.Message) error {
	m.messages[msg.ConversationID] = append(m.messages[msg.ConversationID], *msg)
	return nil
}
func (m *mockChatRepo) FindMessagesByConversation(ctx context.Context, cid uuid.UUID, opt domain.QueryOptions) ([]domain.Message, int64, error) {
	return m.messages[cid], int64(len(m.messages[cid])), nil
}
func (m *mockChatRepo) UpdateMessageStatus(ctx context.Context, s *domain.MessageStatusEntity) error {
	m.messageStatus[s.MessageID] = *s
	return nil
}
func (m *mockChatRepo) FindMessageStatuses(ctx context.Context, mid uuid.UUID) ([]domain.MessageStatusEntity, error) {
	return nil, nil
}

type mockBroker struct{}

func (b *mockBroker) PublishMessage(ctx context.Context, event domain.MessageEvent) error { return nil }
func (b *mockBroker) SubscribeMessages(ctx context.Context) (<-chan domain.MessageEvent, error) {
	return nil, nil
}

func TestChatService_FollowAuthorization(t *testing.T) {
	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())
	userC := uuid.Must(uuid.NewV4())
	userD := uuid.Must(uuid.NewV4())

	socialRepo := NewMockSocialGraphRepository()
	chatRepo := newMockChatRepo()
	hub := websockets.NewHub()
	chatService := NewChatService(chatRepo, nil, nil, socialRepo, &mockBroker{}, hub)

	// Case 1: userA and userB have no follow relationship -> cannot create conversation
	_, err := chatService.CreateConversation(userA, domain.CreateConversationInput{
		Name:             "Direct Chat",
		ConversationType: domain.ConversationTypeDirect,
		ParticipantIDs:   []uuid.UUID{userA, userB},
	})
	if err == nil {
		t.Fatal("expected error creating conversation without follow relationship, got nil")
	}

	// Case 2: One-way follow (userA follows userB, userB does not follow userA) -> CAN initiate conversation
	_ = socialRepo.Follow(context.Background(), userA, userB)
	conv, err := chatService.CreateConversation(userA, domain.CreateConversationInput{
		Name:             "Direct Chat",
		ConversationType: domain.ConversationTypeDirect,
		ParticipantIDs:   []uuid.UUID{userA, userB},
	})
	if err != nil {
		t.Fatalf("expected 1-way follow to succeed, got error: %v", err)
	}

	// Case 3: Conversation creation idempotency (initiating again returns existing conversation)
	conv2, err := chatService.CreateConversation(userA, domain.CreateConversationInput{
		Name:             "Direct Chat",
		ConversationType: domain.ConversationTypeDirect,
		ParticipantIDs:   []uuid.UUID{userA, userB},
	})
	if err != nil {
		t.Fatalf("expected idempotent conversation creation, got err: %v", err)
	}
	if conv2.ID != conv.ID {
		t.Fatalf("expected same conversation ID %v, got %v", conv.ID, conv2.ID)
	}

	// Case 4: Sending message while 1-way follow is active
	msg, err := chatService.SendMessage(userA, domain.CreateMessageInput{
		ConversationID: conv.ID,
		MessageType:    domain.MessageTypeText,
		Content:        "Hello there!",
	})
	if err != nil {
		t.Fatalf("expected message to succeed with 1-way follow, got err: %v", err)
	}
	if msg.Content != "Hello there!" {
		t.Errorf("expected content 'Hello there!', got '%s'", msg.Content)
	}

	// Also recipient (userB) can reply because userA follows userB
	replyMsg, err := chatService.SendMessage(userB, domain.CreateMessageInput{
		ConversationID: conv.ID,
		MessageType:    domain.MessageTypeText,
		Content:        "Hey userA!",
	})
	if err != nil {
		t.Fatalf("expected recipient to be able to reply, got err: %v", err)
	}
	if replyMsg.Content != "Hey userA!" {
		t.Errorf("expected content 'Hey userA!', got '%s'", replyMsg.Content)
	}

	// Case 5: Conversation membership validation
	// Non-participant userD cannot read messages, get conversation, or send messages
	_, _, err = chatService.GetMessages(userD, conv.ID, domain.QueryOptions{})
	if err == nil {
		t.Fatal("expected non-participant to be forbidden from reading messages, got nil")
	}
	_, err = chatService.GetConversation(userD, conv.ID)
	if err == nil {
		t.Fatal("expected non-participant to be forbidden from getting conversation, got nil")
	}
	_, err = chatService.SendMessage(userD, domain.CreateMessageInput{
		ConversationID: conv.ID,
		MessageType:    domain.MessageTypeText,
		Content:        "Intruder message",
	})
	if err == nil {
		t.Fatal("expected non-participant to be forbidden from sending messages, got nil")
	}

	// Case 6: Unfollow removes follow relationship completely -> new messages blocked
	_ = socialRepo.Unfollow(context.Background(), userA, userB)

	_, err = chatService.SendMessage(userA, domain.CreateMessageInput{
		ConversationID: conv.ID,
		MessageType:    domain.MessageTypeText,
		Content:        "Are you still there?",
	})
	if err == nil {
		t.Fatal("expected message to be rejected after follow relationship is removed, got nil")
	}

	// But existing conversation and messages are still preserved and retrievable by participants
	retrievedConv, err := chatService.GetConversation(userA, conv.ID)
	if err != nil {
		t.Fatalf("expected existing conversation to remain accessible by participant, got err: %v", err)
	}
	if retrievedConv.ID != conv.ID {
		t.Fatalf("expected conversation ID %v, got %v", conv.ID, retrievedConv.ID)
	}
	messages, total, err := chatService.GetMessages(userA, conv.ID, domain.QueryOptions{})
	if err != nil || total < 2 || len(messages) < 2 {
		t.Fatalf("expected message history to be preserved, total: %d, err: %v", total, err)
	}

	// Case 7: Blocked user prevents chat even if follow relationship is present
	_ = socialRepo.Follow(context.Background(), userA, userC)
	_ = socialRepo.BlockUser(context.Background(), userA, userC, "spam")

	_, err = chatService.CreateConversation(userA, domain.CreateConversationInput{
		Name:             "Chat with blocked user",
		ConversationType: domain.ConversationTypeDirect,
		ParticipantIDs:   []uuid.UUID{userA, userC},
	})
	if err == nil {
		t.Fatal("expected error creating conversation with blocked user, got nil")
	}
}

func TestWebSockets_HubAuthorizer(t *testing.T) {
	socialRepo := NewMockSocialGraphRepository()
	hub := websockets.NewHub()
	hub.SetAuthorizer(socialRepo)

	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())

	// No follow established
	canChat, err := hub.Authorizer.CanChat(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canChat {
		t.Fatalf("expected canChat=false when no follow relationship exists")
	}

	// 1-way follow: userA follows userB
	_ = socialRepo.Follow(context.Background(), userA, userB)

	canChat, err = hub.Authorizer.CanChat(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !canChat {
		t.Fatalf("expected canChat=true when 1-way follow is established")
	}

	// Reverse check also allows communication between them
	canChat, err = hub.Authorizer.CanChat(context.Background(), userB, userA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !canChat {
		t.Fatalf("expected canChat=true for reverse direction when 1-way follow is established")
	}
}
