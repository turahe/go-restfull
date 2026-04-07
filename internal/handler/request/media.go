package request

type MediaListRequest struct {
	PageRequest
	SearchRequest
	Name string `form:"name" json:"name" binding:"omitempty,min=1,max=255"`
}

type CreateMediaFolderRootBody struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type CreateMediaFolderChildBody struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type UpdateMediaBody struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}
