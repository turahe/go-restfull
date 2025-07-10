package tag

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/repository"

	"github.com/google/uuid"
)

type TagApp interface {
	CreateTag(ctx context.Context, tag *model.Tag) error
	GetTagByID(ctx context.Context, id uuid.UUID) (*model.Tag, error)
	GetAllTags(ctx context.Context) ([]*model.Tag, error)
	UpdateTag(ctx context.Context, tag *model.Tag) error
	DeleteTag(ctx context.Context, id uuid.UUID) error
}

type tagApp struct {
	Repo *repository.Repository
}

func NewTagApp(repo *repository.Repository) TagApp {
	return &tagApp{
		Repo: repo,
	}
}

func (a *tagApp) CreateTag(ctx context.Context, tag *model.Tag) error {
	return a.Repo.Tag.Create(ctx, tag)
}

func (a *tagApp) GetTagByID(ctx context.Context, id uuid.UUID) (*model.Tag, error) {
	return a.Repo.Tag.GetByID(ctx, id)
}

func (a *tagApp) GetAllTags(ctx context.Context) ([]*model.Tag, error) {
	return a.Repo.Tag.GetAll(ctx)
}

func (a *tagApp) UpdateTag(ctx context.Context, tag *model.Tag) error {
	return a.Repo.Tag.Update(ctx, tag)
}

func (a *tagApp) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return a.Repo.Tag.Delete(ctx, id)
}
