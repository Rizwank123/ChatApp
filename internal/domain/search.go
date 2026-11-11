package domain

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

// SearchType specifies search category filter.
type SearchType string

const (
	SearchTypeAll      SearchType = "all"
	SearchTypeUsers    SearchType = "users"
	SearchTypePosts    SearchType = "posts"
	SearchTypeHashtags SearchType = "hashtags"
)

// UserSearchResult represents an account matched in search.
type UserSearchResult struct {
	PersonnelID uuid.UUID `json:"personnel_id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	DisplayName string    `json:"display_name"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	IsFollowing bool      `json:"is_following"`
	FollowsYou  bool      `json:"follows_you"`
	IsMutual    bool      `json:"is_mutual"`
	IsPrivate   bool      `json:"is_private"`
	MatchScore  float64   `json:"match_score"`
}

// PostSearchResult represents a post matched in search.
type PostSearchResult struct {
	Post       Post    `json:"post"`
	MatchScore float64 `json:"match_score"`
}

// HashtagSearchResult represents a hashtag with popularity count.
type HashtagSearchResult struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	PostCount  int64     `json:"post_count"`
	MatchScore float64   `json:"match_score"`
}

// UnifiedSearchResult represents the multi-category search response.
type UnifiedSearchResult struct {
	Users    []UserSearchResult    `json:"users"`
	Posts    []Post                `json:"posts"`
	Hashtags []HashtagSearchResult `json:"hashtags"`
}

// SearchRepository defines database search queries leveraging pg_trgm and tsvector.
type SearchRepository interface {
	SearchUsers(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]UserSearchResult, int64, error)
	SearchPosts(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]Post, int64, error)
	SearchHashtags(ctx context.Context, query string, limit, offset int) ([]HashtagSearchResult, int64, error)
	GetPostsByHashtag(ctx context.Context, viewerID uuid.UUID, tag string, limit, offset int) ([]Post, int64, error)
	GetHashtagDetails(ctx context.Context, name string) (*HashtagSearchResult, error)
}

// SearchService provides unified and categorized search logic.
type SearchService interface {
	UnifiedSearch(ctx context.Context, viewerID uuid.UUID, query string, limit int) (*UnifiedSearchResult, error)
	SearchUsers(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]UserSearchResult, int64, error)
	SearchPosts(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]Post, int64, error)
	SearchHashtags(ctx context.Context, query string, limit, offset int) ([]HashtagSearchResult, int64, error)
	GetHashtagPosts(ctx context.Context, viewerID uuid.UUID, tag string, limit, offset int) ([]Post, int64, error)
	GetHashtagDetails(ctx context.Context, tag string) (*HashtagSearchResult, error)
}
