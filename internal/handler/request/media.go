package request

type MediaListRequest struct {
	PageRequest
	SearchRequest
	Name string `form:"name" json:"name" binding:"omitempty,min=1,max=255"`
}
