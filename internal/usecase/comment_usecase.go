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
	ErrCommentPostMissing     = errors.New("post not found")
	ErrCommentInvalid         = errors.New("invalid comment")
	ErrCommentInvalidPostRef  = errors.New("invalid post reference")
	ErrCommentNotFound        = errors.New("comment not found")
	ErrCommentParentNotFound  = errors.New("parent comment not found")
	ErrCommentSubtreeHasMedia = errors.New("cannot delete comment subtree: media still attached")
	ErrCommentInvalidContent  = errors.New("invalid comment content")
)

// CommentTreeNode is the JSON shape for tree responses (id, name from content, children).
type CommentTreeNode struct {
	ID       uint              `json:"id"`
	Name     string            `json:"name"`
	Children []CommentTreeNode `json:"children"`
}

type CommentUsecase struct {
	comments *repository.CommentRepository
	tags     *repository.TagRepository
	log      *zap.Logger
}

func NewCommentUsecase(comments *repository.CommentRepository, tags *repository.TagRepository, log *zap.Logger) *CommentUsecase {
	return &CommentUsecase{comments: comments, tags: tags, log: log}
}

func (u *CommentUsecase) CreateRoot(ctx context.Context, postID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error) {
	exists, err := u.comments.PostExists(ctx, postID)
	if err != nil {
		u.log.Error("failed to check if post exists", zap.Error(err))
		return nil, err
	}
	if !exists {
		return nil, ErrCommentPostMissing
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, ErrCommentInvalidContent
	}

	cmt, err := u.comments.CreateRoot(ctx, postID, userID, content, userID)
	if err != nil {
		u.log.Error("failed to create root comment", zap.Error(err))
		return nil, err
	}

	if err := u.applyTags(ctx, cmt.ID, req.TagIDs); err != nil {
		return nil, err
	}
	if len(req.TagIDs) > 0 {
		if full, gerr := u.comments.GetByIDInPost(ctx, postID, cmt.ID); gerr == nil {
			cmt = full
		}
	}
	return cmt, nil
}

func (u *CommentUsecase) CreateChild(ctx context.Context, postID uint, parentID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error) {
	exists, err := u.comments.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCommentPostMissing
	}
	if parentID == 0 {
		return nil, ErrCommentParentNotFound
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, ErrCommentInvalidContent
	}

	cmt, err := u.comments.CreateChild(ctx, postID, parentID, userID, content, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentParentNotFound
		}
		u.log.Error("failed to create child comment", zap.Error(err))
		return nil, err
	}

	if err := u.applyTags(ctx, cmt.ID, req.TagIDs); err != nil {
		return nil, err
	}
	if len(req.TagIDs) > 0 {
		if full, gerr := u.comments.GetByIDInPost(ctx, postID, cmt.ID); gerr == nil {
			cmt = full
		}
	}
	return cmt, nil
}

func (u *CommentUsecase) applyTags(ctx context.Context, commentID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 || u.tags == nil {
		return nil
	}
	ids := uniqueUint(tagIDs)
	tags, err := u.tags.FindByIDs(ctx, ids)
	if err != nil {
		u.log.Error("failed to find tags by ids", zap.Error(err))
		return err
	}
	if len(tags) != len(ids) {
		return errors.New("one or more tags not found")
	}
	if err := u.comments.ReplaceTags(ctx, commentID, tags); err != nil {
		u.log.Error("failed to replace tags", zap.Error(err))
		return err
	}
	return nil
}

func (u *CommentUsecase) GetTree(ctx context.Context, postID uint) ([]CommentTreeNode, error) {
	rows, err := u.comments.GetTree(ctx, postID)
	if err != nil {
		return nil, err
	}
	return buildCommentTree(rows), nil
}

func (u *CommentUsecase) GetSubtree(ctx context.Context, postID uint, commentID uint) ([]CommentTreeNode, error) {
	if commentID == 0 {
		return nil, ErrCommentNotFound
	}
	rows, err := u.comments.GetSubtree(ctx, postID, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrCommentNotFound
	}
	return buildCommentTree(rows), nil
}

func (u *CommentUsecase) Update(ctx context.Context, postID uint, commentID uint, userID uint, req request.UpdateCommentBody) (*model.Comment, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, ErrCommentInvalidContent
	}
	cmt, err := u.comments.UpdateContent(ctx, postID, commentID, content, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return cmt, nil
}

func (u *CommentUsecase) Delete(ctx context.Context, postID uint, commentID uint, userID uint) error {
	err := u.comments.DeleteSubtree(ctx, postID, commentID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		if errors.Is(err, repository.ErrCommentSubtreeHasMedia) {
			return ErrCommentSubtreeHasMedia
		}
		return err
	}
	return nil
}

func (u *CommentUsecase) List(ctx context.Context, req request.CommentListRequest) (repository.CursorPage, error) {
	page, err := u.comments.List(ctx, req)
	if err != nil {
		u.log.Error("failed to list comments", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}

func uniqueUint(in []uint) []uint {
	if len(in) == 0 {
		return in
	}
	seen := make(map[uint]struct{}, len(in))
	out := make([]uint, 0, len(in))
	for _, v := range in {
		if v == 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func buildCommentTree(rows []model.Comment) []CommentTreeNode {
	type stackItem struct {
		node *CommentTreeNode
		rgt  int
	}
	var stack []stackItem
	var roots []CommentTreeNode
	for _, row := range rows {
		label := row.Content
		if len(label) > 500 {
			label = label[:500] + "…"
		}
		n := CommentTreeNode{ID: row.ID, Name: label, Children: []CommentTreeNode{}}
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
