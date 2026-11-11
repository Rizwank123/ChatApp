package service

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/feed"
	"github.com/Rizwank123/chatApp/internal/websockets"
)

// SocialService defines operations for social graph, feed, and notifications.
type SocialService interface {
	FollowUser(ctx context.Context, followerID, followingID uuid.UUID) (status string, err error)
	UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error
	GetRelationship(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID) (*domain.UserRelationship, error)
	CanChat(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	GetFollowersList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error)
	GetFollowingList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error)
	SearchUsers(ctx context.Context, currentPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.UserProfileSearchResult, error)
	RespondFollowRequest(ctx context.Context, reqID uuid.UUID, recipientID uuid.UUID, accept bool) error
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	SavePost(ctx context.Context, personnelID, postID uuid.UUID) error
	UnsavePost(ctx context.Context, personnelID, postID uuid.UUID) error
	GetSavedPosts(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]domain.Post, int64, error)
	GetUserCounters(ctx context.Context, personnelID uuid.UUID) (*domain.UserCounters, error)
	GetFeed(ctx context.Context, userID uuid.UUID, cursor string, limit int64) (*feed.FeedPage, error)
	GetNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error)
	MarkNotificationRead(ctx context.Context, id, recipientID uuid.UUID) error
	MarkAllNotificationsRead(ctx context.Context, recipientID uuid.UUID) error
}

type socialService struct {
	socialRepo domain.SocialGraphRepository
	postRepo   domain.PostRepository
	notifRepo  domain.NotificationRepository
	feedEngine *feed.HybridFeedEngine
	wsHub      *websockets.DistributedHub
}

// NewSocialService constructs SocialService.
func NewSocialService(
	socialRepo domain.SocialGraphRepository,
	postRepo domain.PostRepository,
	notifRepo domain.NotificationRepository,
	feedEngine *feed.HybridFeedEngine,
	wsHub *websockets.DistributedHub,
) SocialService {
	return &socialService{
		socialRepo: socialRepo,
		postRepo:   postRepo,
		notifRepo:  notifRepo,
		feedEngine: feedEngine,
		wsHub:      wsHub,
	}
}

func (s *socialService) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) (string, error) {
	if followerID == followingID {
		return "", fmt.Errorf("users cannot follow themselves")
	}

	// 1. Check if blocked
	blocked, err := s.socialRepo.IsBlocked(ctx, followerID, followingID)
	if err != nil {
		return "", err
	}
	if blocked {
		return "", fmt.Errorf("cannot follow blocked user")
	}

	// 2. Check if already following
	alreadyFollowing, err := s.postRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return "", err
	}
	if alreadyFollowing {
		return "FOLLOWING", nil
	}

	// 3. Create follow relation
	follow := &domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	if err := s.postRepo.CreateFollow(ctx, follow); err != nil {
		return "", err
	}

	// 4. Concurrency-safe counter increments
	_ = s.socialRepo.IncrementFollowingCount(ctx, followerID, 1)
	_ = s.socialRepo.IncrementFollowerCount(ctx, followingID, 1)

	// 5. Send notification
	notif := &domain.Notification{
		RecipientID:      followingID,
		ActorID:          &followerID,
		NotificationType: domain.NotificationFollow,
		Title:            "New Follower",
		Message:          "A user started following you",
	}
	if err := s.notifRepo.Create(ctx, notif); err == nil && s.wsHub != nil {
		_ = s.wsHub.BroadcastToUser(ctx, followingID, "notification", notif)
	}

	return "FOLLOWING", nil
}

func (s *socialService) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	if err := s.socialRepo.Unfollow(ctx, followerID, followingID); err != nil {
		return err
	}

	_ = s.socialRepo.IncrementFollowingCount(ctx, followerID, -1)
	_ = s.socialRepo.IncrementFollowerCount(ctx, followingID, -1)

	return nil
}

func (s *socialService) GetRelationship(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID) (*domain.UserRelationship, error) {
	return s.socialRepo.GetRelationship(ctx, currentPersonnelID, targetPersonnelID)
}

func (s *socialService) CanChat(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	return s.socialRepo.CanChat(ctx, userA, userB)
}

func (s *socialService) GetFollowersList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error) {
	return s.socialRepo.GetFollowersList(ctx, currentPersonnelID, targetPersonnelID, query, limit, offset)
}

func (s *socialService) GetFollowingList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error) {
	return s.socialRepo.GetFollowingList(ctx, currentPersonnelID, targetPersonnelID, query, limit, offset)
}

func (s *socialService) SearchUsers(ctx context.Context, currentPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.UserProfileSearchResult, error) {
	return s.socialRepo.SearchUsers(ctx, currentPersonnelID, query, limit, offset)
}

func (s *socialService) RespondFollowRequest(ctx context.Context, reqID uuid.UUID, recipientID uuid.UUID, accept bool) error {
	status := domain.FollowRequestRejected
	if accept {
		status = domain.FollowRequestAccepted
	}
	return s.socialRepo.UpdateFollowRequestStatus(ctx, reqID, status)
}

func (s *socialService) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error {
	if blockerID == blockedID {
		return fmt.Errorf("cannot block yourself")
	}
	return s.socialRepo.BlockUser(ctx, blockerID, blockedID, reason)
}

func (s *socialService) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return s.socialRepo.UnblockUser(ctx, blockerID, blockedID)
}

func (s *socialService) SavePost(ctx context.Context, personnelID, postID uuid.UUID) error {
	return s.socialRepo.SavePost(ctx, personnelID, postID)
}

func (s *socialService) UnsavePost(ctx context.Context, personnelID, postID uuid.UUID) error {
	return s.socialRepo.UnsavePost(ctx, personnelID, postID)
}

func (s *socialService) GetSavedPosts(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]domain.Post, int64, error) {
	return s.socialRepo.GetSavedPosts(ctx, personnelID, limit, offset)
}

func (s *socialService) GetUserCounters(ctx context.Context, personnelID uuid.UUID) (*domain.UserCounters, error) {
	return s.socialRepo.GetUserCounters(ctx, personnelID)
}

func (s *socialService) GetFeed(ctx context.Context, userID uuid.UUID, cursor string, limit int64) (*feed.FeedPage, error) {
	if s.feedEngine != nil {
		return s.feedEngine.GetUserFeed(ctx, userID, cursor, limit)
	}

	// Fallback to postRepo when feedEngine is nil (e.g. running without Redis/hybrid engine)
	posts, _, err := s.postRepo.FindFeedPosts(ctx, userID, domain.QueryOptions{
		Limit:  limit,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	items := make([]feed.FeedPostItem, len(posts))
	for i, p := range posts {
		items[i] = feed.FeedPostItem{
			Post:  p,
			Score: float64(p.CreatedAt.UnixMilli()),
		}
	}

	return &feed.FeedPage{
		Items:   items,
		HasMore: int64(len(posts)) == limit,
	}, nil
}

func (s *socialService) GetNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	return s.notifRepo.FindUserNotifications(ctx, recipientID, limit, offset)
}

func (s *socialService) MarkNotificationRead(ctx context.Context, id, recipientID uuid.UUID) error {
	return s.notifRepo.MarkAsRead(ctx, id, recipientID)
}

func (s *socialService) MarkAllNotificationsRead(ctx context.Context, recipientID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(ctx, recipientID)
}
