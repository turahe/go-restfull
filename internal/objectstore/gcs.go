package objectstore

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/turahe/go-restfull/internal/config"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

type gcsStore struct {
	client *storage.Client
	bucket string
	log    *zap.Logger
}

func newGCSStore(cfg config.Config, log *zap.Logger) (*gcsStore, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Error("failed to create GCS client", zap.Error(err))
		return nil, err
	}
	return &gcsStore{client: client, bucket: cfg.GCSBucket, log: log}, nil
}

func (g *gcsStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	w := g.client.Bucket(g.bucket).Object(key).NewWriter(ctx)
	w.ContentType = contentType
	if size > 0 {
		w.Size = size
	}
	if _, err := io.Copy(w, r); err != nil {
		_ = w.CloseWithError(err)
		return err
	}
	return w.Close()
}

func (g *gcsStore) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := g.client.Bucket(g.bucket).SignedURL(key, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	})
	if err != nil {
		g.log.Error("failed to sign GCS URL", zap.Error(err))
		return "", errPresignFailed
	}
	return u, nil
}

func (g *gcsStore) Delete(ctx context.Context, key string) error {
	err := g.client.Bucket(g.bucket).Object(key).Delete(ctx)
	if err == nil || errors.Is(err, storage.ErrObjectNotExist) {
		return nil
	}
	return err
}
