package repository

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxMediaRepository struct {
	db *pgxpool.Pool
}

// NewMediaRepository instantiates a PostgreSQL media repository.
func NewMediaRepository(db *pgxpool.Pool) domain.MediaRepository {
	return &pgxMediaRepository{db: db}
}

func (r *pgxMediaRepository) CreateUpload(ctx context.Context, upload *domain.MediaUpload) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if upload.ID.IsNil() {
		newID, _ := uuid.NewV4()
		upload.ID = newID
	}

	query := `
		INSERT INTO media_uploads (id, owner_id, bucket, object_key, original_filename, mime_type, file_size, media_type, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		upload.ID,
		upload.OwnerID,
		upload.Bucket,
		upload.ObjectKey,
		upload.OriginalFilename,
		upload.MIMEType,
		upload.FileSize,
		upload.MediaType,
		upload.Status,
		upload.ExpiresAt,
	).Scan(&upload.CreatedAt, &upload.UpdatedAt)
}

func (r *pgxMediaRepository) GetUploadByID(ctx context.Context, id uuid.UUID) (*domain.MediaUpload, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT id, owner_id, bucket, object_key, original_filename, mime_type, file_size, media_type, status, expires_at, created_at, updated_at
		FROM media_uploads
		WHERE id = $1
	`
	var u domain.MediaUpload
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.OwnerID,
		&u.Bucket,
		&u.ObjectKey,
		&u.OriginalFilename,
		&u.MIMEType,
		&u.FileSize,
		&u.MediaType,
		&u.Status,
		&u.ExpiresAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("media upload not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgxMediaRepository) UpdateUploadStatus(ctx context.Context, id uuid.UUID, status string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `UPDATE media_uploads SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *pgxMediaRepository) CreatePostMedia(ctx context.Context, media *domain.PostMedia) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if media.ID.IsNil() {
		newID, _ := uuid.NewV4()
		media.ID = newID
	}

	query := `
		INSERT INTO post_media (
			id, post_id, owner_id, bucket, object_key, original_filename, mime_type,
			file_size, media_type, width, height, duration, thumbnail_key, status, display_order
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		media.ID,
		media.PostID,
		media.OwnerID,
		media.Bucket,
		media.ObjectKey,
		media.OriginalFilename,
		media.MIMEType,
		media.FileSize,
		media.MediaType,
		media.Width,
		media.Height,
		media.Duration,
		media.ThumbnailKey,
		media.Status,
		media.DisplayOrder,
	).Scan(&media.CreatedAt, &media.UpdatedAt)
}

func (r *pgxMediaRepository) GetMediaByPostID(ctx context.Context, postID uuid.UUID) ([]domain.PostMedia, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT id, post_id, owner_id, bucket, object_key, original_filename, mime_type,
		       file_size, media_type, width, height, duration, thumbnail_key, status, display_order,
		       created_at, updated_at
		FROM post_media
		WHERE post_id = $1
		ORDER BY display_order ASC, created_at ASC
	`
	rows, err := r.db.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PostMedia
	for rows.Next() {
		var m domain.PostMedia
		var thumbKey *string
		var origName *string
		err := rows.Scan(
			&m.ID,
			&m.PostID,
			&m.OwnerID,
			&m.Bucket,
			&m.ObjectKey,
			&origName,
			&m.MIMEType,
			&m.FileSize,
			&m.MediaType,
			&m.Width,
			&m.Height,
			&m.Duration,
			&thumbKey,
			&m.Status,
			&m.DisplayOrder,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if thumbKey != nil {
			m.ThumbnailKey = *thumbKey
		}
		if origName != nil {
			m.OriginalFilename = *origName
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *pgxMediaRepository) GetMediaByID(ctx context.Context, id uuid.UUID) (*domain.PostMedia, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT id, post_id, owner_id, bucket, object_key, original_filename, mime_type,
		       file_size, media_type, width, height, duration, thumbnail_key, status, display_order,
		       created_at, updated_at
		FROM post_media
		WHERE id = $1
	`
	var m domain.PostMedia
	var thumbKey *string
	var origName *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.PostID,
		&m.OwnerID,
		&m.Bucket,
		&m.ObjectKey,
		&origName,
		&m.MIMEType,
		&m.FileSize,
		&m.MediaType,
		&m.Width,
		&m.Height,
		&m.Duration,
		&thumbKey,
		&m.Status,
		&m.DisplayOrder,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("post media not found")
		}
		return nil, err
	}
	if thumbKey != nil {
		m.ThumbnailKey = *thumbKey
	}
	if origName != nil {
		m.OriginalFilename = *origName
	}
	return &m, nil
}

func (r *pgxMediaRepository) UpdatePostMediaStatus(ctx context.Context, id uuid.UUID, status, thumbKey string, width, height, duration int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		UPDATE post_media
		SET status = $1, thumbnail_key = $2, width = $3, height = $4, duration = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, status, thumbKey, width, height, duration, id)
	return err
}

func (r *pgxMediaRepository) DeletePostMedia(ctx context.Context, id uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `DELETE FROM post_media WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
