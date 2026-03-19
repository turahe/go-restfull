package request

type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UpdateTagRequest struct {
	Name string `json:"name" binding:"omitempty,min=2,max=100"`
}

type TagListRequest struct {
	Limit int    `json:"limit" binding:"omitempty,min=1,max=200"`
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Page  int    `json:"page" binding:"omitempty,min=1"`
}
