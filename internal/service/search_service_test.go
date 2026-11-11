package service

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/storage"
)

type mockSearchRepo struct {
	users    []domain.UserSearchResult
	posts    []domain.Post
	hashtags []domain.HashtagSearchResult
}

func (m *mockSearchRepo) SearchUsers(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]domain.UserSearchResult, int64, error) {
	return m.users, int64(len(m.users)), nil
}

func (m *mockSearchRepo) SearchPosts(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]domain.Post, int64, error) {
	return m.posts, int64(len(m.posts)), nil
}

func (m *mockSearchRepo) SearchHashtags(ctx context.Context, query string, limit, offset int) ([]domain.HashtagSearchResult, int64, error) {
	return m.hashtags, int64(len(m.hashtags)), nil
}

func (m *mockSearchRepo) GetPostsByHashtag(ctx context.Context, viewerID uuid.UUID, tag string, limit, offset int) ([]domain.Post, int64, error) {
	return m.posts, int64(len(m.posts)), nil
}

func (m *mockSearchRepo) GetHashtagDetails(ctx context.Context, name string) (*domain.HashtagSearchResult, error) {
	if len(m.hashtags) > 0 {
		return &m.hashtags[0], nil
	}
	return nil, nil
}

func TestSearchService_UnifiedSearch(t *testing.T) {
	repo := &mockSearchRepo{
		users: []domain.UserSearchResult{
			{UserName: "alex_coder", DisplayName: "Alex Developer", MatchScore: 0.9},
		},
		hashtags: []domain.HashtagSearchResult{
			{Name: "golang", PostCount: 42, MatchScore: 0.8},
		},
		posts: []domain.Post{
			{Caption: "Building a new search engine in Go #golang"},
		},
	}

	mockStorage := storage.NewMockStorageProvider("test-bucket", "https://cdn.test.com")
	mediaSvc := NewMediaService(newMockMediaRepo(), mockStorage)
	searchSvc := NewSearchService(repo, mediaSvc)

	ctx := context.Background()
	viewerID := uuid.Must(uuid.NewV4())

	result, err := searchSvc.UnifiedSearch(ctx, viewerID, "go", 10)
	if err != nil {
		t.Fatalf("expected no error from UnifiedSearch, got %v", err)
	}

	if len(result.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(result.Users))
	}
	if len(result.Hashtags) != 1 {
		t.Errorf("expected 1 hashtag, got %d", len(result.Hashtags))
	}
	if len(result.Posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(result.Posts))
	}
}
