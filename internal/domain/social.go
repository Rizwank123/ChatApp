package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// UserCounters holds pre-aggregated counters for fast profile and feed queries.
type UserCounters struct {
	PersonnelID    uuid.UUID `json:"personnel_id"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
	PostCount      int64     `json:"post_count"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Block represents a user blocking another user.
type Block struct {
	Base
	BlockerID uuid.UUID `json:"blocker_id"`
	BlockedID uuid.UUID `json:"blocked_id"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FollowRequestStatus represents the state of a private account follow request.
type FollowRequestStatus string

const (
	FollowRequestPending  FollowRequestStatus = "PENDING"
	FollowRequestAccepted FollowRequestStatus = "ACCEPTED"
	FollowRequestRejected FollowRequestStatus = "REJECTED"
)

// FollowRequest represents a request to follow a private account.
type FollowRequest struct {
	Base
	RequesterID uuid.UUID           `json:"requester_id"`
	RequestedID uuid.UUID           `json:"requested_id"`
	Status      FollowRequestStatus `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// SavedPost represents a bookmarked post.
type SavedPost struct {
	ID          uuid.UUID `json:"id"`
	PersonnelID uuid.UUID `json:"personnel_id"`
	PostID      uuid.UUID `json:"post_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationType defines categories of notifications.
type NotificationType string

const (
	NotificationFollow         NotificationType = "FOLLOW"
	NotificationFollowRequest  NotificationType = "FOLLOW_REQUEST"
	NotificationFollowAccepted NotificationType = "FOLLOW_ACCEPTED"
	NotificationLike           NotificationType = "LIKE"
	NotificationComment        NotificationType = "COMMENT"
	NotificationMention        NotificationType = "MENTION"
	NotificationSystem         NotificationType = "SYSTEM"
)

// Notification represents an in-app or real-time notification.
type Notification struct {
	Base
	RecipientID      uuid.UUID        `json:"recipient_id"`
	ActorID          *uuid.UUID       `json:"actor_id,omitempty"`
	NotificationType NotificationType `json:"notification_type"`
	EntityType       string           `json:"entity_type,omitempty"`
	EntityID         *uuid.UUID       `json:"entity_id,omitempty"`
	Title            string           `json:"title"`
	Message          string           `json:"message"`
	IsRead           bool             `json:"is_read"`
	ReadAt           *time.Time       `json:"read_at,omitempty"`
	BaseAudit
}

// FeedCacheEntry represents pre-materialized feed entries for fanout-on-write fallback.
type FeedCacheEntry struct {
	ID          uuid.UUID `json:"id"`
	PersonnelID uuid.UUID `json:"personnel_id"`
	PostID      uuid.UUID `json:"post_id"`
	AuthorID    uuid.UUID `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// OutboxEvent represents a domain event stored transactionally in PostgreSQL before publishing.
type OutboxEvent struct {
	ID            uuid.UUID  `json:"id"`
	AggregateType string     `json:"aggregate_type"`
	AggregateID   uuid.UUID  `json:"aggregate_id"`
	EventType     string     `json:"event_type"`
	Payload       []byte     `json:"payload"`
	Status        string     `json:"status"` // PENDING, PUBLISHED, FAILED
	RetryCount    int        `json:"retry_count"`
	LastError     string     `json:"last_error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}

// NotificationRepository defines persistence operations for notifications.
type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FindUserNotifications(ctx context.Context, recipientID uuid.UUID, limit int, offset int) ([]Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, recipientID uuid.UUID) error
	GetUnreadCount(ctx context.Context, recipientID uuid.UUID) (int64, error)
}

// UserRelationship details the social relationship between two users.
type UserRelationship struct {
	TargetPersonnelID uuid.UUID `json:"target_personnel_id"`
	IsFollowing       bool      `json:"is_following"`
	FollowsYou        bool      `json:"follows_you"`
	IsMutual          bool      `json:"is_mutual"`
	CanChat           bool      `json:"can_chat"`
}

// FollowUserItem represents an entry in a followers/following list.
type FollowUserItem struct {
	PersonnelID uuid.UUID `json:"personnel_id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	DisplayName string    `json:"display_name"`
	Avatar      *string   `json:"avatar,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	IsFollowing bool      `json:"is_following"`
	FollowsYou  bool      `json:"follows_you"`
	IsMutual    bool      `json:"is_mutual"`
	CreatedAt   time.Time `json:"created_at"`
}

// SocialGraphRepository defines advanced social graph operations (counters, blocks, follow requests, saved posts).
type SocialGraphRepository interface {
	GetUserCounters(ctx context.Context, personnelID uuid.UUID) (*UserCounters, error)
	IncrementFollowerCount(ctx context.Context, personnelID uuid.UUID, delta int64) error
	IncrementFollowingCount(ctx context.Context, personnelID uuid.UUID, delta int64) error
	IncrementPostCount(ctx context.Context, personnelID uuid.UUID, delta int64) error

	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	IsMutualFollow(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	GetRelationship(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID) (*UserRelationship, error)
	CanChat(ctx context.Context, userA, userB uuid.UUID) (bool, error)

	GetFollowersList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]FollowUserItem, int64, error)
	GetFollowingList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]FollowUserItem, int64, error)
	SearchUsers(ctx context.Context, currentPersonnelID uuid.UUID, query string, limit, offset int) ([]UserProfileSearchResult, error)

	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error)

	CreateFollowRequest(ctx context.Context, req *FollowRequest) error
	UpdateFollowRequestStatus(ctx context.Context, reqID uuid.UUID, status FollowRequestStatus) error
	GetPendingFollowRequests(ctx context.Context, requestedID uuid.UUID) ([]FollowRequest, error)

	SavePost(ctx context.Context, personnelID, postID uuid.UUID) error
	UnsavePost(ctx context.Context, personnelID, postID uuid.UUID) error
	GetSavedPosts(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]Post, int64, error)

	GetFollowerIDs(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]uuid.UUID, error)
	GetFollowingCelebrityIDs(ctx context.Context, personnelID uuid.UUID, celebrityThreshold int64) ([]uuid.UUID, error)
}

// OutboxRepository manages the transactional outbox table.
type OutboxRepository interface {
	CreateEvent(ctx context.Context, event *OutboxEvent) error
	FetchPendingEvents(ctx context.Context, batchSize int) ([]OutboxEvent, error)
	MarkEventPublished(ctx context.Context, id uuid.UUID) error
	MarkEventFailed(ctx context.Context, id uuid.UUID, errMessage string) error
}
