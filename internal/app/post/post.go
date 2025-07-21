package post

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/http/requests"
	"webapi/internal/repository"

	"github.com/google/uuid"
)

type PostApp interface {
	GetPostByIDWithContents(ctx context.Context, id uuid.UUID) (*model.Post, error)
	GetAllPostsWithContents(ctx context.Context) ([]*model.Post, error)
	CreatePost(ctx context.Context, post *model.Post) error
	UpdatePost(ctx context.Context, post *model.Post) error
	DeletePost(ctx context.Context, id uuid.UUID) error
	CreatePostWithTags(ctx context.Context, post *model.Post, tagIDs []uuid.UUID) error
	UpdatePostWithTags(ctx context.Context, post *model.Post, tagIDs []uuid.UUID) error
	GetPostsWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) ([]*model.Post, int, error)
}

type postApp struct {
	repo *repository.Repository
}

func NewPostApp(repo *repository.Repository) PostApp {
	return &postApp{repo: repo}
}

func (a *postApp) GetPostByIDWithContents(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	return a.repo.Post.GetByIDWithContents(ctx, id)
}

func (a *postApp) GetAllPostsWithContents(ctx context.Context) ([]*model.Post, error) {
	return a.repo.Post.GetAllWithContents(ctx)
}

func (a *postApp) CreatePost(ctx context.Context, post *model.Post) error {
	return a.repo.Post.Create(ctx, post)
}

func (a *postApp) UpdatePost(ctx context.Context, post *model.Post) error {
	return a.repo.Post.Update(ctx, post)
}

func (a *postApp) DeletePost(ctx context.Context, id uuid.UUID) error {
	return a.repo.Post.Delete(ctx, id)
}

func (a *postApp) CreatePostWithTags(ctx context.Context, post *model.Post, tagIDs []uuid.UUID) error {
	if err := a.repo.Post.Create(ctx, post); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if err := a.repo.Tag.AttachTag(ctx, tagID, post.ID, "post"); err != nil {
			return err
		}
	}
	return nil
}

func (a *postApp) UpdatePostWithTags(ctx context.Context, post *model.Post, tagIDs []uuid.UUID) error {
	if err := a.repo.Post.Update(ctx, post); err != nil {
		return err
	}
	// Remove all existing tags for this post
	// (Assume a method to remove all tags for a post, or detach individually)
	// For now, detach all then re-attach
	// Get current tags
	tags, err := a.repo.Tag.GetTagsForEntity(ctx, post.ID, "post")
	if err != nil {
		return err
	}
	for _, tag := range tags {
		tagUUID, err := uuid.Parse(tag.ID)
		if err == nil {
			a.repo.Tag.DetachTag(ctx, tagUUID, post.ID, "post")
		}
	}
	for _, tagID := range tagIDs {
		if err := a.repo.Tag.AttachTag(ctx, tagID, post.ID, "post"); err != nil {
			return err
		}
	}
	return nil
}

func (a *postApp) GetPostsWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) ([]*model.Post, int, error) {
	return a.repo.Post.GetPostsWithPagination(ctx, input)
}
