package objectstore

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
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
		Scheme:  storage.SigningSchemeV4,
	})
	if err != nil {
		// Some runtimes can upload/list objects but cannot sign URLs (no signing creds).
		// Fall back to the canonical object URL so API responses still include downloadUrl.
		// This works for public buckets and keeps behavior observable instead of silently empty URLs.
		fallback := g.publicObjectURL(key)
		g.log.Warn("failed to sign GCS URL; falling back to public URL", zap.Error(err), zap.String("bucket", g.bucket), zap.String("key", key))
		return fallback, nil
	}
	return u, nil
}

func (g *gcsStore) publicObjectURL(key string) string {
	parts := strings.Split(strings.TrimPrefix(key, "/"), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	escapedKey := strings.Join(parts, "/")
	return "https://storage.googleapis.com/" + g.bucket + "/" + escapedKey
}

func (g *gcsStore) Delete(ctx context.Context, key string) error {
	err := g.client.Bucket(g.bucket).Object(key).Delete(ctx)
	if err == nil || errors.Is(err, storage.ErrObjectNotExist) {
		return nil
	}
	return err
}
