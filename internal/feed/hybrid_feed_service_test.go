package feed

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type mockFeedSocialRepo struct {
	counters     map[uuid.UUID]*domain.UserCounters
	followerIDs  map[uuid.UUID][]uuid.UUID
	celebrityIDs map[uuid.UUID][]uuid.UUID
}

func newMockFeedSocialRepo() *mockFeedSocialRepo {
	return &mockFeedSocialRepo{
		counters:     make(map[uuid.UUID]*domain.UserCounters),
		followerIDs:  make(map[uuid.UUID][]uuid.UUID),
		celebrityIDs: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockFeedSocialRepo) GetUserCounters(ctx context.Context, pID uuid.UUID) (*domain.UserCounters, error) {
	if c, ok := m.counters[pID]; ok {
		return c, nil
	}
	return &domain.UserCounters{PersonnelID: pID, FollowerCount: 0}, nil
}
func (m *mockFeedSocialRepo) IncrementFollowerCount(ctx context.Context, pID uuid.UUID, delta int64) error {
	return nil
}
func (m *mockFeedSocialRepo) IncrementFollowingCount(ctx context.Context, pID uuid.UUID, delta int64) error {
	return nil
}
func (m *mockFeedSocialRepo) IncrementPostCount(ctx context.Context, pID uuid.UUID, delta int64) error {
	return nil
}
func (m *mockFeedSocialRepo) BlockUser(ctx context.Context, b1, b2 uuid.UUID, reason string) error {
	return nil
}
func (m *mockFeedSocialRepo) UnblockUser(ctx context.Context, b1, b2 uuid.UUID) error { return nil }
func (m *mockFeedSocialRepo) IsBlocked(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockFeedSocialRepo) GetBlockedUsers(ctx context.Context, b uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *mockFeedSocialRepo) CreateFollowRequest(ctx context.Context, r *domain.FollowRequest) error {
	return nil
}
func (m *mockFeedSocialRepo) UpdateFollowRequestStatus(ctx context.Context, rID uuid.UUID, s domain.FollowRequestStatus) error {
	return nil
}
func (m *mockFeedSocialRepo) GetPendingFollowRequests(ctx context.Context, reqID uuid.UUID) ([]domain.FollowRequest, error) {
	return nil, nil
}
func (m *mockFeedSocialRepo) SavePost(ctx context.Context, pID, postID uuid.UUID) error   { return nil }
func (m *mockFeedSocialRepo) UnsavePost(ctx context.Context, pID, postID uuid.UUID) error { return nil }
func (m *mockFeedSocialRepo) GetSavedPosts(ctx context.Context, pID uuid.UUID, l, o int) ([]domain.Post, int64, error) {
	return nil, 0, nil
}
func (m *mockFeedSocialRepo) GetFollowerIDs(ctx context.Context, pID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	return m.followerIDs[pID], nil
}
func (m *mockFeedSocialRepo) GetFollowingCelebrityIDs(ctx context.Context, pID uuid.UUID, threshold int64) ([]uuid.UUID, error) {
	return m.celebrityIDs[pID], nil
}
func (m *mockFeedSocialRepo) Follow(ctx context.Context, f1, f2 uuid.UUID) error   { return nil }
func (m *mockFeedSocialRepo) Unfollow(ctx context.Context, f1, f2 uuid.UUID) error { return nil }
func (m *mockFeedSocialRepo) IsFollowing(ctx context.Context, f1, f2 uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockFeedSocialRepo) IsMutualFollow(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockFeedSocialRepo) GetRelationship(ctx context.Context, c, t uuid.UUID) (*domain.UserRelationship, error) {
	return &domain.UserRelationship{}, nil
}
func (m *mockFeedSocialRepo) CanChat(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockFeedSocialRepo) GetFollowersList(ctx context.Context, c, t uuid.UUID, q string, l, o int) ([]domain.FollowUserItem, int64, error) {
	return nil, 0, nil
}
func (m *mockFeedSocialRepo) GetFollowingList(ctx context.Context, c, t uuid.UUID, q string, l, o int) ([]domain.FollowUserItem, int64, error) {
	return nil, 0, nil
}
func (m *mockFeedSocialRepo) SearchUsers(ctx context.Context, c uuid.UUID, q string, l, o int) ([]domain.UserProfileSearchResult, error) {
	return nil, nil
}

func TestHybridFeedEngine_CelebrityThresholdDecision(t *testing.T) {
	socialRepo := newMockFeedSocialRepo()

	cfg := &HybridFeedConfig{
		CelebrityThreshold: 25000,
		MaxInboxCapacity:   800,
		DefaultPageLimit:   20,
	}

	engine := NewHybridFeedEngine(nil, nil, socialRepo, nil, cfg)

	celebrityID := uuid.Must(uuid.NewV4())
	socialRepo.counters[celebrityID] = &domain.UserCounters{
		PersonnelID:   celebrityID,
		FollowerCount: 50000, // Exceeds 25,000 threshold
	}

	post := &domain.Post{
		Base:        domain.Base{ID: uuid.Must(uuid.NewV4())},
		PersonnelID: celebrityID,
		BaseAudit:   domain.BaseAudit{CreatedAt: time.Now()},
	}

	// Should skip writing to follower inboxes (Fanout-on-Read path)
	err := engine.FanoutPost(context.Background(), post)
	if err != nil {
		t.Fatalf("unexpected error on celebrity fanout: %v", err)
	}

	// Standard user below threshold
	standardUserID := uuid.Must(uuid.NewV4())
	socialRepo.counters[standardUserID] = &domain.UserCounters{
		PersonnelID:   standardUserID,
		FollowerCount: 150, // Below 25,000 threshold
	}
	followerA := uuid.Must(uuid.NewV4())
	socialRepo.followerIDs[standardUserID] = []uuid.UUID{followerA}

	postStandard := &domain.Post{
		Base:        domain.Base{ID: uuid.Must(uuid.NewV4())},
		PersonnelID: standardUserID,
		BaseAudit:   domain.BaseAudit{CreatedAt: time.Now()},
	}

	err = engine.FanoutPost(context.Background(), postStandard)
	if err != nil {
		t.Fatalf("unexpected error on standard fanout: %v", err)
	}
}

type mockPostRepoForFeed struct {
	feedPosts   []domain.Post
	publicPosts []domain.Post
}

func (m *mockPostRepoForFeed) CreatePost(ctx context.Context, post *domain.Post) error { return nil }
func (m *mockPostRepoForFeed) UpdatePost(ctx context.Context, post *domain.Post) error { return nil }
func (m *mockPostRepoForFeed) DeletePost(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *mockPostRepoForFeed) FindPostByID(ctx context.Context, id uuid.UUID, viewerID ...uuid.UUID) (domain.Post, error) {
	return domain.Post{Base: domain.Base{ID: id}}, nil
}
func (m *mockPostRepoForFeed) FindUserPosts(ctx context.Context, personnelID uuid.UUID, options domain.QueryOptions, viewerID ...uuid.UUID) ([]domain.Post, int64, error) {
	return nil, 0, nil
}
func (m *mockPostRepoForFeed) FindFeedPosts(ctx context.Context, personnelID uuid.UUID, options domain.QueryOptions) ([]domain.Post, int64, error) {
	return m.feedPosts, int64(len(m.feedPosts)), nil
}
func (m *mockPostRepoForFeed) FindPublicPosts(ctx context.Context, options domain.QueryOptions, viewerID ...uuid.UUID) ([]domain.Post, int64, error) {
	return m.publicPosts, int64(len(m.publicPosts)), nil
}
func (m *mockPostRepoForFeed) IncrementPostView(ctx context.Context, postID uuid.UUID) error  { return nil }
func (m *mockPostRepoForFeed) IncrementPostShare(ctx context.Context, postID uuid.UUID) error { return nil }
func (m *mockPostRepoForFeed) CreateFollow(ctx context.Context, follow *domain.Follow) error  { return nil }
func (m *mockPostRepoForFeed) DeleteFollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	return nil
}
func (m *mockPostRepoForFeed) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPostRepoForFeed) CheckMutualFollow(ctx context.Context, p1, p2 uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPostRepoForFeed) CreateLike(ctx context.Context, like *domain.Like) error     { return nil }
func (m *mockPostRepoForFeed) DeleteLike(ctx context.Context, pID, postID uuid.UUID) error { return nil }
func (m *mockPostRepoForFeed) CreateComment(ctx context.Context, comment *domain.Comment) error {
	return nil
}
func (m *mockPostRepoForFeed) DeleteComment(ctx context.Context, id, pID uuid.UUID) error { return nil }
func (m *mockPostRepoForFeed) FindPostComments(ctx context.Context, postID uuid.UUID, options domain.QueryOptions) ([]domain.Comment, int64, error) {
	return nil, 0, nil
}
func (m *mockPostRepoForFeed) CreateHashtag(ctx context.Context, hashtag *domain.Hashtag) error {
	return nil
}
func (m *mockPostRepoForFeed) FindHashtagByName(ctx context.Context, name string) (domain.Hashtag, error) {
	return domain.Hashtag{}, nil
}
func (m *mockPostRepoForFeed) CreatePostHashtag(ctx context.Context, postHashtag *domain.PostHashtag) error {
	return nil
}
func (m *mockPostRepoForFeed) CreateUserMention(ctx context.Context, postID, mentionedPersonnelID uuid.UUID) error {
	return nil
}
func (m *mockPostRepoForFeed) GetUserMentions(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *mockPostRepoForFeed) FindOrCreateDirectConversation(ctx context.Context, p1, p2 uuid.UUID) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (m *mockPostRepoForFeed) SaveDirectMessage(ctx context.Context, message *domain.Message) error {
	return nil
}
func (m *mockPostRepoForFeed) FindDirectMessages(ctx context.Context, p1, p2 uuid.UUID, options domain.QueryOptions) ([]domain.Message, int64, error) {
	return nil, 0, nil
}

func TestHybridFeedEngine_FallbackToPublicExplore(t *testing.T) {
	socialRepo := newMockFeedSocialRepo()
	postRepo := &mockPostRepoForFeed{
		feedPosts: nil, // User has no friends/feed posts
		publicPosts: []domain.Post{
			{
				Base:    domain.Base{ID: uuid.Must(uuid.NewV4())},
				Caption: "Explore the world!",
			},
		},
	}

	engine := NewHybridFeedEngine(nil, nil, socialRepo, postRepo, nil)
	viewerID := uuid.Must(uuid.NewV4())

	feedPage, err := engine.GetUserFeed(context.Background(), viewerID, "", 10)
	if err != nil {
		t.Fatalf("unexpected error getting feed: %v", err)
	}

	if len(feedPage.Items) != 1 {
		t.Fatalf("expected 1 fallback public post, got %d", len(feedPage.Items))
	}
	if feedPage.Items[0].Post.Caption != "Explore the world!" {
		t.Errorf("expected explore post caption, got %s", feedPage.Items[0].Post.Caption)
	}
}
