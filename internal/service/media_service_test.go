package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/storage"
)

type mockMediaRepo struct {
	uploads map[uuid.UUID]*domain.MediaUpload
	posts   map[uuid.UUID][]domain.PostMedia
}

func newMockMediaRepo() *mockMediaRepo {
	return &mockMediaRepo{
		uploads: make(map[uuid.UUID]*domain.MediaUpload),
		posts:   make(map[uuid.UUID][]domain.PostMedia),
	}
}

func (m *mockMediaRepo) CreateUpload(ctx context.Context, upload *domain.MediaUpload) error {
	m.uploads[upload.ID] = upload
	return nil
}

func (m *mockMediaRepo) GetUploadByID(ctx context.Context, id uuid.UUID) (*domain.MediaUpload, error) {
	u, ok := m.uploads[id]
	if !ok {
		return nil, errors.New("upload not found")
	}
	return u, nil
}

func (m *mockMediaRepo) UpdateUploadStatus(ctx context.Context, id uuid.UUID, status string) error {
	u, ok := m.uploads[id]
	if !ok {
		return errors.New("upload not found")
	}
	u.Status = status
	return nil
}

func (m *mockMediaRepo) CreatePostMedia(ctx context.Context, media *domain.PostMedia) error {
	m.posts[media.PostID] = append(m.posts[media.PostID], *media)
	return nil
}

func (m *mockMediaRepo) GetMediaByPostID(ctx context.Context, postID uuid.UUID) ([]domain.PostMedia, error) {
	return m.posts[postID], nil
}

func (m *mockMediaRepo) GetMediaByID(ctx context.Context, id uuid.UUID) (*domain.PostMedia, error) {
	for _, list := range m.posts {
		for _, item := range list {
			if item.ID == id {
				return &item, nil
			}
		}
	}
	return nil, errors.New("media not found")
}

func (m *mockMediaRepo) UpdatePostMediaStatus(ctx context.Context, id uuid.UUID, status, thumbKey string, width, height, duration int) error {
	return nil
}

func (m *mockMediaRepo) DeletePostMedia(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestMediaService_InitiateAndComplete(t *testing.T) {
	repo := newMockMediaRepo()
	mockStorage := storage.NewMockStorageProvider("test-bucket", "https://cdn.test.com/test-bucket")
	mediaSvc := NewMediaService(repo, mockStorage)

	ctx := context.Background()
	ownerID := uuid.Must(uuid.NewV4())

	// Test InitiateUpload
	req := domain.InitiateUploadInput{
		Filename:  "vacation.jpg",
		MIMEType:  "image/jpeg",
		FileSize:  2 * 1024 * 1024,
		MediaType: "IMAGE",
	}

	resp, err := mediaSvc.InitiateUpload(ctx, ownerID, req)
	if err != nil {
		t.Fatalf("expected no error initiating upload, got %v", err)
	}
	if resp.UploadID == uuid.Nil {
		t.Errorf("expected non-nil upload ID")
	}
	if resp.UploadURL == "" {
		t.Errorf("expected presigned upload URL")
	}

	// Simulate client uploading bytes to presigned URL
	err = mockStorage.PutObject(ctx, resp.Bucket, resp.ObjectKey, []byte("fake-image-bytes"), "image/jpeg")
	if err != nil {
		t.Fatalf("expected no error putting object in mock storage, got %v", err)
	}

	// Test CompleteUpload
	completed, err := mediaSvc.CompleteUpload(ctx, ownerID, resp.UploadID)
	if err != nil {
		t.Fatalf("expected no error completing upload, got %v", err)
	}
	if completed.Status != "READY" {
		t.Errorf("expected status READY, got %s", completed.Status)
	}

	// Test AttachMediaToPost
	postID := uuid.Must(uuid.NewV4())
	attached, err := mediaSvc.AttachMediaToPost(ctx, postID, ownerID, []uuid.UUID{completed.ID})
	if err != nil {
		t.Fatalf("expected no error attaching media to post, got %v", err)
	}
	if len(attached) != 1 {
		t.Fatalf("expected 1 attached media item, got %d", len(attached))
	}
	if attached[0].URL == "" {
		t.Errorf("expected resolved public URL on attached media")
	}
}
