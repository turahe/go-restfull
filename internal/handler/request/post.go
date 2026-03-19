package request

type CreatePostRequest struct {
	Title      string `json:"title" binding:"required,min=3,max=200"`
	Content    string `json:"content" binding:"required,min=1"`
	CategoryID uint   `json:"categoryId" binding:"required,gt=0"`
	TagIDs     []uint `json:"tagIds" binding:"omitempty,dive,gt=0"`
}

type UpdatePostRequest struct {
	Title      string `json:"title" binding:"omitempty,min=3,max=200"`
	Content    string `json:"content" binding:"omitempty,min=1"`
	CategoryID *uint  `json:"categoryId" binding:"omitempty,gt=0"`
	TagIDs     []uint `json:"tagIds" binding:"omitempty,dive,gt=0"`
}

type PostListRequest struct {
	Limit      int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=200"`
	Title      string `form:"title" json:"title" binding:"omitempty,min=3,max=200"`
	CategoryID *uint  `form:"categoryId" json:"categoryId" binding:"omitempty,gt=0"`
	Content    string `form:"content" json:"content" binding:"omitempty,min=1"`
	Page       int    `form:"page" json:"page" binding:"omitempty,min=1"`
}
