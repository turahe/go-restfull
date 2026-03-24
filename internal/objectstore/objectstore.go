package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/turahe/go-restfull/internal/config"

	"go.uber.org/zap"
)

var (
	errBucketRequired = errors.New("object storage bucket is required")
	errPresignFailed    = errors.New("failed to presign get object")
)

// Store is object storage for media (S3-compatible backends or Google Cloud Storage).
type Store interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// NewFromConfig builds a Store for the configured MEDIA_STORAGE backend (s3 or gcs).
func NewFromConfig(cfg config.Config, log *zap.Logger) (Store, error) {
	switch cfg.MediaStorage {
	case "s3":
		return newS3Store(cfg, log)
	case "gcs":
		return newGCSStore(cfg, log)
	default:
		return nil, fmt.Errorf("unknown MEDIA_STORAGE: %q", cfg.MediaStorage)
	}
}
