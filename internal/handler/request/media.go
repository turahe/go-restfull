package request

type MediaListRequest struct {
	Limit int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=200"`
	Page  int    `form:"page" json:"page" binding:"omitempty,min=1"`
	Name  string `form:"name" json:"name" binding:"omitempty,min=1,max=255"`
}
