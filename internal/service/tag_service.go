package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrTagNotFound   = errors.New("tag not found")
	ErrInvalidTagID  = errors.New("invalid tag id")
	ErrInvalidTagReq = errors.New("invalid tag payload")
)

type TagService struct {
	tags *repository.TagRepository
}

func NewTagService(tags *repository.TagRepository) *TagService {
	return &TagService{tags: tags}
}

func (s *TagService) List(ctx context.Context, limit int) ([]model.Tag, error) {
	return s.tags.List(ctx, limit)
}

func (s *TagService) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrInvalidSlug
	}
	t, err := s.tags.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *TagService) Create(ctx context.Context, actorUserID uint, name string) (*model.Tag, error) {
	_ = actorUserID // reserved for future tagstamps; Tag currently has no CreatedBy/UpdatedBy

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTagReq
	}

	base := slugify(name)
	if base == "" {
		base = "tag"
	}
	slug, err := s.uniqueSlug(ctx, base)
	if err != nil {
		return nil, err
	}

	t := &model.Tag{Name: name, Slug: slug}
	if err := s.tags.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TagService) Update(ctx context.Context, id uint, actorUserID uint, name string) (*model.Tag, error) {
	_ = actorUserID
	if id == 0 {
		return nil, ErrInvalidTagID
	}
	t, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name != "" {
		t.Name = name
	}
	if err := s.tags.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TagService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	_ = actorUserID
	if id == 0 {
		return ErrInvalidTagID
	}
	_, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	return s.tags.DeleteByID(ctx, id)
}

func (s *TagService) FindByIDs(ctx context.Context, ids []uint) ([]model.Tag, error) {
	return s.tags.FindByIDs(ctx, ids)
}

func (s *TagService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 50; i++ {
		exists, err := s.tags.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i+1)
	}
	return "", errors.New("could not generate unique slug")
}

