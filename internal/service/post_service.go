package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrPostNotFound   = errors.New("post not found")
	ErrNotPostOwner   = errors.New("not the post owner")
	ErrInvalidPostID  = errors.New("invalid post id")
	ErrInvalidSlug    = errors.New("invalid slug")
	ErrInvalidPayload = errors.New("invalid payload")
)

type PostService struct {
	posts      *repository.PostRepository
	categories *repository.CategoryRepository
	tags       *repository.TagRepository
	log        *zap.Logger
}

func NewPostService(posts *repository.PostRepository, categories *repository.CategoryRepository, tags *repository.TagRepository, log *zap.Logger) *PostService {
	return &PostService{posts: posts, categories: categories, tags: tags, log: log}
}

func (s *PostService) List(ctx context.Context, req request.PostListRequest) (repository.CursorPage, error) {
	page, err := s.posts.ListCursor(ctx, req)
	if err != nil {
		s.log.Error("failed to list posts", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}

func (s *PostService) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		s.log.Error("invalid slug")
		return nil, ErrInvalidSlug
	}
	p, err := s.posts.FindBySlugWithCategory(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PostService) Create(ctx context.Context, userID uint, req request.CreatePostRequest) (*model.Post, error) {
	base := slugify(req.Title)
	if base == "" {
		base = "post"
	}

	slug, err := s.uniqueSlug(ctx, base)
	if err != nil {
		s.log.Error("failed to generate unique slug", zap.Error(err))
		return nil, err
	}

	cats, err := s.categories.FindByIDs(ctx, []uint{req.CategoryID})
	if err != nil {
		s.log.Error("failed to find categories by ids", zap.Error(err))
		return nil, err
	}
	if len(cats) != 1 {
		s.log.Error("category not found")
		return nil, errors.New("category not found")
	}

	p := &model.Post{
		Title:      req.Title,
		Slug:       slug,
		Content:    req.Content,
		UserID:     userID,
		CategoryID: req.CategoryID,
		CreatedBy:  userID,
		UpdatedBy:  userID,
	}
	if err := s.posts.Create(ctx, p); err != nil {
		s.log.Error("failed to create post", zap.Error(err))
		return nil, err
	}

	if len(req.TagIDs) > 0 && s.tags != nil {
		tags, err := s.tags.FindByIDs(ctx, UniqueUint(req.TagIDs))
		if err != nil {
			s.log.Error("failed to find tags by ids", zap.Error(err))
			return nil, err
		}
		if len(tags) != len(UniqueUint(req.TagIDs)) {
			s.log.Error("one or more tags not found")
			return nil, errors.New("one or more tags not found")
		}
		if err := s.posts.ReplaceTags(ctx, p.ID, tags); err != nil {
			s.log.Error("failed to replace tags", zap.Error(err))
			return nil, err
		}
		p.Tags = tags
	}

	return p, nil
}

func (s *PostService) Update(ctx context.Context, id uint, actorUserID uint, req request.UpdatePostRequest) (*model.Post, error) {
	p, err := s.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	if p.UserID != actorUserID {
		s.log.Error("not the post owner")
		return nil, ErrNotPostOwner
	}

	if req.Title != "" {
		p.Title = strings.TrimSpace(req.Title)
	}
	if req.Content != "" {
		p.Content = req.Content
	}
	if req.CategoryID != nil {
		if *req.CategoryID == 0 {
			s.log.Error("invalid payload")
			return nil, ErrInvalidPayload
		}
		cats, err := s.categories.FindByIDs(ctx, []uint{*req.CategoryID})
		if err != nil {
			s.log.Error("failed to find categories by ids", zap.Error(err))
			return nil, err
		}
		if len(cats) != 1 {
			s.log.Error("category not found")
			return nil, errors.New("category not found")
		}
		p.CategoryID = *req.CategoryID
	}
	p.UpdatedBy = actorUserID

	if err := s.posts.Update(ctx, p); err != nil {
		s.log.Error("failed to update post", zap.Error(err))
		return nil, err
	}

	// If tagIds is present in JSON, gin will bind it as either [] (empty) or [..].
	// If not present, it stays nil. That lets us distinguish "no change" vs "replace/clear".
	if req.TagIDs != nil && s.tags != nil {
		tags, err := s.tags.FindByIDs(ctx, UniqueUint(req.TagIDs))
		if err != nil {
			s.log.Error("failed to find tags by ids", zap.Error(err))
			return nil, err
		}
		if len(tags) != len(UniqueUint(req.TagIDs)) {
			s.log.Error("one or more tags not found")
			return nil, errors.New("one or more tags not found")
		}
		if err := s.posts.ReplaceTags(ctx, p.ID, tags); err != nil {
			s.log.Error("failed to replace tags", zap.Error(err))
			return nil, err
		}
		p.Tags = tags
	}
	return p, nil
}

func (s *PostService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	p, err := s.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("post not found")
			return ErrPostNotFound
		}
		s.log.Error("failed to find post by id", zap.Error(err))
		return err
	}
	// Safer default: only author can delete.
	if p.UserID != actorUserID {
		s.log.Error("not the post owner")
		return ErrNotPostOwner
	}
	if err := s.posts.SoftDeleteByID(ctx, id, actorUserID); err != nil {
		s.log.Error("failed to soft delete post by id", zap.Error(err))
		return err
	}
	return nil
}

func (s *PostService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 50; i++ {
		exists, err := s.posts.SlugExists(ctx, slug)
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

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}
