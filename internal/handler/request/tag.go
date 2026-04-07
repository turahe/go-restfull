package request

type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UpdateTagRequest struct {
	Name string `json:"name" binding:"omitempty,min=2,max=100"`
}

type TagListRequest struct {
	PageRequest
	SearchRequest
	Name string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
}
