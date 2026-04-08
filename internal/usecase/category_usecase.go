package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound       = errors.New("category not found")
	ErrInvalidName            = errors.New("invalid category name")
	ErrCategoryDuplicateName  = errors.New("category name already exists under this parent")
	ErrCategoryDeleteHasPosts = errors.New("cannot delete category: posts still reference this category or its descendants")
)

// CategoryTreeNode is the strict JSON shape for tree and subtree responses (id, name, children only).
type CategoryTreeNode struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Children []CategoryTreeNode `json:"children"`
}

type CategoryUsecase struct {
	repo *repository.CategoryRepository
	log  *zap.Logger
}

func NewCategoryUsecase(repo *repository.CategoryRepository, log *zap.Logger) *CategoryUsecase {
	return &CategoryUsecase{repo: repo, log: log}
}

func (u *CategoryUsecase) CreateRoot(ctx context.Context, name string, actorUserID uint) (*model.CategoryModel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}
	out, err := u.repo.CreateRoot(ctx, name, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrCategoryDuplicateName
		}
		return nil, err
	}
	return out, nil
}

func (u *CategoryUsecase) CreateChild(ctx context.Context, parentID uint, name string, actorUserID uint) (*model.CategoryModel, error) {
	if parentID == 0 {
		return nil, ErrCategoryNotFound
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}
	out, err := u.repo.CreateChild(ctx, parentID, name, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrCategoryDuplicateName
		}
		return nil, err
	}
	return out, nil
}

func (u *CategoryUsecase) List(ctx context.Context, req request.CategoryListRequest) (repository.CursorPage, error) {
	return u.repo.List(ctx, req)
}

func (u *CategoryUsecase) GetTree(ctx context.Context) ([]CategoryTreeNode, error) {
	rows, err := u.repo.GetTree(ctx)
	if err != nil {
		return nil, err
	}
	return buildTree(rows), nil
}

func (u *CategoryUsecase) GetSubtree(ctx context.Context, categoryID uint) ([]CategoryTreeNode, error) {
	if categoryID == 0 {
		return nil, ErrCategoryNotFound
	}
	rows, err := u.repo.GetSubtree(ctx, categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrCategoryNotFound
	}
	return buildTree(rows), nil
}

// Update changes the category name (id must exist).
func (u *CategoryUsecase) Update(ctx context.Context, id uint, name string, actorUserID uint) (*model.CategoryModel, error) {
	if id == 0 {
		return nil, ErrCategoryNotFound
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}
	c, err := u.repo.UpdateName(ctx, id, name, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrCategoryDuplicateName
		}
		return nil, err
	}
	return c, nil
}

// Delete soft-deletes the category and its entire subtree. Blocked if any post references a category id in that subtree.
func (u *CategoryUsecase) Delete(ctx context.Context, id uint, actorUserID uint) error {
	if id == 0 {
		return ErrCategoryNotFound
	}
	err := u.repo.DeleteSubtree(ctx, id, actorUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		if errors.Is(err, repository.ErrCategorySubtreeHasPosts) {
			return ErrCategoryDeleteHasPosts
		}
		return err
	}
	return nil
}

// buildTree builds a forest from a flat list ordered by lft (O(n)).
func buildTree(rows []model.CategoryModel) []CategoryTreeNode {
	type stackItem struct {
		node *CategoryTreeNode
		rgt  int
	}
	var stack []stackItem
	var roots []CategoryTreeNode
	for _, row := range rows {
		n := CategoryTreeNode{ID: row.ID, Name: row.Name, Children: []CategoryTreeNode{}}
		for len(stack) > 0 && stack[len(stack)-1].rgt < row.Lft {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, n)
			stack = append(stack, stackItem{node: &roots[len(roots)-1], rgt: row.Rgt})
			continue
		}
		p := stack[len(stack)-1].node
		p.Children = append(p.Children, n)
		stack = append(stack, stackItem{node: &p.Children[len(p.Children)-1], rgt: row.Rgt})
	}
	return roots
}
