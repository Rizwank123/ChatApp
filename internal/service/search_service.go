package service

import (
	"context"
	"sync"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type searchService struct {
	searchRepo   domain.SearchRepository
	mediaService domain.MediaService
}

// NewSearchService instantiates the search service.
func NewSearchService(searchRepo domain.SearchRepository, mediaService domain.MediaService) domain.SearchService {
	return &searchService{
		searchRepo:   searchRepo,
		mediaService: mediaService,
	}
}

func (s *searchService) UnifiedSearch(ctx context.Context, viewerID uuid.UUID, query string, limit int) (*domain.UnifiedSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	result := &domain.UnifiedSearchResult{
		Users:    []domain.UserSearchResult{},
		Posts:    []domain.Post{},
		Hashtags: []domain.HashtagSearchResult{},
	}

	var wg sync.WaitGroup
	var userErr, postErr, tagErr error

	wg.Add(3)

	go func() {
		defer wg.Done()
		users, _, err := s.searchRepo.SearchUsers(ctx, viewerID, query, limit, 0)
		if err != nil {
			userErr = err
			return
		}
		result.Users = users
	}()

	go func() {
		defer wg.Done()
		posts, _, err := s.searchRepo.SearchPosts(ctx, viewerID, query, limit, 0)
		if err != nil {
			postErr = err
			return
		}
		if s.mediaService != nil {
			for i := range posts {
				s.mediaService.ResolveMediaListURLs(posts[i].Media)
			}
		}
		result.Posts = posts
	}()

	go func() {
		defer wg.Done()
		tags, _, err := s.searchRepo.SearchHashtags(ctx, query, limit, 0)
		if err != nil {
			tagErr = err
			return
		}
		result.Hashtags = tags
	}()

	wg.Wait()

	if userErr != nil && postErr != nil && tagErr != nil {
		return nil, userErr
	}

	return result, nil
}

func (s *searchService) SearchUsers(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]domain.UserSearchResult, int64, error) {
	return s.searchRepo.SearchUsers(ctx, viewerID, query, limit, offset)
}

func (s *searchService) SearchPosts(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]domain.Post, int64, error) {
	posts, total, err := s.searchRepo.SearchPosts(ctx, viewerID, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if s.mediaService != nil {
		for i := range posts {
			s.mediaService.ResolveMediaListURLs(posts[i].Media)
		}
	}
	return posts, total, nil
}

func (s *searchService) SearchHashtags(ctx context.Context, query string, limit, offset int) ([]domain.HashtagSearchResult, int64, error) {
	return s.searchRepo.SearchHashtags(ctx, query, limit, offset)
}

func (s *searchService) GetHashtagPosts(ctx context.Context, viewerID uuid.UUID, tag string, limit, offset int) ([]domain.Post, int64, error) {
	posts, total, err := s.searchRepo.GetPostsByHashtag(ctx, viewerID, tag, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if s.mediaService != nil {
		for i := range posts {
			s.mediaService.ResolveMediaListURLs(posts[i].Media)
		}
	}
	return posts, total, nil
}

func (s *searchService) GetHashtagDetails(ctx context.Context, tag string) (*domain.HashtagSearchResult, error) {
	return s.searchRepo.GetHashtagDetails(ctx, tag)
}
