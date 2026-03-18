package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/repository"

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
	posts *repository.PostRepository
}

func NewPostService(posts *repository.PostRepository) *PostService {
	return &PostService{posts: posts}
}

func (s *PostService) List(ctx context.Context, cursor *uint, limit int, dir repository.CursorDirection) (repository.CursorPage, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if dir != repository.CursorNext && dir != repository.CursorPrev {
		dir = repository.CursorNext
	}
	return s.posts.ListCursor(ctx, cursor, limit, dir)
}

func (s *PostService) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrInvalidSlug
	}
	p, err := s.posts.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PostService) Create(ctx context.Context, userID uint, title, content string) (*model.Post, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return nil, ErrInvalidPayload
	}

	base := slugify(title)
	if base == "" {
		base = "post"
	}

	slug, err := s.uniqueSlug(ctx, base)
	if err != nil {
		return nil, err
	}

	p := &model.Post{
		Title:   title,
		Slug:    slug,
		Content: content,
		UserID:  userID,
	}
	if err := s.posts.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) Update(ctx context.Context, id uint, actorUserID uint, title, content string) (*model.Post, error) {
	p, err := s.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	if p.UserID != actorUserID {
		return nil, ErrNotPostOwner
	}

	if strings.TrimSpace(title) != "" {
		p.Title = strings.TrimSpace(title)
	}
	if strings.TrimSpace(content) != "" {
		p.Content = strings.TrimSpace(content)
	}

	if err := s.posts.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	p, err := s.posts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}
	// Safer default: only author can delete.
	if p.UserID != actorUserID {
		return ErrNotPostOwner
	}
	return s.posts.DeleteByID(ctx, id)
}

func (s *PostService) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 50; i++ {
		exists, err := s.posts.SlugExists(ctx, slug)
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

