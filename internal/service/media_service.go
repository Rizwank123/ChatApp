package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
)

var (
	ErrInvalidMediaType = errors.New("unsupported media type: only IMAGE and VIDEO are allowed")
	ErrFileTooLarge     = errors.New("file size exceeds maximum allowed limit")
	ErrInvalidExtension = errors.New("invalid file extension")
	ErrObjectNotFound   = errors.New("uploaded object not found in storage")
	ErrUnauthorized     = errors.New("unauthorized action on media object")
)

type mediaService struct {
	mediaRepo domain.MediaRepository
	storage   domain.StorageProvider
}

// NewMediaService instantiates a media service for presigned uploads and URL resolution.
func NewMediaService(mediaRepo domain.MediaRepository, storage domain.StorageProvider) domain.MediaService {
	return &mediaService{
		mediaRepo: mediaRepo,
		storage:   storage,
	}
}

func (s *mediaService) InitiateUpload(ctx context.Context, ownerID uuid.UUID, input domain.InitiateUploadInput) (*domain.InitiateUploadOutput, error) {
	mediaType := strings.ToUpper(input.MediaType)
	if mediaType != "IMAGE" && mediaType != "VIDEO" {
		return nil, ErrInvalidMediaType
	}

	// Max 30MB for images, 100MB for video
	if mediaType == "IMAGE" && input.FileSize > 30*1024*1024 {
		return nil, ErrFileTooLarge
	}
	if mediaType == "VIDEO" && input.FileSize > 100*1024*1024 {
		return nil, ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(input.Filename))
	if !isAllowedExtension(mediaType, ext) {
		return nil, ErrInvalidExtension
	}

	mediaID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	bucket := s.storage.BucketName()
	folder := "images"
	if mediaType == "VIDEO" {
		folder = "videos"
	}

	// Predictable hierarchical key
	objectKey := fmt.Sprintf("social-media/users/%s/uploads/%s/%s%s", ownerID.String(), folder, mediaID.String(), ext)
	expiresAt := time.Now().Add(15 * time.Minute)

	upload := &domain.MediaUpload{
		ID:               mediaID,
		OwnerID:          ownerID,
		Bucket:           bucket,
		ObjectKey:        objectKey,
		OriginalFilename: input.Filename,
		MIMEType:         input.MIMEType,
		FileSize:         input.FileSize,
		MediaType:        mediaType,
		Status:           "PENDING",
		ExpiresAt:        expiresAt,
	}

	if err := s.mediaRepo.CreateUpload(ctx, upload); err != nil {
		return nil, fmt.Errorf("failed to create media upload record: %w", err)
	}

	presignedURL, err := s.storage.GeneratePresignedUploadURL(ctx, bucket, objectKey, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned upload url: %w", err)
	}

	return &domain.InitiateUploadOutput{
		UploadID:  mediaID,
		UploadURL: presignedURL,
		ObjectKey: objectKey,
		Bucket:    bucket,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *mediaService) CompleteUpload(ctx context.Context, ownerID uuid.UUID, uploadID uuid.UUID) (*domain.MediaUpload, error) {
	upload, err := s.mediaRepo.GetUploadByID(ctx, uploadID)
	if err != nil {
		return nil, err
	}

	if upload.OwnerID != ownerID {
		return nil, ErrUnauthorized
	}

	// Verify object physically exists in MinIO
	size, _, err := s.storage.StatObject(ctx, upload.Bucket, upload.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrObjectNotFound, err)
	}

	upload.FileSize = size
	upload.Status = "READY"

	if err := s.mediaRepo.UpdateUploadStatus(ctx, uploadID, "READY"); err != nil {
		return nil, err
	}

	return upload, nil
}

func (s *mediaService) GetMedia(ctx context.Context, mediaID uuid.UUID) (*domain.PostMedia, error) {
	media, err := s.mediaRepo.GetMediaByID(ctx, mediaID)
	if err != nil {
		return nil, err
	}
	s.ResolveMediaURLs(media)
	return media, nil
}

func (s *mediaService) AttachMediaToPost(ctx context.Context, postID, ownerID uuid.UUID, uploadIDs []uuid.UUID) ([]domain.PostMedia, error) {
	var result []domain.PostMedia
	for idx, uploadID := range uploadIDs {
		upload, err := s.mediaRepo.GetUploadByID(ctx, uploadID)
		if err != nil {
			return nil, fmt.Errorf("invalid media upload %s: %w", uploadID, err)
		}
		if upload.OwnerID != ownerID {
			return nil, ErrUnauthorized
		}

		pm := &domain.PostMedia{
			PostID:           postID,
			OwnerID:          ownerID,
			Bucket:           upload.Bucket,
			ObjectKey:        upload.ObjectKey,
			OriginalFilename: upload.OriginalFilename,
			MIMEType:         upload.MIMEType,
			FileSize:         upload.FileSize,
			MediaType:        upload.MediaType,
			Status:           "READY",
			DisplayOrder:     idx,
		}

		if err := s.mediaRepo.CreatePostMedia(ctx, pm); err != nil {
			return nil, fmt.Errorf("failed to attach media to post: %w", err)
		}

		s.ResolveMediaURLs(pm)
		result = append(result, *pm)
	}
	return result, nil
}

func (s *mediaService) ResolveMediaURLs(media *domain.PostMedia) {
	if media == nil {
		return
	}
	media.URL = s.storage.ResolvePublicURL(media.Bucket, media.ObjectKey)
	if media.ThumbnailKey != "" {
		media.ThumbnailURL = s.storage.ResolvePublicURL(media.Bucket, media.ThumbnailKey)
	} else if media.MediaType == "IMAGE" {
		media.ThumbnailURL = media.URL
	}
}

func (s *mediaService) ResolveMediaListURLs(mediaList []domain.PostMedia) {
	for i := range mediaList {
		s.ResolveMediaURLs(&mediaList[i])
	}
}

func isAllowedExtension(mediaType, ext string) bool {
	if mediaType == "IMAGE" {
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".heic":
			return true
		}
		return false
	}
	if mediaType == "VIDEO" {
		switch ext {
		case ".mp4", ".mov", ".webm", ".mkv":
			return true
		}
		return false
	}
	return false
}
