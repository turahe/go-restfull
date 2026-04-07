package usecase

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/turahe/go-restfull/pkg/ids"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrMediaInvalid          = errors.New("invalid media")
	ErrMediaTooLarge         = errors.New("media too large")
	ErrMediaNotFound         = errors.New("media not found")
	ErrMediaInvalidUserID    = errors.New("invalid media user id")
	ErrMediaInvalidActor     = errors.New("invalid actor id")
	ErrMediaDuplicateName    = errors.New("media name already exists under this parent")
	ErrMediaParentNotFound   = errors.New("parent media not found")
)

// MediaTreeNode is the JSON shape for tree responses (id, name, children).
type MediaTreeNode struct {
	ID       uint            `json:"id"`
	Name     string          `json:"name"`
	Children []MediaTreeNode `json:"children"`
}

type MediaUsecase struct {
	repo     *repository.MediaRepository
	maxBytes int64

	store objectstore.Store
	log   *zap.Logger
}

func NewMediaUsecase(repo *repository.MediaRepository, cfg config.Config, log *zap.Logger) (*MediaUsecase, error) {
	store, err := objectstore.NewFromConfig(cfg, log)
	if err != nil {
		return nil, err
	}
	return &MediaUsecase{
		repo:     repo,
		maxBytes: cfg.MediaMaxUploadBytes,
		store:    store,
		log:      log,
	}, nil
}

func (u *MediaUsecase) PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if objectKey == "" {
		return "", nil
	}
	url, err := u.store.SignedURL(ctx, objectKey, expiry)
	if err != nil {
		u.log.Error("failed to presign get object", zap.Error(err))
		return "", errors.New("failed to presign get object")
	}
	if url == "" {
		u.log.Error("failed to presign get object")
		return "", errors.New("failed to presign get object")
	}
	return url, nil
}

// Upload uploads the file to storage and creates a Media row (optional parent folder id).
func (u *MediaUsecase) Upload(
	ctx context.Context,
	actorUserID uint,
	fh *multipart.FileHeader,
	parentID *uint,
) (*model.Media, error) {
	if actorUserID == 0 {
		return nil, ErrMediaInvalidActor
	}
	if fh == nil {
		return nil, ErrMediaInvalid
	}
	if fh.Size > u.maxBytes {
		return nil, ErrMediaTooLarge
	}

	origName := strings.TrimSpace(fh.Filename)
	if origName == "" {
		origName = "upload"
	}
	ext := strings.ToLower(filepath.Ext(origName))
	if ext == "" {
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

	id, err := ids.New()
	if err != nil {
		u.log.Error("failed to generate id", zap.Error(err))
		return nil, err
	}

	relDir := time.Now().Format("2006/01/02")
	storageFilename := id + ext
	objectKey := filepath.ToSlash(filepath.Join(relDir, storageFilename))

	if err := u.store.Put(ctx, objectKey, f, fh.Size, mimeType); err != nil {
		return nil, err
	}

	baseName := filepath.Base(origName)
	var m *model.Media
	for attempt := 0; attempt < 10; attempt++ {
		name := baseName
		if attempt > 0 {
			name = fmt.Sprintf("%s_%d", filepath.Base(origName), attempt)
		}

		m = &model.Media{
			UserID:       actorUserID,
			Name:         name,
			MediaType:    mediaType,
			OriginalName: origName,
			MimeType:     mimeType,
			Size:         fh.Size,
			StoragePath:  objectKey,
			CreatedBy:    actorUserID,
			UpdatedBy:    actorUserID,
		}

		var insErr error
		if parentID == nil || *parentID == 0 {
			insErr = u.repo.CreateFileRoot(ctx, m)
		} else {
			insErr = u.repo.CreateFileChild(ctx, actorUserID, *parentID, m)
		}
		if insErr == nil {
			goto done
		}
		if errors.Is(insErr, gorm.ErrDuplicatedKey) {
			continue
		}
		if errors.Is(insErr, gorm.ErrRecordNotFound) {
			_ = u.store.Delete(ctx, objectKey)
			return nil, ErrMediaParentNotFound
		}
		_ = u.store.Delete(ctx, objectKey)
		return nil, insErr
	}
	_ = u.store.Delete(ctx, objectKey)
	return nil, errors.New("could not allocate unique media name")

done:
	downloadURL, err := u.PresignGet(ctx, objectKey, 1*time.Hour)
	if err != nil {
		return nil, err
	}
	m.DownloadURL = downloadURL
	if strings.TrimSpace(m.DownloadURL) == "" {
		return nil, errors.New("failed to presign get object")
	}

	return m, nil
}

func (u *MediaUsecase) CreateFolderRoot(ctx context.Context, actorUserID uint, name string) (*model.Media, error) {
	if actorUserID == 0 {
		return nil, ErrMediaInvalidActor
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrMediaInvalid
	}
	out, err := u.repo.CreateFolderRoot(ctx, actorUserID, name, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrMediaDuplicateName
		}
		return nil, err
	}
	return out, nil
}

func (u *MediaUsecase) CreateFolderChild(ctx context.Context, actorUserID uint, parentID uint, name string) (*model.Media, error) {
	if actorUserID == 0 || parentID == 0 {
		return nil, ErrMediaInvalid
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrMediaInvalid
	}
	out, err := u.repo.CreateFolderChild(ctx, actorUserID, parentID, name, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaParentNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrMediaDuplicateName
		}
		return nil, err
	}
	return out, nil
}

func (u *MediaUsecase) GetTree(ctx context.Context, actorUserID uint) ([]MediaTreeNode, error) {
	if actorUserID == 0 {
		return nil, ErrMediaInvalidUserID
	}
	rows, err := u.repo.GetTree(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	return buildMediaTree(rows), nil
}

func (u *MediaUsecase) GetSubtree(ctx context.Context, actorUserID uint, mediaID uint) ([]MediaTreeNode, error) {
	if actorUserID == 0 || mediaID == 0 {
		return nil, ErrMediaNotFound
	}
	rows, err := u.repo.GetSubtree(ctx, actorUserID, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrMediaNotFound
	}
	return buildMediaTree(rows), nil
}

func (u *MediaUsecase) Update(ctx context.Context, actorUserID uint, id uint, name string) (*model.Media, error) {
	if actorUserID == 0 || id == 0 {
		return nil, ErrMediaInvalid
	}
	m, err := u.repo.UpdateName(ctx, actorUserID, id, name, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrMediaDuplicateName
		}
		return nil, err
	}
	return m, nil
}

func buildMediaTree(rows []model.Media) []MediaTreeNode {
	type stackItem struct {
		node *MediaTreeNode
		rgt  int
	}
	var stack []stackItem
	var roots []MediaTreeNode
	for _, row := range rows {
		n := MediaTreeNode{ID: row.ID, Name: row.Name, Children: []MediaTreeNode{}}
		for len(stack) > 0 && stack[len(stack)-1].rgt < row.Lft {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, n)
			stack = append(stack, stackItem{node: &roots[len(roots)-1], rgt: row.Rgt})
			continue
		}
		p := stack[len(stack)-1].node
		p.Children = append(p.Children, n)
		stack = append(stack, stackItem{node: &p.Children[len(p.Children)-1], rgt: row.Rgt})
	}
	return roots
}

func (u *MediaUsecase) List(ctx context.Context, actorUserID uint, req request.MediaListRequest) (repository.CursorPage, error) {
	if actorUserID == 0 {
		return repository.CursorPage{}, ErrMediaInvalidUserID
	}
	page, err := u.repo.List(ctx, actorUserID, req)
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
		url, uerr := u.PresignGet(ctx, items[i].StoragePath, expiry)
		if uerr != nil || url == "" {
			continue
		}
		items[i].DownloadURL = url
	}
	page.Items = items
	return page, nil
}

func (u *MediaUsecase) GetByID(ctx context.Context, actorUserID, id uint) (*model.Media, error) {
	if actorUserID == 0 || id == 0 {
		return nil, ErrMediaInvalid
	}
	m, err := u.repo.FindByIDAndUserID(ctx, id, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(m.StoragePath) == "" {
		m.DownloadURL = ""
		return m, nil
	}
	downloadURL, err := u.PresignGet(ctx, m.StoragePath, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	m.DownloadURL = downloadURL
	if m.DownloadURL == "" {
		return nil, errors.New("failed to presign get object")
	}
	return m, nil
}

func (u *MediaUsecase) UserAvatar(ctx context.Context, user *model.User) (*string, error) {
	if user == nil {
		return nil, ErrMediaInvalidUserID
	}
	return u.repo.UserAvatar(ctx, user)
}

func (u *MediaUsecase) Delete(ctx context.Context, actorUserID, id uint) error {
	if actorUserID == 0 || id == 0 {
		return ErrMediaInvalid
	}
	rows, err := u.repo.GetSubtree(ctx, actorUserID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaNotFound
		}
		return err
	}
	if len(rows) == 0 {
		return ErrMediaNotFound
	}
	for i := range rows {
		if rows[i].MediaType == "folder" || rows[i].StoragePath == "" {
			continue
		}
		if err := u.store.Delete(ctx, rows[i].StoragePath); err != nil {
			return err
		}
	}
	return u.repo.DeleteSubtree(ctx, actorUserID, id, actorUserID)
}
