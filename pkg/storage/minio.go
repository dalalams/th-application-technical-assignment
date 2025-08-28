package storage

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient interface {
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	PresignedPutObject(ctx context.Context, bucketName, objectName string, expiration time.Duration) (*url.URL, error)
}

type MinIOStorage struct {
	client     MinioClient
	bucketName string
}

type UploadResult struct {
	Key      string
	Size     int64
	MimeType string
	ETag     string
}

func NewMinIOClient(cfg *Config) (*MinIOStorage, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOStorage{
		client:     minioClient,
		bucketName: cfg.BucketName,
	}, nil
}

func (m *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

func (m *MinIOStorage) GeneratePresignedPutURL(ctx context.Context, key string, expiry time.Duration) (*url.URL, error) {
	return m.client.PresignedPutObject(ctx, m.bucketName, key, expiry)
}

func (m *MinIOStorage) GenerateKey(seriesID, episodeID uuid.UUID, filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("episodes/%s/%s_%d%s", seriesID.String(), episodeID.String(), timestamp, ext)
}

func (m *MinIOStorage) GetBucketName() string {
	return m.bucketName
}
