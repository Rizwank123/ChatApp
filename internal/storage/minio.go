package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/Rizwank123/chatApp/internal/domain"
)

// MinIOConfig holds configuration settings for MinIO / S3 storage.
type MinIOConfig struct {
	Provider        string        `mapstructure:"STORAGE_PROVIDER"`
	Endpoint        string        `mapstructure:"MINIO_ENDPOINT"`
	AccessKey       string        `mapstructure:"MINIO_ACCESS_KEY"`
	SecretKey       string        `mapstructure:"MINIO_SECRET_KEY"`
	Bucket          string        `mapstructure:"MINIO_BUCKET"`
	UseSSL          bool          `mapstructure:"MINIO_USE_SSL"`
	Region          string        `mapstructure:"MINIO_REGION"`
	PublicURL       string        `mapstructure:"MINIO_PUBLIC_URL"`
	PresignedExpiry time.Duration `mapstructure:"MINIO_PRESIGNED_EXPIRY"`
}

type minioStorageProvider struct {
	client    *minio.Client
	config    MinIOConfig
	isMock    bool
	mockStore map[string][]byte
}

// NewMinIOStorageProvider initializes a MinIO client and ensures the default bucket exists.
func NewMinIOStorageProvider(cfg MinIOConfig) (domain.StorageProvider, error) {
	if cfg.Bucket == "" {
		cfg.Bucket = "social-media"
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "localhost:9000"
	}
	if cfg.AccessKey == "" {
		cfg.AccessKey = "minioadmin"
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = "minioadmin"
	}
	if cfg.PresignedExpiry <= 0 {
		cfg.PresignedExpiry = 15 * time.Minute
	}
	if cfg.PublicURL == "" {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		cfg.PublicURL = fmt.Sprintf("%s://%s/%s", scheme, cfg.Endpoint, cfg.Bucket)
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	p := &minioStorageProvider{
		client:    client,
		config:    cfg,
		mockStore: make(map[string][]byte),
	}

	// Try initializing bucket if possible
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err == nil && !exists {
		_ = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region})
	}

	return p, nil
}

// NewMockStorageProvider returns an in-memory storage provider for hermetic testing.
func NewMockStorageProvider(bucket, publicURL string) domain.StorageProvider {
	if bucket == "" {
		bucket = "social-media"
	}
	if publicURL == "" {
		publicURL = "http://localhost:9000/" + bucket
	}
	return &minioStorageProvider{
		config: MinIOConfig{
			Bucket:          bucket,
			PublicURL:       publicURL,
			PresignedExpiry: 15 * time.Minute,
		},
		isMock:    true,
		mockStore: make(map[string][]byte),
	}
}

func (p *minioStorageProvider) BucketName() string {
	return p.config.Bucket
}

func (p *minioStorageProvider) GeneratePresignedUploadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = p.config.PresignedExpiry
	}
	if bucket == "" {
		bucket = p.config.Bucket
	}

	if p.isMock {
		return fmt.Sprintf("http://localhost:9000/%s/%s?upload_session=mock", bucket, objectKey), nil
	}

	u, err := p.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload url: %w", err)
	}
	return u.String(), nil
}

func (p *minioStorageProvider) GeneratePresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = p.config.PresignedExpiry
	}
	if bucket == "" {
		bucket = p.config.Bucket
	}

	if p.isMock {
		return fmt.Sprintf("http://localhost:9000/%s/%s?token=mock_presigned", bucket, objectKey), nil
	}

	reqParams := make(url.Values)
	u, err := p.client.PresignedGetObject(ctx, bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned get url: %w", err)
	}
	return u.String(), nil
}

func (p *minioStorageProvider) StatObject(ctx context.Context, bucket, objectKey string) (int64, string, error) {
	if bucket == "" {
		bucket = p.config.Bucket
	}

	if p.isMock {
		data, ok := p.mockStore[bucket+"/"+objectKey]
		if !ok {
			return 0, "", fmt.Errorf("object %s not found in mock storage", objectKey)
		}
		return int64(len(data)), "application/octet-stream", nil
	}

	info, err := p.client.StatObject(ctx, bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return 0, "", fmt.Errorf("failed to stat object: %w", err)
	}
	return info.Size, info.ContentType, nil
}

func (p *minioStorageProvider) PutObject(ctx context.Context, bucket, objectKey string, data []byte, contentType string) error {
	if bucket == "" {
		bucket = p.config.Bucket
	}

	if p.isMock {
		p.mockStore[bucket+"/"+objectKey] = data
		return nil
	}

	reader := bytes.NewReader(data)
	_, err := p.client.PutObject(ctx, bucket, objectKey, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (p *minioStorageProvider) GetObject(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	if bucket == "" {
		bucket = p.config.Bucket
	}

	if p.isMock {
		data, ok := p.mockStore[bucket+"/"+objectKey]
		if !ok {
			return nil, fmt.Errorf("object not found")
		}
		return data, nil
	}

	obj, err := p.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

func (p *minioStorageProvider) DeleteObject(ctx context.Context, bucket, objectKey string) error {
	if bucket == "" {
		bucket = p.config.Bucket
	}

	if p.isMock {
		delete(p.mockStore, bucket+"/"+objectKey)
		return nil
	}

	return p.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
}

func (p *minioStorageProvider) ResolvePublicURL(bucket, objectKey string) string {
	if bucket == "" {
		bucket = p.config.Bucket
	}
	objectKey = strings.TrimPrefix(objectKey, "/")
	cleanBase := strings.TrimSuffix(p.config.PublicURL, "/")
	// If public URL already includes bucket name, avoid duplicating it
	if strings.HasSuffix(cleanBase, "/"+bucket) {
		return fmt.Sprintf("%s/%s", cleanBase, objectKey)
	}
	return fmt.Sprintf("%s/%s/%s", cleanBase, bucket, objectKey)
}
