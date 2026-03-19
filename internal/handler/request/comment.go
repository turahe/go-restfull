package request

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
	TagIDs  []uint `json:"tagIds" binding:"omitempty,dive,gt=0"`
}

type CommentListRequest struct {
	Limit   int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=200"`
	PostID  uint   `form:"postId" json:"postId" binding:"required,gt=0"`
	Content string `form:"content" json:"content" binding:"omitempty,min=1,max=2000"`
	Page    int    `form:"page" json:"page" binding:"omitempty,min=1"`
}
