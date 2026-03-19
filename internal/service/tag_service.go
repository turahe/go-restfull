package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"
	"go-rest/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrTagNotFound   = errors.New("tag not found")
	ErrInvalidTagID  = errors.New("invalid tag id")
	ErrInvalidTagReq = errors.New("invalid tag payload")
)

type TagService struct {
	tags *repository.TagRepository
	log  *zap.Logger
}

func NewTagService(tags *repository.TagRepository, log *zap.Logger) *TagService {
	return &TagService{tags: tags, log: log}
}

func (s *TagService) List(ctx context.Context, req request.TagListRequest) (repository.CursorPage, error) {
	page, err := s.tags.List(ctx, req)
	if err != nil {
		s.log.Error("failed to list tags", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}

func (s *TagService) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		s.log.Error("invalid slug")
		return nil, ErrInvalidSlug
	}
	t, err := s.tags.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("tag not found")
			return nil, ErrTagNotFound
		}
		s.log.Error("failed to find tag by slug", zap.Error(err))
		return nil, err
	}
	return t, nil
}

func (s *TagService) Create(ctx context.Context, actorUserID uint, req request.CreateTagRequest) (*model.Tag, error) {
	_ = actorUserID // reserved for future tagstamps; Tag currently has no CreatedBy/UpdatedBy

	base := slugify(req.Name)
	if base == "" {
		base = "tag"
	}
	slug, err := s.uniqueSlug(ctx, base)
	if err != nil {
		s.log.Error("failed to generate unique slug", zap.Error(err))
		return nil, err
	}

	t := &model.Tag{Name: req.Name, Slug: slug}
	if err := s.tags.Create(ctx, t); err != nil {
		s.log.Error("failed to create tag", zap.Error(err))
		return nil, err
	}
	return t, nil
}

func (s *TagService) Update(ctx context.Context, id uint, actorUserID uint, req request.UpdateTagRequest) (*model.Tag, error) {
	_ = actorUserID
	if id == 0 {
		s.log.Error("invalid tag id")
		return nil, ErrInvalidTagID
	}
	t, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("tag not found")
			return nil, ErrTagNotFound
		}
		s.log.Error("failed to find tag by id", zap.Error(err))
		return nil, err
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if err := s.tags.Update(ctx, t); err != nil {
		s.log.Error("failed to update tag", zap.Error(err))
		return nil, err
	}
	return t, nil
}

func (s *TagService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	_ = actorUserID
	if id == 0 {
		s.log.Error("invalid tag id")
		return ErrInvalidTagID
	}
	_, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("tag not found")
			return ErrTagNotFound
		}
		s.log.Error("failed to find tag by id", zap.Error(err))
		return err
	}
	return s.tags.DeleteByID(ctx, id)
}

func (s *TagService) FindByIDs(ctx context.Context, ids []uint) ([]model.Tag, error) {
	tags, err := s.tags.FindByIDs(ctx, ids)
	if err != nil {
		s.log.Error("failed to find tags by ids", zap.Error(err))
		return nil, err
	}
	return tags, nil
}

func (s *TagService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 50; i++ {
		exists, err := s.tags.SlugExists(ctx, slug)
		if err != nil {
			s.log.Error("failed to check if slug exists", zap.Error(err))
			return "", err
		}
		if !exists {
			s.log.Error("slug already exists")
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i+1)
	}
	s.log.Error("could not generate unique slug")
	return "", errors.New("could not generate unique slug")
}
