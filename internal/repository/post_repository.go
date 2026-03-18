package repository

import (
	"context"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, p *model.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PostRepository) Update(ctx context.Context, p *model.Post) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *PostRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Post{}, id).Error
}

func (r *PostRepository) SoftDeleteByID(ctx context.Context, id uint, deletedBy uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Post{}, id).Error
	})
}

func (r *PostRepository) FindByID(ctx context.Context, id uint) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) FindBySlug(ctx context.Context, slug string) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) FindBySlugWithCategories(ctx context.Context, slug string) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Where("slug = ?", slug).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) ReplaceCategories(ctx context.Context, postID uint, categories []model.Category) error {
	p := model.Post{ID: postID}
	return r.db.WithContext(ctx).Model(&p).Association("Categories").Replace(categories)
}

func (r *PostRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var id uint
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Select("id").
		Where("slug = ?", slug).
		Limit(1).
		Scan(&id).Error
	if err != nil {
		return false, err
	}
	return id != 0, nil
}

type CursorDirection string

const (
	CursorNext CursorDirection = "next"
	CursorPrev CursorDirection = "prev"
)

type CursorPage struct {
	Items      []model.Post
	NextCursor *uint
	PrevCursor *uint
}

func (r *PostRepository) ListCursor(ctx context.Context, cursor *uint, limit int, dir CursorDirection) (CursorPage, error) {
	q := r.db.WithContext(ctx).Model(&model.Post{})

	if cursor != nil {
		if dir == CursorPrev {
			q = q.Where("id < ?", *cursor).Order("id desc")
		} else {
			q = q.Where("id > ?", *cursor).Order("id asc")
		}
	} else {
		if dir == CursorPrev {
			q = q.Order("id desc")
		} else {
			q = q.Order("id asc")
		}
	}

	var rows []model.Post
	err := q.Limit(limit + 1).Find(&rows).Error
	if err != nil {
		return CursorPage{}, err
	}

	hasMoreInDir := len(rows) > limit
	if hasMoreInDir {
		rows = rows[:limit]
	}

	// If we fetched in desc order for prev, reverse to keep stable asc output.
	if dir == CursorPrev {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}

	var nextCursor *uint
	var prevCursor *uint

	if len(rows) == 0 {
		return CursorPage{Items: rows, NextCursor: nil, PrevCursor: nil}, nil
	}

	firstID := rows[0].ID
	lastID := rows[len(rows)-1].ID

	// Check existence without COUNT(*).
	if hasMoreInDir {
		if dir == CursorNext {
			tmp := lastID
			nextCursor = &tmp
		} else {
			tmp := firstID
			prevCursor = &tmp
		}
	}

	// Opposite direction cursors (cheap 1-row existence checks).
	if prevCursor == nil {
		var tmpID uint
		err = r.db.WithContext(ctx).Model(&model.Post{}).Select("id").Where("id < ?", firstID).Order("id desc").Limit(1).Scan(&tmpID).Error
		if err != nil {
			return CursorPage{}, err
		}
		if tmpID != 0 {
			tmp := firstID
			prevCursor = &tmp
		}
	}
	if nextCursor == nil {
		var tmpID uint
		err = r.db.WithContext(ctx).Model(&model.Post{}).Select("id").Where("id > ?", lastID).Order("id asc").Limit(1).Scan(&tmpID).Error
		if err != nil {
			return CursorPage{}, err
		}
		if tmpID != 0 {
			tmp := lastID
			nextCursor = &tmp
		}
	}

	return CursorPage{Items: rows, NextCursor: nextCursor, PrevCursor: prevCursor}, nil
}

