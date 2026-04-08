package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var categorySlugNonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

func categorySlugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = categorySlugNonAlpha.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

// nextCategorySlug returns a unique slug for categories (global uniqueness), excluding excludeID when > 0.
func nextCategorySlug(db *gorm.DB, name string, excludeID uint) (string, error) {
	base := categorySlugify(name)
	if base == "" {
		base = "category"
	}
	slug := base
	for i := 0; i < 100; i++ {
		var n int64
		q := db.Model(&model.CategoryModel{}).Where("slug = ?", slug)
		if excludeID > 0 {
			q = q.Where("id <> ?", excludeID)
		}
		if err := q.Count(&n).Error; err != nil {
			return "", err
		}
		if n == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i+1)
	}
	return "", errors.New("could not allocate unique category slug")
}

// ErrCategorySubtreeHasPosts is returned when delete would remove categories still referenced by posts.category_id.
var ErrCategorySubtreeHasPosts = errors.New("posts reference this category or its descendants")

// MySQL advisory lock name: serializes nested-set mutations to prevent concurrent insert/shift races.
const categoryNestedSetLockName = "go_restfull_categories_nested_set"

type CategoryRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewCategoryRepository(db *gorm.DB, log *zap.Logger) *CategoryRepository {
	return &CategoryRepository{db: db, log: log}
}

func lockNestedSet(tx *gorm.DB) error {
	if tx.Dialector.Name() != "mysql" {
		return nil
	}
	return tx.Exec("SELECT GET_LOCK(?, -1)", categoryNestedSetLockName).Error
}

func unlockNestedSet(tx *gorm.DB) {
	if tx.Dialector.Name() != "mysql" {
		return
	}
	_ = tx.Exec("SELECT RELEASE_LOCK(?)", categoryNestedSetLockName).Error
}

// CreateRoot inserts a root category (depth 0). First root uses lft=1, rgt=2; further roots append after max(rgt).
func (r *CategoryRepository) CreateRoot(ctx context.Context, name string, actorUserID uint) (*model.CategoryModel, error) {
	var out *model.CategoryModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockNestedSet(tx); err != nil {
			return err
		}
		defer unlockNestedSet(tx)

		var cnt int64
		if err := tx.Model(&model.CategoryModel{}).Where("parent_id IS NULL AND name = ?", name).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return gorm.ErrDuplicatedKey
		}

		var maxRgt int
		q := tx.Model(&model.CategoryModel{})
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.Select("COALESCE(MAX(rgt), 0)").Scan(&maxRgt).Error; err != nil {
			return err
		}
		var n int64
		if err := tx.Model(&model.CategoryModel{}).Count(&n).Error; err != nil {
			return err
		}
		var lft, rgt int
		if n == 0 {
			lft, rgt = 1, 2
		} else {
			lft = maxRgt + 1
			rgt = maxRgt + 2
		}
		slug, err := nextCategorySlug(tx, name, 0)
		if err != nil {
			return err
		}
		out = &model.CategoryModel{
			Name:      name,
			Slug:      slug,
			Lft:       lft,
			Rgt:       rgt,
			Depth:     0,
			CreatedBy: actorUserID,
			UpdatedBy: actorUserID,
		}
		return tx.Create(out).Error
	})
	if err != nil {
		r.log.Error("create root category failed", zap.Error(err))
		return nil, err
	}
	return out, nil
}

// CreateChild inserts a category as the last child of parentID using nested-set shift + insert (single transaction).
func (r *CategoryRepository) CreateChild(ctx context.Context, parentID uint, name string, actorUserID uint) (*model.CategoryModel, error) {
	var out *model.CategoryModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockNestedSet(tx); err != nil {
			return err
		}
		defer unlockNestedSet(tx)

		var dup int64
		if err := tx.Model(&model.CategoryModel{}).Where("parent_id = ? AND name = ?", parentID, name).Count(&dup).Error; err != nil {
			return err
		}
		if dup > 0 {
			return gorm.ErrDuplicatedKey
		}

		var parent model.CategoryModel
		q := tx
		if tx.Dialector.Name() == "mysql" {
			q = tx.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&parent, parentID).Error; err != nil {
			return err
		}
		parentRgt := parent.Rgt
		if err := tx.Exec("UPDATE categories SET rgt = rgt + 2 WHERE deleted_at IS NULL AND rgt >= ?", parentRgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE categories SET lft = lft + 2 WHERE deleted_at IS NULL AND lft > ?", parentRgt).Error; err != nil {
			return err
		}
		slug, err := nextCategorySlug(tx, name, 0)
		if err != nil {
			return err
		}
		pid := parentID
		out = &model.CategoryModel{
			Name:      name,
			Slug:      slug,
			ParentID:  &pid,
			Lft:       parentRgt,
			Rgt:       parentRgt + 1,
			Depth:     parent.Depth + 1,
			CreatedBy: actorUserID,
			UpdatedBy: actorUserID,
		}
		return tx.Create(out).Error
	})
	if err != nil {
		r.log.Error("create child category failed", zap.Error(err))
		return nil, err
	}
	return out, nil
}

// List returns a paginated flat list of categories ordered by lft (tree order).
func (r *CategoryRepository) List(ctx context.Context, req request.CategoryListRequest) (CursorPage, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	countQ := r.db.WithContext(ctx).Model(&model.CategoryModel{})
	if req.Name != "" {
		countQ = countQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if s := strings.TrimSpace(req.Search); s != "" {
		countQ = countQ.Where("name LIKE ?", "%"+s+"%")
	}

	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count categories", zap.Error(err))
		return CursorPage{}, err
	}

	q := r.db.WithContext(ctx).Model(&model.CategoryModel{})
	if req.Name != "" {
		q = q.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if s := strings.TrimSpace(req.Search); s != "" {
		q = q.Where("name LIKE ?", "%"+s+"%")
	}

	var rows []model.CategoryModel
	if err := q.Order("lft asc").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		r.log.Error("failed to list categories", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.CategoryModel{}, NextCursor: nil, PrevCursor: nil}, nil
	}

	var nextCursor *uint
	if int64(offset)+int64(limit) < totalRows {
		tmp := rows[len(rows)-1].ID
		nextCursor = &tmp
	}

	var prevCursor *uint
	if offset > 0 {
		tmp := rows[0].ID
		prevCursor = &tmp
	}

	return CursorPage{Items: rows, NextCursor: nextCursor, PrevCursor: prevCursor}, nil
}

// GetTree returns all categories ordered by lft ascending (single query).
func (r *CategoryRepository) GetTree(ctx context.Context) ([]model.CategoryModel, error) {
	var rows []model.CategoryModel
	if err := r.db.WithContext(ctx).Order("lft ASC").Find(&rows).Error; err != nil {
		r.log.Error("get category tree failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// GetSubtree returns categories in the subtree rooted at categoryID, ordered by lft (single query).
func (r *CategoryRepository) GetSubtree(ctx context.Context, categoryID uint) ([]model.CategoryModel, error) {
	var anchor model.CategoryModel
	if err := r.db.WithContext(ctx).First(&anchor, categoryID).Error; err != nil {
		r.log.Error("get subtree anchor failed", zap.Error(err))
		return nil, err
	}
	var rows []model.CategoryModel
	err := r.db.WithContext(ctx).
		Where("lft BETWEEN ? AND ?", anchor.Lft, anchor.Rgt).
		Order("lft ASC").
		Find(&rows).Error
	if err != nil {
		r.log.Error("get category subtree failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// GetByID returns a category by primary key.
func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*model.CategoryModel, error) {
	var c model.CategoryModel
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		r.log.Error("find category by id failed", zap.Error(err))
		return nil, err
	}
	return &c, nil
}

// FindByIDs returns categories for the given ids (used by posts and other features).
func (r *CategoryRepository) FindByIDs(ctx context.Context, ids []uint) ([]model.CategoryModel, error) {
	if len(ids) == 0 {
		return nil, errors.New("no categories found")
	}
	var rows []model.CategoryModel
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	if err != nil {
		r.log.Error("find categories by ids failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// UpdateName updates the category display name (nested-set lft/rgt/depth unchanged).
func (r *CategoryRepository) UpdateName(ctx context.Context, id uint, name string, actorUserID uint) (*model.CategoryModel, error) {
	var c model.CategoryModel
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		r.log.Error("find category for update failed", zap.Error(err))
		return nil, err
	}
	dupQ := r.db.WithContext(ctx).Model(&model.CategoryModel{}).Where("name = ? AND id <> ?", name, id)
	if c.ParentID == nil {
		dupQ = dupQ.Where("parent_id IS NULL")
	} else {
		dupQ = dupQ.Where("parent_id = ?", *c.ParentID)
	}
	var dup int64
	if err := dupQ.Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, gorm.ErrDuplicatedKey
	}
	newSlug, err := nextCategorySlug(r.db.WithContext(ctx), name, id)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&model.CategoryModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":       name,
		"slug":       newSlug,
		"updated_by": actorUserID,
		"updated_at": time.Now(),
	}).Error; err != nil {
		r.log.Error("update category name failed", zap.Error(err))
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// DeleteSubtree soft-deletes a category and all descendants (nested-set rebalance on active rows only). Fails if any post references a category id in that subtree.
func (r *CategoryRepository) DeleteSubtree(ctx context.Context, id uint, deletedBy uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockNestedSet(tx); err != nil {
			return err
		}
		defer unlockNestedSet(tx)

		var anchor model.CategoryModel
		q := tx
		if tx.Dialector.Name() == "mysql" {
			q = tx.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&anchor, id).Error; err != nil {
			return err
		}
		var ids []uint
		if err := tx.Model(&model.CategoryModel{}).Where("lft BETWEEN ? AND ?", anchor.Lft, anchor.Rgt).Pluck("id", &ids).Error; err != nil {
			return err
		}
		var n int64
		if err := tx.Model(&model.Post{}).Where("category_id IN ?", ids).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return ErrCategorySubtreeHasPosts
		}
		width := anchor.Rgt - anchor.Lft + 1
		now := time.Now()
		if err := tx.Model(&model.CategoryModel{}).Where("lft BETWEEN ? AND ?", anchor.Lft, anchor.Rgt).Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
			"updated_by": deletedBy,
		}).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE categories SET rgt = rgt - ? WHERE deleted_at IS NULL AND rgt > ?", width, anchor.Rgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE categories SET lft = lft - ? WHERE deleted_at IS NULL AND lft > ?", width, anchor.Rgt).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r.log.Error("delete category subtree failed", zap.Error(err))
		return err
	}
	return nil
}
