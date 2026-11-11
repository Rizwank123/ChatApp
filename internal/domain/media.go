package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// PostMedia represents an image or video attached to a post.
type PostMedia struct {
	ID               uuid.UUID `json:"id"`
	PostID           uuid.UUID `json:"post_id"`
	OwnerID          uuid.UUID `json:"owner_id"`
	Bucket           string    `json:"bucket"`
	ObjectKey        string    `json:"object_key"`
	OriginalFilename string    `json:"original_filename"`
	MIMEType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	MediaType        string    `json:"media_type"` // IMAGE, VIDEO
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	Duration         int       `json:"duration"`
	ThumbnailKey     string    `json:"thumbnail_key,omitempty"`
	URL              string    `json:"url"`
	ThumbnailURL     string    `json:"thumbnail_url,omitempty"`
	Status           string    `json:"status"` // READY, PROCESSING, FAILED
	DisplayOrder     int       `json:"display_order"`
	BaseAudit
}

// MediaUpload tracks a presigned direct-upload session.
type MediaUpload struct {
	ID               uuid.UUID `json:"id"`
	OwnerID          uuid.UUID `json:"owner_id"`
	Bucket           string    `json:"bucket"`
	ObjectKey        string    `json:"object_key"`
	OriginalFilename string    `json:"original_filename"`
	MIMEType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	MediaType        string    `json:"media_type"` // IMAGE, VIDEO
	Status           string    `json:"status"`     // PENDING, UPLOADED, READY, FAILED
	UploadURL        string    `json:"upload_url,omitempty"`
	ExpiresAt        time.Time `json:"expires_at"`
	BaseAudit
}

// MediaProcessingJob tracks asynchronous media transcoding/thumbnailing.
type MediaProcessingJob struct {
	ID           uuid.UUID  `json:"id"`
	MediaID      *uuid.UUID `json:"media_id,omitempty"`
	UploadID     *uuid.UUID `json:"upload_id,omitempty"`
	JobType      string     `json:"job_type"` // IMAGE_OPTIMIZE, THUMBNAIL_GEN, VIDEO_METADATA
	Status       string     `json:"status"`   // PENDING, PROCESSING, COMPLETED, FAILED
	RetryCount   int        `json:"retry_count"`
	MaxRetries   int        `json:"max_retries"`
	ErrorMessage string     `json:"error_message,omitempty"`
	BaseAudit
}

// InitiateUploadInput represents the client payload to start a media upload.
type InitiateUploadInput struct {
	Filename  string `json:"filename" validate:"required"`
	MIMEType  string `json:"mime_type" validate:"required"`
	FileSize  int64  `json:"file_size" validate:"required,min=1,max=104857600"` // Max 100MB
	MediaType string `json:"media_type" validate:"required,oneof=IMAGE VIDEO"`
}

// InitiateUploadOutput returns the presigned PUT URL and upload session details.
type InitiateUploadOutput struct {
	UploadID  uuid.UUID `json:"upload_id"`
	UploadURL string    `json:"upload_url"`
	ObjectKey string    `json:"object_key"`
	Bucket    string    `json:"bucket"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CompleteUploadInput marks an upload finished and enqueues verification.
type CompleteUploadInput struct {
	UploadID uuid.UUID `json:"upload_id" validate:"required"`
}

// StorageProvider abstracts MinIO and S3-compatible cloud object storage.
type StorageProvider interface {
	// GeneratePresignedUploadURL generates a presigned PUT URL with an expiration period.
	GeneratePresignedUploadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
	// GeneratePresignedGetURL generates a presigned GET URL for private objects.
	GeneratePresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
	// StatObject checks if an object exists and returns its size and content-type.
	StatObject(ctx context.Context, bucket, objectKey string) (int64, string, error)
	// PutObject stores a stream into object storage.
	PutObject(ctx context.Context, bucket, objectKey string, data []byte, contentType string) error
	// GetObject retrieves object data.
	GetObject(ctx context.Context, bucket, objectKey string) ([]byte, error)
	// DeleteObject deletes an object.
	DeleteObject(ctx context.Context, bucket, objectKey string) error
	// ResolvePublicURL creates the stable canonical or CDN URL for public objects.
	ResolvePublicURL(bucket, objectKey string) string
	// BucketName returns the default bucket.
	BucketName() string
}

// MediaRepository defines DB operations for media uploads and post media.
type MediaRepository interface {
	CreateUpload(ctx context.Context, upload *MediaUpload) error
	GetUploadByID(ctx context.Context, id uuid.UUID) (*MediaUpload, error)
	UpdateUploadStatus(ctx context.Context, id uuid.UUID, status string) error
	CreatePostMedia(ctx context.Context, media *PostMedia) error
	GetMediaByPostID(ctx context.Context, postID uuid.UUID) ([]PostMedia, error)
	GetMediaByID(ctx context.Context, id uuid.UUID) (*PostMedia, error)
	UpdatePostMediaStatus(ctx context.Context, id uuid.UUID, status, thumbKey string, width, height, duration int) error
	DeletePostMedia(ctx context.Context, id uuid.UUID) error
}

// MediaService provides high-level upload lifecycle and URL resolution.
type MediaService interface {
	InitiateUpload(ctx context.Context, ownerID uuid.UUID, input InitiateUploadInput) (*InitiateUploadOutput, error)
	CompleteUpload(ctx context.Context, ownerID uuid.UUID, uploadID uuid.UUID) (*MediaUpload, error)
	GetMedia(ctx context.Context, mediaID uuid.UUID) (*PostMedia, error)
	AttachMediaToPost(ctx context.Context, postID, ownerID uuid.UUID, uploadIDs []uuid.UUID) ([]PostMedia, error)
	ResolveMediaURLs(media *PostMedia)
	ResolveMediaListURLs(mediaList []PostMedia)
}
