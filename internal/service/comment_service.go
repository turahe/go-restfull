package service

import (
	"context"
	"errors"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/repository"
)

var (
	ErrPostMissing     = errors.New("post not found")
	ErrInvalidComment  = errors.New("invalid comment")
	ErrInvalidPostRef  = errors.New("invalid post reference")
)

type CommentService struct {
	comments *repository.CommentRepository
	tags     *repository.TagRepository
}

func NewCommentService(comments *repository.CommentRepository, tags *repository.TagRepository) *CommentService {
	return &CommentService{comments: comments, tags: tags}
}

func (s *CommentService) Create(ctx context.Context, postID uint, userID uint, content string, tagIDs []uint) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if postID == 0 {
		return nil, ErrInvalidPostRef
	}
	if content == "" {
		return nil, ErrInvalidComment
	}

	exists, err := s.comments.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostMissing
	}

	cmt := &model.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.comments.Create(ctx, cmt); err != nil {
		return nil, err
	}

	if len(tagIDs) > 0 && s.tags != nil {
		ids := uniqueUint(tagIDs)
		tags, err := s.tags.FindByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		if len(tags) != len(ids) {
			return nil, errors.New("one or more tags not found")
		}
		if err := s.comments.ReplaceTags(ctx, cmt.ID, tags); err != nil {
			return nil, err
		}
		cmt.Tags = tags
	}
	return cmt, nil
}

func (s *CommentService) List(ctx context.Context, postID uint, limit int) ([]model.Comment, error) {
	if postID == 0 {
		return nil, ErrInvalidPostRef
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.comments.ListByPostID(ctx, postID, limit)
}

