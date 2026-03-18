package request

type CreatePostRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=200"`
	Content     string `json:"content" binding:"required,min=1"`
	CategoryIDs []uint `json:"category_ids" binding:"omitempty,dive,gt=0"`
}

type UpdatePostRequest struct {
	Title       string `json:"title" binding:"omitempty,min=3,max=200"`
	Content     string `json:"content" binding:"omitempty,min=1"`
	CategoryIDs []uint `json:"category_ids" binding:"omitempty,dive,gt=0"`
}

