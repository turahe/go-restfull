package service

import (
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/config"
	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/objectstore"
	"github.com/turahe/go-restfull/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrInvalidMedia       = errors.New("invalid media")
	ErrMediaTooLarge      = errors.New("media too large")
	ErrMediaNotFound      = errors.New("media not found")
	ErrInvalidMediaUserID = errors.New("invalid media user id")
	ErrInvalidActorID     = errors.New("invalid actor id")
)

type MediaService struct {
	repo     *repository.MediaRepository
	maxBytes int64

	store objectstore.Store
	log   *zap.Logger
}

func NewMediaService(repo *repository.MediaRepository, cfg config.Config, log *zap.Logger) (*MediaService, error) {
	store, err := objectstore.NewFromConfig(cfg, log)
	if err != nil {
		return nil, err
	}
	return &MediaService{
		repo:     repo,
		maxBytes: cfg.MediaMaxUploadBytes,
		store:    store,
		log:      log,
	}, nil
}

// PresignGet returns a temporary URL for downloading the object.
func (s *MediaService) PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if objectKey == "" {
		return "", nil
	}
	u, err := s.store.SignedURL(ctx, objectKey, expiry)
	if err != nil {
		s.log.Error("failed to presign get object", zap.Error(err))
		return "", errors.New("failed to presign get object")
	}
	if u == "" {
		s.log.Error("failed to presign get object")
		return "", errors.New("failed to presign get object")
	}
	return u, nil
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

	if err := s.store.Put(ctx, objectKey, f, fh.Size, mimeType); err != nil {
		return nil, err
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
		_ = s.store.Delete(ctx, objectKey)
		return nil, err
	}

	downloadURL, err := s.PresignGet(ctx, objectKey, 1*time.Hour)
	if err != nil {
		return nil, err
	}
	m.DownloadURL = downloadURL
	if strings.TrimSpace(m.DownloadURL) == "" {
		return nil, errors.New("failed to presign get object")
	}

	return m, nil
}

func (s *MediaService) List(ctx context.Context, actorUserID uint, req request.MediaListRequest) (repository.CursorPage, error) {
	if actorUserID == 0 {
		return repository.CursorPage{}, ErrInvalidMediaUserID
	}
	page, err := s.repo.List(ctx, actorUserID, req)
	if err != nil {
		return repository.CursorPage{}, err
	}

	items, ok := page.Items.([]model.Media)
	if !ok || len(items) == 0 {
		return page, nil
	}

	expiry := 15 * time.Minute
	for i := range items {
		if items[i].StoragePath == "" {
			continue
		}
		url, uerr := s.PresignGet(ctx, items[i].StoragePath, expiry)
		if uerr != nil || url == "" {
			continue
		}
		items[i].DownloadURL = url
	}
	page.Items = items
	return page, nil
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
	downloadURL, err := s.PresignGet(ctx, m.StoragePath, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	m.DownloadURL = downloadURL
	if m.DownloadURL == "" {
		return nil, errors.New("failed to presign get object")
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
	m, err := s.repo.FindByIDAndUserID(ctx, id, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaNotFound
		}
		return err
	}

	if err := s.store.Delete(ctx, m.StoragePath); err != nil {
		return err
	}
	return s.repo.SoftDeleteByID(ctx, id, actorUserID, actorUserID)
}
