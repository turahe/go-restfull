package objectstore

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/turahe/go-restfull/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type s3Store struct {
	client *minio.Client
	bucket string
	region string
	log    *zap.Logger
}

func newS3Store(cfg config.Config, log *zap.Logger) (*s3Store, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	}
	if cfg.S3Region != "" {
		opts.Region = cfg.S3Region
	}
	client, err := minio.New(cfg.S3Endpoint, opts)
	if err != nil {
		log.Error("failed to create S3 client", zap.Error(err))
		return nil, err
	}
	return &s3Store{client: client, bucket: cfg.S3Bucket, region: cfg.S3Region, log: log}, nil
}

func (s *s3Store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if s.bucket == "" {
		return errBucketRequired
	}
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		mbOpts := minio.MakeBucketOptions{}
		if s.region != "" {
			mbOpts.Region = s.region
		}
		if err := s.client.MakeBucket(ctx, s.bucket, mbOpts); err != nil {
			return err
		}
	}
	opts := minio.PutObjectOptions{ContentType: contentType}
	_, err = s.client.PutObject(ctx, s.bucket, key, r, size, opts)
	return err
}

func (s *s3Store) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if s.bucket == "" {
		return "", errBucketRequired
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		s.log.Error("failed to presign get object", zap.Error(err))
		return "", errPresignFailed
	}
	if u == nil {
		return "", errPresignFailed
	}
	return u.String(), nil
}

func (s *s3Store) Delete(ctx context.Context, key string) error {
	if s.bucket == "" {
		return errBucketRequired
	}
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		var respErr minio.ErrorResponse
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}
