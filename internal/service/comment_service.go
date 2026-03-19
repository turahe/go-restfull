package service

import (
	"context"
	"errors"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"
	"go-rest/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrPostMissing    = errors.New("post not found")
	ErrInvalidComment = errors.New("invalid comment")
	ErrInvalidPostRef = errors.New("invalid post reference")
)

type CommentService struct {
	comments *repository.CommentRepository
	tags     *repository.TagRepository
	log      *zap.Logger
}

func NewCommentService(comments *repository.CommentRepository, tags *repository.TagRepository, log *zap.Logger) *CommentService {
	return &CommentService{comments: comments, tags: tags, log: log}
}

func (s *CommentService) Create(ctx context.Context, postID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error) {

	exists, err := s.comments.PostExists(ctx, postID)
	if err != nil {
		s.log.Error("failed to check if post exists", zap.Error(err))
		return nil, err
	}
	if !exists {
		s.log.Error("post not found")
		return nil, ErrPostMissing
	}

	cmt := &model.Comment{
		PostID:    postID,
		UserID:    userID,
		Content:   req.Content,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.comments.Create(ctx, cmt); err != nil {
		s.log.Error("failed to create comment", zap.Error(err))
		return nil, err
	}

	if len(req.TagIDs) > 0 && s.tags != nil {
		ids := UniqueUint(req.TagIDs)
		tags, err := s.tags.FindByIDs(ctx, ids)
		if err != nil {
			s.log.Error("failed to find tags by ids", zap.Error(err))
			return nil, err
		}
		if len(tags) != len(UniqueUint(req.TagIDs)) {
			s.log.Error("one or more tags not found")
			return nil, errors.New("one or more tags not found")
		}
		if err := s.comments.ReplaceTags(ctx, cmt.ID, tags); err != nil {
			s.log.Error("failed to replace tags", zap.Error(err))
			return nil, err
		}
		cmt.Tags = tags
	}
	return cmt, nil
}

func (s *CommentService) List(ctx context.Context, req request.CommentListRequest) (repository.CursorPage, error) {
	page, err := s.comments.List(ctx, req)
	if err != nil {
		s.log.Error("failed to list comments", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}
