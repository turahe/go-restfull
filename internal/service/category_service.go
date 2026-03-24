package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound   = errors.New("category not found")
	ErrInvalidCategoryID  = errors.New("invalid category id")
	ErrInvalidCategoryReq = errors.New("invalid category payload")
)

type CategoryService struct {
	categories *repository.CategoryRepository
	log        *zap.Logger
}

func NewCategoryService(categories *repository.CategoryRepository, log *zap.Logger) *CategoryService {
	return &CategoryService{categories: categories, log: log}
}

func (s *CategoryService) List(ctx context.Context, req request.CategoryListRequest) (repository.CursorPage, error) {
	page, err := s.categories.List(ctx, req)
	if err != nil {
		s.log.Error("failed to list categories", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}

func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		s.log.Error("invalid slug")
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

func (s *CategoryService) Create(ctx context.Context, actorUserID uint, req request.CreateCategoryRequest) (*model.Category, error) {
	base := slugify(req.Name)
	if base == "" {
		base = "category"
	}

	slug, err := s.uniqueSlug(ctx, base)
	if err != nil {
		s.log.Error("failed to generate unique slug", zap.Error(err))
		return nil, err
	}

	c := &model.Category{
		Name:      req.Name,
		Slug:      slug,
		CreatedBy: actorUserID,
		UpdatedBy: actorUserID,
	}
	if err := s.categories.Create(ctx, c); err != nil {
		s.log.Error("failed to create category", zap.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Update(ctx context.Context, id uint, actorUserID uint, req request.UpdateCategoryRequest) (*model.Category, error) {
	if id == 0 {
		s.log.Error("invalid category id")
		return nil, ErrInvalidCategoryID
	}
	c, err := s.categories.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	if req.Name != "" {
		c.Name = req.Name
	}
	c.UpdatedBy = actorUserID

	if err := s.categories.Update(ctx, c); err != nil {
		s.log.Error("failed to update category", zap.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	if id == 0 {
		s.log.Error("invalid category id")
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
	cats, err := s.categories.FindByIDs(ctx, ids)
	if err != nil {
		s.log.Error("failed to find categories by ids", zap.Error(err))
		return nil, err
	}
	return cats, nil
}

func (s *CategoryService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 50; i++ {
		exists, err := s.categories.SlugExists(ctx, slug)
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
