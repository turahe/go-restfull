package request

type PageRequest struct {
	Page  int `form:"page" json:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" json:"limit" binding:"omitempty,min=1,max=200"`
}

type SortRequest struct {
	Sort   string `form:"sort" json:"sort" binding:"omitempty,oneof=asc desc"`
	SortBy string `form:"sortBy" json:"sortBy" binding:"omitempty,min=1,max=255"`
}

type SearchRequest struct {
	Search string `form:"search" json:"search" binding:"omitempty,min=1,max=255"`
}
