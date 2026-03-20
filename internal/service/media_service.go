package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-rest/internal/config"
	"go-rest/internal/handler/request"
	"go-rest/internal/model"
	"go-rest/internal/repository"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrInvalidMedia       = errors.New("invalid media")
	ErrMediaTooLarge      = errors.New("media too large")
	ErrMediaNotFound      = errors.New("media not found")
	ErrInvalidMediaUserID = errors.New("invalid media user id")
	ErrInvalidActorID     = errors.New("invalid actor id")
	ErrInvalidUploadDir   = errors.New("invalid upload dir")
)

type MediaService struct {
	repo      *repository.MediaRepository
	uploadDir string
	maxBytes  int64

	minioClient *minio.Client
	minioBucket string
	log         *zap.Logger
}

func NewMediaService(repo *repository.MediaRepository, cfg config.Config, log *zap.Logger) *MediaService {
	var minioClient *minio.Client
	var minioBucket string
	if cfg.MinioEndpoint != "" && cfg.MinioAccessKey != "" && cfg.MinioSecretKey != "" {
		minioBucket = cfg.MinioBucket
		client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
			Secure: cfg.MinioUseSSL,
		})
		if err != nil {
			log.Error("failed to create minio client", zap.Error(err))
			return nil
		}
		minioClient = client
	}

	return &MediaService{
		repo:        repo,
		uploadDir:   cfg.MediaUploadDir,
		maxBytes:    cfg.MediaMaxUploadBytes,
		minioClient: minioClient,
		minioBucket: minioBucket,
		log:         log,
	}
}

// PresignGet returns a temporary URL for downloading the object.
// When MinIO is not enabled, it returns an empty string.
func (s *MediaService) PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if s.minioClient == nil || s.minioBucket == "" {
		return "", nil
	}
	u, err := s.minioClient.PresignedGetObject(ctx, s.minioBucket, objectKey, expiry, nil)
	if err != nil {
		s.log.Error("failed to presign get object", zap.Error(err))
		return "", errors.New("failed to presign get object")
	}
	if u == nil {
		s.log.Error("failed to presign get object")
		return "", errors.New("failed to presign get object")
	}
	return u.String(), nil
}

// Upload uploads the file to storage and creates a Media row.
// mediaableType/mediaableID attach the media to Post/User/Category/Comment via GORM polymorphism.
func (s *MediaService) Upload(
	ctx context.Context,
	actorUserID uint,
	fh *multipart.FileHeader,
) (*model.Media, error) {
	if actorUserID == 0 {
		return nil, ErrInvalidActorID
	}
	if fh == nil {
		return nil, ErrInvalidMedia
	}
	if fh.Size > s.maxBytes {
		return nil, ErrMediaTooLarge
	}
	// If MinIO is enabled, we don't need local upload directory.
	if s.minioClient == nil && strings.TrimSpace(s.uploadDir) == "" {
		return nil, ErrInvalidUploadDir
	}

	origName := strings.TrimSpace(fh.Filename)
	if origName == "" {
		origName = "upload"
	}
	ext := strings.ToLower(filepath.Ext(origName))
	if ext == "" {
		// try mime -> ext
		if mt := fh.Header.Get("Content-Type"); mt != "" {
			if exts, _ := mime.ExtensionsByType(mt); len(exts) > 0 {
				ext = exts[0]
			}
		}
	}
	if ext == "" {
		ext = ".bin"
	}

	mimeType := fh.Header.Get("Content-Type")
	if mimeType == "" {
		// best-effort fallback from filename
		if mt := mime.TypeByExtension(ext); mt != "" {
			mimeType = mt
		} else {
			mimeType = "application/octet-stream"
		}
	}

	mediaType := "file"
	if strings.HasPrefix(mimeType, "image/") {
		mediaType = "image"
	}

	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	id, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return nil, err
	}

	relDir := time.Now().Format("2006/01/02")
	storageFilename := id + ext
	objectKey := filepath.ToSlash(filepath.Join(relDir, storageFilename))
	var cleanupLocalPath string

	// Store into MinIO when configured; otherwise fallback to local filesystem.
	if s.minioClient != nil {
		if s.minioBucket == "" {
			return nil, errors.New("MINIO_BUCKET is required")
		}
		exists, err := s.minioClient.BucketExists(ctx, s.minioBucket)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := s.minioClient.MakeBucket(ctx, s.minioBucket, minio.MakeBucketOptions{}); err != nil {
				return nil, err
			}
		}

		opts := minio.PutObjectOptions{ContentType: mimeType}
		if _, err := s.minioClient.PutObject(ctx, s.minioBucket, objectKey, f, fh.Size, opts); err != nil {
			return nil, err
		}
	} else {
		fullDir := s.uploadDir
		if strings.TrimSpace(fullDir) == "" {
			return nil, ErrInvalidUploadDir
		}
		fullDir = filepath.Clean(fullDir)
		destDir := filepath.Join(fullDir, relDir)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return nil, err
		}

		destPath := filepath.Join(destDir, storageFilename)
		cleanupLocalPath = destPath
		dst, err := os.Create(destPath)
		if err != nil {
			return nil, err
		}
		defer func() { _ = dst.Close() }()

		if _, err := io.Copy(dst, f); err != nil {
			_ = os.Remove(destPath)
			return nil, err
		}
	}

	m := &model.Media{
		UserID:       actorUserID,
		MediaType:    mediaType,
		OriginalName: origName,
		MimeType:     mimeType,
		Size:         fh.Size,
		StoragePath:  objectKey,
		CreatedBy:    actorUserID,
		UpdatedBy:    actorUserID,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		if cleanupLocalPath != "" {
			_ = os.Remove(cleanupLocalPath)
		}
		return nil, err
	}

	downloadURL, err := s.PresignGet(ctx, objectKey, 1*time.Hour)
	if err != nil {
		return nil, err
	}
	m.DownloadURL = downloadURL
	// PresignGet returns ("", nil) when MinIO isn't enabled; in that case we should not fail the upload.
	if s.minioClient != nil && s.minioBucket != "" && strings.TrimSpace(m.DownloadURL) == "" {
		return nil, errors.New("failed to presign get object")
	}

	return m, nil
}

func (s *MediaService) List(ctx context.Context, actorUserID uint, req request.MediaListRequest) (repository.CursorPage, error) {
	if actorUserID == 0 {
		return repository.CursorPage{}, ErrInvalidMediaUserID
	}
	return s.repo.List(ctx, actorUserID, req)
}

func (s *MediaService) GetByID(ctx context.Context, actorUserID, id uint) (*model.Media, error) {
	if actorUserID == 0 || id == 0 {
		return nil, ErrInvalidMedia
	}
	m, err := s.repo.FindByIDAndUserID(ctx, id, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	return m, nil
}

func (s *MediaService) UserAvatar(ctx context.Context, user *model.User) (*string, error) {
	if user == nil {
		return nil, ErrInvalidMediaUserID
	}
	return s.repo.UserAvatar(ctx, user)
}

func (s *MediaService) Delete(ctx context.Context, actorUserID, id uint) error {
	if actorUserID == 0 || id == 0 {
		return ErrInvalidMedia
	}
	// Load to know storage path
	m, err := s.repo.FindByIDAndUserID(ctx, id, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaNotFound
		}
		return err
	}

	// Delete from MinIO when configured.
	if s.minioClient != nil {
		if s.minioBucket == "" {
			return errors.New("MINIO_BUCKET is required")
		}
		err := s.minioClient.RemoveObject(ctx, s.minioBucket, m.StoragePath, minio.RemoveObjectOptions{})
		if err != nil {
			var respErr minio.ErrorResponse
			if !errors.As(err, &respErr) || respErr.StatusCode != http.StatusNotFound {
				return err
			}
		}
		return s.repo.SoftDeleteByID(ctx, id, actorUserID, actorUserID)
	}

	// Fallback local filesystem: soft delete row then remove file.
	if err := s.repo.SoftDeleteByID(ctx, id, actorUserID, actorUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaNotFound
		}
		return err
	}

	fullPath := filepath.Join(s.uploadDir, filepath.FromSlash(m.StoragePath))
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("media deleted but file removal failed: %w", err)
	}
	return nil
}
