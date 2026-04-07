package request

type CreateCategoryRootBody struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
}

type CreateCategoryChildBody struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
}

type UpdateCategoryBody struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
}
