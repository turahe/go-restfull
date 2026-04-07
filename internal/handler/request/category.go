package request

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type CategoryListRequest struct {
	PageRequest
	SearchRequest
	Name string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
}
