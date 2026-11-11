package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
)

// MockSocialGraphRepository implements domain.SocialGraphRepository for testing.
type MockSocialGraphRepository struct {
	counters             map[uuid.UUID]*domain.UserCounters
	blockedMap           map[string]bool
	followerCounts       map[uuid.UUID]int64
	followingCounts      map[uuid.UUID]int64
	postCounts           map[uuid.UUID]int64
	savedPosts           map[string]bool
	followerIDs          map[uuid.UUID][]uuid.UUID
	celebrityIDs         map[uuid.UUID][]uuid.UUID
	followRequests       map[uuid.UUID]domain.FollowRequest
	incrementFollowerErr error
}

func NewMockSocialGraphRepository() *MockSocialGraphRepository {
	return &MockSocialGraphRepository{
		counters:        make(map[uuid.UUID]*domain.UserCounters),
		blockedMap:      make(map[string]bool),
		followerCounts:  make(map[uuid.UUID]int64),
		followingCounts: make(map[uuid.UUID]int64),
		postCounts:      make(map[uuid.UUID]int64),
		savedPosts:      make(map[string]bool),
		followerIDs:     make(map[uuid.UUID][]uuid.UUID),
		celebrityIDs:    make(map[uuid.UUID][]uuid.UUID),
		followRequests:  make(map[uuid.UUID]domain.FollowRequest),
	}
}

func (m *MockSocialGraphRepository) GetUserCounters(ctx context.Context, personnelID uuid.UUID) (*domain.UserCounters, error) {
	if c, ok := m.counters[personnelID]; ok {
		return c, nil
	}
	return &domain.UserCounters{PersonnelID: personnelID, FollowerCount: 0, FollowingCount: 0, PostCount: 0}, nil
}

func (m *MockSocialGraphRepository) IncrementFollowerCount(ctx context.Context, personnelID uuid.UUID, delta int64) error {
	if m.incrementFollowerErr != nil {
		return m.incrementFollowerErr
	}
	m.followerCounts[personnelID] += delta
	return nil
}

func (m *MockSocialGraphRepository) IncrementFollowingCount(ctx context.Context, personnelID uuid.UUID, delta int64) error {
	m.followingCounts[personnelID] += delta
	return nil
}

func (m *MockSocialGraphRepository) IncrementPostCount(ctx context.Context, personnelID uuid.UUID, delta int64) error {
	m.postCounts[personnelID] += delta
	return nil
}

func (m *MockSocialGraphRepository) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error {
	key := blockerID.String() + ":" + blockedID.String()
	m.blockedMap[key] = true
	return nil
}

func (m *MockSocialGraphRepository) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	key := blockerID.String() + ":" + blockedID.String()
	delete(m.blockedMap, key)
	return nil
}

func (m *MockSocialGraphRepository) IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	key1 := userA.String() + ":" + userB.String()
	key2 := userB.String() + ":" + userA.String()
	return m.blockedMap[key1] || m.blockedMap[key2], nil
}

func (m *MockSocialGraphRepository) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error) {
	var res []uuid.UUID
	for k := range m.blockedMap {
		if len(k) > 36 && k[:36] == blockerID.String() {
			if id, err := uuid.FromString(k[37:]); err == nil {
				res = append(res, id)
			}
		}
	}
	return res, nil
}

func (m *MockSocialGraphRepository) CreateFollowRequest(ctx context.Context, req *domain.FollowRequest) error {
	req.ID = uuid.Must(uuid.NewV4())
	m.followRequests[req.ID] = *req
	return nil
}

func (m *MockSocialGraphRepository) UpdateFollowRequestStatus(ctx context.Context, reqID uuid.UUID, status domain.FollowRequestStatus) error {
	if req, ok := m.followRequests[reqID]; ok {
		req.Status = status
		m.followRequests[reqID] = req
		return nil
	}
	return errors.New("not found")
}

func (m *MockSocialGraphRepository) GetPendingFollowRequests(ctx context.Context, requestedID uuid.UUID) ([]domain.FollowRequest, error) {
	var res []domain.FollowRequest
	for _, req := range m.followRequests {
		if req.RequestedID == requestedID && req.Status == domain.FollowRequestPending {
			res = append(res, req)
		}
	}
	return res, nil
}

func (m *MockSocialGraphRepository) SavePost(ctx context.Context, personnelID, postID uuid.UUID) error {
	m.savedPosts[personnelID.String()+":"+postID.String()] = true
	return nil
}

func (m *MockSocialGraphRepository) UnsavePost(ctx context.Context, personnelID, postID uuid.UUID) error {
	delete(m.savedPosts, personnelID.String()+":"+postID.String())
	return nil
}

func (m *MockSocialGraphRepository) GetSavedPosts(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]domain.Post, int64, error) {
	return []domain.Post{}, 0, nil
}

func (m *MockSocialGraphRepository) GetFollowerIDs(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	return m.followerIDs[personnelID], nil
}

func (m *MockSocialGraphRepository) GetFollowingCelebrityIDs(ctx context.Context, personnelID uuid.UUID, celebrityThreshold int64) ([]uuid.UUID, error) {
	return m.celebrityIDs[personnelID], nil
}

func (m *MockSocialGraphRepository) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	m.followerIDs[followingID] = append(m.followerIDs[followingID], followerID)
	return nil
}

func (m *MockSocialGraphRepository) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	var remaining []uuid.UUID
	for _, f := range m.followerIDs[followingID] {
		if f != followerID {
			remaining = append(remaining, f)
		}
	}
	m.followerIDs[followingID] = remaining
	return nil
}

func (m *MockSocialGraphRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	for _, f := range m.followerIDs[followingID] {
		if f == followerID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockSocialGraphRepository) IsMutualFollow(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	f1, _ := m.IsFollowing(ctx, userA, userB)
	f2, _ := m.IsFollowing(ctx, userB, userA)
	return f1 && f2, nil
}

func (m *MockSocialGraphRepository) GetRelationship(ctx context.Context, current, target uuid.UUID) (*domain.UserRelationship, error) {
	isFollowing, _ := m.IsFollowing(ctx, current, target)
	followsYou, _ := m.IsFollowing(ctx, target, current)
	isBlocked, _ := m.IsBlocked(ctx, current, target)
	return &domain.UserRelationship{
		TargetPersonnelID: target,
		IsFollowing:       isFollowing,
		FollowsYou:        followsYou,
		IsMutual:          isFollowing && followsYou,
		CanChat:           (isFollowing || followsYou) && !isBlocked,
	}, nil
}

func (m *MockSocialGraphRepository) CanChat(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	rel, err := m.GetRelationship(ctx, userA, userB)
	if err != nil {
		return false, err
	}
	return rel.CanChat, nil
}

func (m *MockSocialGraphRepository) GetFollowersList(ctx context.Context, current, target uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error) {
	return []domain.FollowUserItem{}, 0, nil
}

func (m *MockSocialGraphRepository) GetFollowingList(ctx context.Context, current, target uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error) {
	return []domain.FollowUserItem{}, 0, nil
}

func (m *MockSocialGraphRepository) SearchUsers(ctx context.Context, current uuid.UUID, query string, limit, offset int) ([]domain.UserProfileSearchResult, error) {
	return []domain.UserProfileSearchResult{}, nil
}

// MockPostRepository implements minimal domain.PostRepository for tests.
type MockPostRepository struct {
	follows map[string]bool
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{follows: make(map[string]bool)}
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *domain.Post) error { return nil }
func (m *MockPostRepository) UpdatePost(ctx context.Context, post *domain.Post) error { return nil }
func (m *MockPostRepository) DeletePost(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockPostRepository) FindPostByID(ctx context.Context, id uuid.UUID, viewerID ...uuid.UUID) (domain.Post, error) {
	return domain.Post{Base: domain.Base{ID: id}}, nil
}
func (m *MockPostRepository) FindUserPosts(ctx context.Context, personnelID uuid.UUID, options domain.QueryOptions, viewerID ...uuid.UUID) ([]domain.Post, int64, error) {
	return nil, 0, nil
}
func (m *MockPostRepository) FindFeedPosts(ctx context.Context, personnelID uuid.UUID, options domain.QueryOptions) ([]domain.Post, int64, error) {
	return nil, 0, nil
}
func (m *MockPostRepository) FindPublicPosts(ctx context.Context, options domain.QueryOptions, viewerID ...uuid.UUID) ([]domain.Post, int64, error) {
	return nil, 0, nil
}
func (m *MockPostRepository) IncrementPostView(ctx context.Context, postID uuid.UUID) error  { return nil }
func (m *MockPostRepository) IncrementPostShare(ctx context.Context, postID uuid.UUID) error { return nil }
func (m *MockPostRepository) CreateUserMention(ctx context.Context, postID, mentionedPersonnelID uuid.UUID) error {
	return nil
}
func (m *MockPostRepository) GetUserMentions(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *MockPostRepository) CreateFollow(ctx context.Context, follow *domain.Follow) error {
	m.follows[follow.FollowerID.String()+":"+follow.FollowingID.String()] = true
	return nil
}
func (m *MockPostRepository) DeleteFollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	delete(m.follows, followerID.String()+":"+followingID.String())
	return nil
}
func (m *MockPostRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return m.follows[followerID.String()+":"+followingID.String()], nil
}
func (m *MockPostRepository) CheckMutualFollow(ctx context.Context, p1, p2 uuid.UUID) (bool, error) {
	return m.follows[p1.String()+":"+p2.String()] && m.follows[p2.String()+":"+p1.String()], nil
}
func (m *MockPostRepository) CreateLike(ctx context.Context, like *domain.Like) error     { return nil }
func (m *MockPostRepository) DeleteLike(ctx context.Context, pID, postID uuid.UUID) error { return nil }
func (m *MockPostRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	return nil
}
func (m *MockPostRepository) DeleteComment(ctx context.Context, id, pID uuid.UUID) error { return nil }
func (m *MockPostRepository) FindPostComments(ctx context.Context, postID uuid.UUID, options domain.QueryOptions) ([]domain.Comment, int64, error) {
	return nil, 0, nil
}
func (m *MockPostRepository) CreateHashtag(ctx context.Context, h *domain.Hashtag) error { return nil }
func (m *MockPostRepository) FindHashtagByName(ctx context.Context, name string) (domain.Hashtag, error) {
	return domain.Hashtag{}, nil
}
func (m *MockPostRepository) CreatePostHashtag(ctx context.Context, ph *domain.PostHashtag) error {
	return nil
}
func (m *MockPostRepository) FindOrCreateDirectConversation(ctx context.Context, p1, p2 uuid.UUID) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (m *MockPostRepository) SaveDirectMessage(ctx context.Context, msg *domain.Message) error {
	return nil
}
func (m *MockPostRepository) FindDirectMessages(ctx context.Context, p1, p2 uuid.UUID, options domain.QueryOptions) ([]domain.Message, int64, error) {
	return nil, 0, nil
}

// MockNotificationRepository implements domain.NotificationRepository for tests.
type MockNotificationRepository struct {
	notifications []domain.Notification
}

func (m *MockNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	n.ID = uuid.Must(uuid.NewV4())
	n.CreatedAt = time.Now()
	m.notifications = append(m.notifications, *n)
	return nil
}
func (m *MockNotificationRepository) FindUserNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	return m.notifications, int64(len(m.notifications)), nil
}
func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, id, recipientID uuid.UUID) error {
	for i := range m.notifications {
		if m.notifications[i].ID == id && m.notifications[i].RecipientID == recipientID {
			m.notifications[i].IsRead = true
		}
	}
	return nil
}
func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, recipientID uuid.UUID) error {
	for i := range m.notifications {
		if m.notifications[i].RecipientID == recipientID {
			m.notifications[i].IsRead = true
		}
	}
	return nil
}
func (m *MockNotificationRepository) GetUnreadCount(ctx context.Context, recipientID uuid.UUID) (int64, error) {
	var c int64
	for _, n := range m.notifications {
		if n.RecipientID == recipientID && !n.IsRead {
			c++
		}
	}
	return c, nil
}

// -------------------------------------------------------------
// Test Cases
// -------------------------------------------------------------

func TestSocialService_FollowUser_Success(t *testing.T) {
	socialRepo := NewMockSocialGraphRepository()
	postRepo := NewMockPostRepository()
	notifRepo := &MockNotificationRepository{}

	svc := NewSocialService(socialRepo, postRepo, notifRepo, nil, nil)

	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())

	status, err := svc.FollowUser(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != "FOLLOWING" {
		t.Errorf("expected status FOLLOWING, got %s", status)
	}

	// Verify counter increments
	if socialRepo.followingCounts[userA] != 1 {
		t.Errorf("expected userA following count 1, got %d", socialRepo.followingCounts[userA])
	}
	if socialRepo.followerCounts[userB] != 1 {
		t.Errorf("expected userB follower count 1, got %d", socialRepo.followerCounts[userB])
	}

	// Verify notification generated
	if len(notifRepo.notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(notifRepo.notifications))
	}
}

func TestSocialService_FollowUser_SelfFollowFails(t *testing.T) {
	socialRepo := NewMockSocialGraphRepository()
	postRepo := NewMockPostRepository()
	notifRepo := &MockNotificationRepository{}

	svc := NewSocialService(socialRepo, postRepo, notifRepo, nil, nil)

	userA := uuid.Must(uuid.NewV4())
	_, err := svc.FollowUser(context.Background(), userA, userA)
	if err == nil {
		t.Fatal("expected error when user follows themselves, got nil")
	}
}

func TestSocialService_FollowUser_BlockedFails(t *testing.T) {
	socialRepo := NewMockSocialGraphRepository()
	postRepo := NewMockPostRepository()
	notifRepo := &MockNotificationRepository{}

	svc := NewSocialService(socialRepo, postRepo, notifRepo, nil, nil)

	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())

	// Block userA by userB
	_ = socialRepo.BlockUser(context.Background(), userB, userA, "spam")

	_, err := svc.FollowUser(context.Background(), userA, userB)
	if err == nil {
		t.Fatal("expected error when following a user who blocked you, got nil")
	}
}

func TestSocialService_UnfollowUser_Success(t *testing.T) {
	socialRepo := NewMockSocialGraphRepository()
	postRepo := NewMockPostRepository()
	notifRepo := &MockNotificationRepository{}

	svc := NewSocialService(socialRepo, postRepo, notifRepo, nil, nil)

	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())

	// Follow first
	_, _ = svc.FollowUser(context.Background(), userA, userB)

	// Unfollow
	err := svc.UnfollowUser(context.Background(), userA, userB)
	if err != nil {
		t.Fatalf("unexpected error on unfollow: %v", err)
	}

	if socialRepo.followingCounts[userA] != 0 {
		t.Errorf("expected userA following count 0, got %d", socialRepo.followingCounts[userA])
	}
	if socialRepo.followerCounts[userB] != 0 {
		t.Errorf("expected userB follower count 0, got %d", socialRepo.followerCounts[userB])
	}
}

func TestSocialService_SaveAndUnsavePost(t *testing.T) {
	socialRepo := NewMockSocialGraphRepository()
	postRepo := NewMockPostRepository()
	notifRepo := &MockNotificationRepository{}

	svc := NewSocialService(socialRepo, postRepo, notifRepo, nil, nil)

	userID := uuid.Must(uuid.NewV4())
	postID := uuid.Must(uuid.NewV4())

	if err := svc.SavePost(context.Background(), userID, postID); err != nil {
		t.Fatalf("failed to save post: %v", err)
	}

	key := userID.String() + ":" + postID.String()
	if !socialRepo.savedPosts[key] {
		t.Errorf("expected post to be saved in mock repo")
	}

	if err := svc.UnsavePost(context.Background(), userID, postID); err != nil {
		t.Fatalf("failed to unsave post: %v", err)
	}

	if socialRepo.savedPosts[key] {
		t.Errorf("expected post to be unsaved")
	}
}
