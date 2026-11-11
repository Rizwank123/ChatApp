package domain

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

// Post represents a social media post
type Post struct {
	Base
	PersonnelID     uuid.UUID   `json:"personnel_id"`
	Caption         string      `json:"caption"`
	MediaURL        []string    `json:"media_url"`
	Media           []PostMedia `json:"media,omitempty"`
	Location        string      `json:"location"`
	Visibility      string      `json:"visibility"`
	Status          string      `json:"status"`
	IsArchived      bool        `json:"is_archived"`
	Likes           int64       `json:"likes_count"`
	Comments        int64       `json:"comments_count"`
	ViewCount       int64       `json:"view_count"`
	ShareCount      int64       `json:"share_count"`
	IsLiked         bool        `json:"is_liked,omitempty"`
	IsSaved         bool        `json:"is_saved,omitempty"`
	CreatorUsername string      `json:"creator_username,omitempty"`
	CreatorName     string      `json:"creator_name,omitempty"`
	CreatorAvatar   string      `json:"creator_avatar,omitempty"`
	BaseAudit
}

// Follow represents a follower relationship between users
type Follow struct {
	Base
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
	BaseAudit
}

// Like represents a like on a post
type Like struct {
	Base
	PersonnelID uuid.UUID `json:"personnel_id"`
	PostID      uuid.UUID `json:"post_id"`
	BaseAudit
}

// Comment represents a comment on a post
type Comment struct {
	Base
	PersonnelID uuid.UUID  `json:"personnel_id"`
	PostID      uuid.UUID  `json:"post_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Content     string     `json:"content"`
	BaseAudit
}

// Hashtag represents a hashtag that can be associated with posts
type Hashtag struct {
	Base
	Name string `json:"name"`
	BaseAudit
}

// PostHashtag represents the many-to-many relationship between posts and hashtags
type PostHashtag struct {
	Base
	PostID    uuid.UUID `json:"post_id"`
	HashtagID uuid.UUID `json:"hashtag_id"`
	BaseAudit
}

// CreatePostInput represents the input for creating a new post
type CreatePostInput struct {
	Caption    string      `json:"caption"`
	MediaURL   []string    `json:"media_url,omitempty"`
	MediaIDs   []uuid.UUID `json:"media_ids,omitempty"`
	Location   string      `json:"location,omitempty"`
	Visibility string      `json:"visibility" validate:"omitempty,oneof=PUBLIC PRIVATE"`
	Status     string      `json:"status" validate:"omitempty,oneof=DRAFT PUBLISHED ARCHIVED"`
	Hashtags   []string    `json:"hashtags,omitempty"`
	Mentions   []string    `json:"mentions,omitempty"`
}

// UpdatePostInput represents the input for updating an existing post
type UpdatePostInput struct {
	Caption    *string   `json:"caption,omitempty"`
	MediaURL   *[]string `json:"media_url,omitempty"`
	Location   *string   `json:"location,omitempty"`
	Visibility *string   `json:"visibility,omitempty" validate:"omitempty,oneof=PUBLIC PRIVATE"`
	Status     *string   `json:"status,omitempty" validate:"omitempty,oneof=DRAFT PUBLISHED ARCHIVED"`
	IsArchived *bool     `json:"is_archived,omitempty"`
	Hashtags   *[]string `json:"hashtags,omitempty"`
}

type CreateCommentInput struct {
	Content  string     `json:"content" validate:"required,trim"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

type FeedFanoutEngine interface {
	FanoutPost(ctx context.Context, post *Post) error
}

type MediaProcessorEnqueuer interface {
	Enqueue(mediaID uuid.UUID)
}

// PostRepository defines the interface for post-related database operations
type (
	PostService interface {
		// Create a new post
		CreatePost(input CreatePostInput, PersonnelID uuid.UUID) (*Post, error)
		// Update an existing post
		UpdatePost(id uuid.UUID, input UpdatePostInput, personnelID uuid.UUID) (err error)
		// Delete a post
		DeletePost(id uuid.UUID, personnelID uuid.UUID) (err error)
		// Get a post by ID
		GetPost(id uuid.UUID, viewerID ...uuid.UUID) (result Post, err error)
		// Get posts for a specific user
		GetUserPosts(personnelID uuid.UUID, options QueryOptions, viewerID ...uuid.UUID) (result []Post, total int64, err error)
		// Get feed posts for a user
		GetFeedPosts(personnelID uuid.UUID, options QueryOptions) (result []Post, total int64, err error)
		// Get public explore posts
		GetPublicPosts(options QueryOptions, viewerID ...uuid.UUID) (result []Post, total int64, err error)
		// Record post interaction metrics
		RecordView(postID uuid.UUID) error
		RecordShare(postID uuid.UUID) error

		// Follow operations
		FollowUser(followerID, followingID uuid.UUID) (err error)
		UnfollowUser(followerID, followingID uuid.UUID) (err error)
		CheckFollowing(followerID, followingID uuid.UUID) (bool, error)
		CheckMutualFollow(personnelID1, personnelID2 uuid.UUID) (bool, error)

		// Like operations
		LikePost(personnelID, postID uuid.UUID) (err error)
		UnlikePost(personnelID, postID uuid.UUID) (err error)

		// Comment operations
		CreateComment(postID uuid.UUID, content string, personnelID uuid.UUID, parentID *uuid.UUID) (err error)
		DeleteComment(commentID uuid.UUID, personnelID uuid.UUID) error
		GetPostComments(postID uuid.UUID, options QueryOptions) (result []Comment, total int64, err error)

		// Chat operations for mutual followers
		CreateDirectChat(senderID, receiverID uuid.UUID) (Conversation, error)
		SendDirectMessage(senderID, receiverID uuid.UUID, content string, messageType MessageType) (Message, error)
		GetDirectMessages(personnelID1, personnelID2 uuid.UUID, options QueryOptions) ([]Message, int64, error)
	}
	PostRepository interface {
		CreatePost(ctx context.Context, post *Post) (err error)
		UpdatePost(ctx context.Context, post *Post) (err error)
		DeletePost(ctx context.Context, id uuid.UUID) (err error)
		FindPostByID(ctx context.Context, id uuid.UUID, viewerID ...uuid.UUID) (result Post, err error)
		FindUserPosts(ctx context.Context, personnelID uuid.UUID, options QueryOptions, viewerID ...uuid.UUID) (result []Post, total int64, err error)
		FindFeedPosts(ctx context.Context, personnelID uuid.UUID, options QueryOptions) (result []Post, total int64, err error)
		FindPublicPosts(ctx context.Context, options QueryOptions, viewerID ...uuid.UUID) (result []Post, total int64, err error)
		IncrementPostView(ctx context.Context, postID uuid.UUID) error
		IncrementPostShare(ctx context.Context, postID uuid.UUID) error
		CreateFollow(ctx context.Context, follow *Follow) (err error)
		DeleteFollow(ctx context.Context, followerID, followingID uuid.UUID) (err error)
		IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
		CheckMutualFollow(ctx context.Context, personnelID1, personnelID2 uuid.UUID) (bool, error)
		CreateLike(ctx context.Context, like *Like) (err error)
		DeleteLike(ctx context.Context, personnelID, postID uuid.UUID) (err error)
		CreateComment(ctx context.Context, comment *Comment) (err error)
		DeleteComment(ctx context.Context, id uuid.UUID, personnelID uuid.UUID) (err error)
		FindPostComments(ctx context.Context, postID uuid.UUID, options QueryOptions) (result []Comment, total int64, err error)
		CreateHashtag(ctx context.Context, hashtag *Hashtag) (err error)
		FindHashtagByName(ctx context.Context, name string) (result Hashtag, err error)
		CreatePostHashtag(ctx context.Context, postHashtag *PostHashtag) (err error)
		CreateUserMention(ctx context.Context, postID, mentionedPersonnelID uuid.UUID) error
		GetUserMentions(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error)

		// For direct chat between mutual followers
		FindOrCreateDirectConversation(ctx context.Context, personnelID1, personnelID2 uuid.UUID) (Conversation, error)
		SaveDirectMessage(ctx context.Context, message *Message) error
		FindDirectMessages(ctx context.Context, personnelID1, personnelID2 uuid.UUID, options QueryOptions) ([]Message, int64, error)
	}
)
