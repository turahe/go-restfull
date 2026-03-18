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
	ErrCategoryNotFound   = errors.New("category not found")
	ErrInvalidCategoryID  = errors.New("invalid category id")
	ErrInvalidCategoryReq = errors.New("invalid category payload")
)

type CategoryService struct {
	categories *repository.CategoryRepository
}

func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

func (s *CategoryService) List(ctx context.Context, limit int) ([]model.Category, error) {
	return s.categories.List(ctx, limit)
}

func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrInvalidSlug
	}
	c, err := s.categories.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Create(ctx context.Context, actorUserID uint, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidCategoryReq
	}

	base := slugify(name)
	if base == "" {
		base = "category"
	}

	slug, err := s.uniqueSlug(ctx, base)
	if err != nil {
		return nil, err
	}

	c := &model.Category{
		Name:      name,
		Slug:      slug,
		CreatedBy: actorUserID,
		UpdatedBy: actorUserID,
	}
	if err := s.categories.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Update(ctx context.Context, id uint, actorUserID uint, name string) (*model.Category, error) {
	if id == 0 {
		return nil, ErrInvalidCategoryID
	}
	c, err := s.categories.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name != "" {
		c.Name = name
	}
	c.UpdatedBy = actorUserID

	if err := s.categories.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	if id == 0 {
		return ErrInvalidCategoryID
	}
	_, err := s.categories.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.categories.SoftDeleteByID(ctx, id, actorUserID)
}

func (s *CategoryService) FindByIDs(ctx context.Context, ids []uint) ([]model.Category, error) {
	return s.categories.FindByIDs(ctx, ids)
}

func (s *CategoryService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 50; i++ {
		exists, err := s.categories.SlugExists(ctx, slug)
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

