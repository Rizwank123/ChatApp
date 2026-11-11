package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
)

var (
	hashtagRegex = regexp.MustCompile(`#([A-Za-z0-9_]+)`)
	mentionRegex = regexp.MustCompile(`@([A-Za-z0-9_]+)`)
)

type postService struct {
	repo           domain.PostRepository
	mediaService   domain.MediaService
	feedEngine     domain.FeedFanoutEngine
	mediaProcessor domain.MediaProcessorEnqueuer
}

func NewPostService(repo domain.PostRepository) domain.PostService {
	return &postService{repo: repo}
}

func NewEnhancedPostService(
	repo domain.PostRepository,
	mediaService domain.MediaService,
	feedEngine domain.FeedFanoutEngine,
	mediaProcessor domain.MediaProcessorEnqueuer,
) domain.PostService {
	return &postService{
		repo:           repo,
		mediaService:   mediaService,
		feedEngine:     feedEngine,
		mediaProcessor: mediaProcessor,
	}
}

func (s *postService) CreatePost(input domain.CreatePostInput, userID uuid.UUID) (*domain.Post, error) {
	ctx := context.Background()

	status := input.Status
	if status == "" {
		status = "PUBLISHED"
	}
	visibility := input.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}

	post := &domain.Post{
		PersonnelID: userID,
		Caption:     input.Caption,
		MediaURL:    input.MediaURL,
		Location:    input.Location,
		Visibility:  visibility,
		Status:      status,
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	// 1. Attach media from media_uploads if provided
	if len(input.MediaIDs) > 0 && s.mediaService != nil {
		attachedMedia, err := s.mediaService.AttachMediaToPost(ctx, post.ID, userID, input.MediaIDs)
		if err == nil {
			post.Media = attachedMedia
			if len(post.MediaURL) == 0 {
				for _, m := range attachedMedia {
					post.MediaURL = append(post.MediaURL, m.URL)
				}
			}
			// Trigger async background processing for dimensions & thumbnails
			if s.mediaProcessor != nil {
				for _, m := range attachedMedia {
					s.mediaProcessor.Enqueue(m.ID)
				}
			}
		}
	}

	// 2. Extract and associate hashtags
	tags := input.Hashtags
	if len(tags) == 0 && post.Caption != "" {
		matches := hashtagRegex.FindAllStringSubmatch(post.Caption, -1)
		for _, m := range matches {
			if len(m) > 1 {
				tags = append(tags, strings.ToLower(m[1]))
			}
		}
	}
	for _, tagName := range tags {
		cleanTag := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tagName)), "#")
		if cleanTag == "" {
			continue
		}
		hashtag, err := s.repo.FindHashtagByName(ctx, cleanTag)
		if err != nil {
			hashtag = domain.Hashtag{Name: cleanTag}
			if err := s.repo.CreateHashtag(ctx, &hashtag); err != nil {
				continue
			}
		}
		_ = s.repo.CreatePostHashtag(ctx, &domain.PostHashtag{
			PostID:    post.ID,
			HashtagID: hashtag.ID,
		})
	}

	// 3. Fanout to followers + author's own timeline via hybrid feed engine
	if s.feedEngine != nil && post.Status == "PUBLISHED" {
		_ = s.feedEngine.FanoutPost(ctx, post)
	}

	// 4. Resolve media URLs for response
	if s.mediaService != nil && len(post.Media) > 0 {
		s.mediaService.ResolveMediaListURLs(post.Media)
	}

	return post, nil
}

func (s *postService) UpdatePost(id uuid.UUID, input domain.UpdatePostInput, userID uuid.UUID) error {
	ctx := context.Background()

	post, err := s.repo.FindPostByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if post.PersonnelID != userID {
		return errors.New("unauthorized to update this post")
	}

	if input.Caption != nil {
		post.Caption = *input.Caption
	}
	if input.MediaURL != nil {
		post.MediaURL = *input.MediaURL
	}
	if input.Location != nil {
		post.Location = *input.Location
	}
	if input.Visibility != nil {
		post.Visibility = *input.Visibility
	}
	if input.Status != nil {
		post.Status = *input.Status
	}
	if input.IsArchived != nil {
		post.IsArchived = *input.IsArchived
	}

	return s.repo.UpdatePost(ctx, &post)
}

func (s *postService) DeletePost(id uuid.UUID, userID uuid.UUID) error {
	ctx := context.Background()

	post, err := s.repo.FindPostByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if post.PersonnelID != userID {
		return errors.New("unauthorized to delete this post")
	}

	return s.repo.DeletePost(ctx, id)
}

func (s *postService) GetPost(id uuid.UUID, viewerID ...uuid.UUID) (domain.Post, error) {
	ctx := context.Background()
	post, err := s.repo.FindPostByID(ctx, id, viewerID...)
	if err != nil {
		return domain.Post{}, err
	}
	if s.mediaService != nil && len(post.Media) > 0 {
		s.mediaService.ResolveMediaListURLs(post.Media)
	}
	return post, nil
}

func (s *postService) GetUserPosts(userID uuid.UUID, options domain.QueryOptions, viewerID ...uuid.UUID) ([]domain.Post, int64, error) {
	posts, total, err := s.repo.FindUserPosts(context.Background(), userID, options, viewerID...)
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

func (s *postService) GetFeedPosts(userID uuid.UUID, options domain.QueryOptions) ([]domain.Post, int64, error) {
	posts, total, err := s.repo.FindFeedPosts(context.Background(), userID, options)
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

func (s *postService) GetPublicPosts(options domain.QueryOptions, viewerID ...uuid.UUID) ([]domain.Post, int64, error) {
	posts, total, err := s.repo.FindPublicPosts(context.Background(), options, viewerID...)
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

func (s *postService) RecordView(postID uuid.UUID) error {
	return s.repo.IncrementPostView(context.Background(), postID)
}

func (s *postService) RecordShare(postID uuid.UUID) error {
	return s.repo.IncrementPostShare(context.Background(), postID)
}

func (s *postService) FollowUser(followerID, followingID uuid.UUID) error {
	ctx := context.Background()
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	isFollowing, err := s.repo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return err
	}
	if isFollowing {
		return errors.New("already following this user")
	}

	follow := &domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return s.repo.CreateFollow(ctx, follow)
}

func (s *postService) UnfollowUser(followerID, followingID uuid.UUID) error {
	return s.repo.DeleteFollow(context.Background(), followerID, followingID)
}

func (s *postService) CheckFollowing(followerID, followingID uuid.UUID) (bool, error) {
	return s.repo.IsFollowing(context.Background(), followerID, followingID)
}

func (s *postService) CheckMutualFollow(userID1, userID2 uuid.UUID) (bool, error) {
	return s.repo.CheckMutualFollow(context.Background(), userID1, userID2)
}

func (s *postService) LikePost(userID, postID uuid.UUID) error {
	if _, err := s.repo.FindPostByID(context.Background(), postID, userID); err != nil {
		return err
	}
	like := &domain.Like{
		PersonnelID: userID,
		PostID:      postID,
	}
	return s.repo.CreateLike(context.Background(), like)
}

func (s *postService) UnlikePost(userID, postID uuid.UUID) error {
	return s.repo.DeleteLike(context.Background(), userID, postID)
}

func (s *postService) CreateComment(postID uuid.UUID, content string, userID uuid.UUID, parentID *uuid.UUID) error {
	if _, err := s.repo.FindPostByID(context.Background(), postID, userID); err != nil {
		return err
	}
	comment := &domain.Comment{
		PersonnelID: userID,
		PostID:      postID,
		ParentID:    parentID,
		Content:     content,
	}
	return s.repo.CreateComment(context.Background(), comment)
}

func (s *postService) DeleteComment(commentID uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteComment(context.Background(), commentID, userID)
}

func (s *postService) GetPostComments(postID uuid.UUID, options domain.QueryOptions) ([]domain.Comment, int64, error) {
	return s.repo.FindPostComments(context.Background(), postID, options)
}

func (s *postService) CreateDirectChat(senderID, receiverID uuid.UUID) (domain.Conversation, error) {
	ctx := context.Background()
	isMutual, err := s.CheckMutualFollow(senderID, receiverID)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("error checking mutual follow status: %w", err)
	}
	if !isMutual {
		return domain.Conversation{}, errors.New("direct chat is only available between mutual followers")
	}
	return s.repo.FindOrCreateDirectConversation(ctx, senderID, receiverID)
}

func (s *postService) SendDirectMessage(senderID, receiverID uuid.UUID, content string, messageType domain.MessageType) (domain.Message, error) {
	ctx := context.Background()
	isMutual, err := s.CheckMutualFollow(senderID, receiverID)
	if err != nil {
		return domain.Message{}, fmt.Errorf("error checking mutual follow status: %w", err)
	}
	if !isMutual {
		return domain.Message{}, errors.New("direct messaging is only available between mutual followers")
	}

	conversation, err := s.repo.FindOrCreateDirectConversation(ctx, senderID, receiverID)
	if err != nil {
		return domain.Message{}, fmt.Errorf("error finding or creating conversation: %w", err)
	}

	message := &domain.Message{
		Base: domain.Base{
			ID: uuid.Must(uuid.NewV4()),
		},
		ConversationID: conversation.ID,
		SenderID:       senderID,
		MessageType:    messageType,
		Content:        content,
	}

	if err := s.repo.SaveDirectMessage(ctx, message); err != nil {
		return domain.Message{}, fmt.Errorf("error saving message: %w", err)
	}

	return *message, nil
}

func (s *postService) GetDirectMessages(userID1, userID2 uuid.UUID, options domain.QueryOptions) ([]domain.Message, int64, error) {
	ctx := context.Background()
	isMutual, err := s.CheckMutualFollow(userID1, userID2)
	if err != nil {
		return nil, 0, fmt.Errorf("error checking mutual follow status: %w", err)
	}
	if !isMutual {
		return nil, 0, errors.New("direct message history is only available between mutual followers")
	}
	return s.repo.FindDirectMessages(ctx, userID1, userID2, options)
}
