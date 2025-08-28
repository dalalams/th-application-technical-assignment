package storage

import (
	"context"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type ObjectStorage interface {
	EnsureBucket(ctx context.Context) error
	GeneratePresignedPutURL(ctx context.Context, key string, expiry time.Duration) (*url.URL, error)
	GenerateKey(seriesID, episodeID uuid.UUID, filename string) string
	GetBucketName() string
}
