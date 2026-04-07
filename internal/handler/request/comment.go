package request

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
	TagIDs  []uint `json:"tagIds" binding:"omitempty,dive,gt=0"`
}

type UpdateCommentBody struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentListRequest struct {
	PageRequest
	SearchRequest
	PostID uint `form:"postId" json:"postId" binding:"required,gt=0"`
}
