package request

type CreatePostRequest struct {
	Title      string `json:"title" binding:"required,min=3,max=200"`
	Content    string `json:"content" binding:"required,min=1"`
	CategoryID uint   `json:"categoryId" binding:"required,gt=0"`
	Layout     string `json:"layout" binding:"omitempty,oneof=simple author book list"`
	Status     string `json:"status" binding:"omitempty,oneof=draft published archived"`
	// SEO (optional)
	Excerpt         string `json:"excerpt" binding:"omitempty,max=2000"`
	MetaTitle       string `json:"metaTitle" binding:"omitempty,max=200"`
	MetaDescription string `json:"metaDescription" binding:"omitempty,max=320"`
	CanonicalURL    string `json:"canonicalUrl" binding:"omitempty,max=512"`
	OgImageURL      string `json:"ogImageUrl" binding:"omitempty,max=512"`
	RobotsMeta      string `json:"robotsMeta" binding:"omitempty,max=100"`
	TagIDs          []uint `json:"tagIds" binding:"omitempty,dive,gt=0"`
}

type UpdatePostRequest struct {
	Title      string `json:"title" binding:"omitempty,min=3,max=200"`
	Content    string `json:"content" binding:"omitempty,min=1"`
	CategoryID *uint  `json:"categoryId" binding:"omitempty,gt=0"`
	Layout     string `json:"layout" binding:"omitempty,oneof=simple author book list"`
	Status     string `json:"status" binding:"omitempty,oneof=draft published archived"`
	// SEO: use pointers so JSON null/absence can mean "no change"; present string (including "") updates/clears.
	Excerpt         *string `json:"excerpt" binding:"omitempty,max=2000"`
	MetaTitle       *string `json:"metaTitle" binding:"omitempty,max=200"`
	MetaDescription *string `json:"metaDescription" binding:"omitempty,max=320"`
	CanonicalURL    *string `json:"canonicalUrl" binding:"omitempty,max=512"`
	OgImageURL      *string `json:"ogImageUrl" binding:"omitempty,max=512"`
	RobotsMeta      *string `json:"robotsMeta" binding:"omitempty,max=100"`
	TagIDs          []uint  `json:"tagIds" binding:"omitempty,dive,gt=0"`
}

type PostListRequest struct {
	PageRequest
	SearchRequest
	Title      string `form:"title" json:"title" binding:"omitempty,min=3,max=200"`
	CategoryID *uint  `form:"categoryId" json:"categoryId" binding:"omitempty,gt=0"`
	Layout     string `form:"layout" json:"layout" binding:"omitempty,oneof=simple author book list"`
	Status     string `form:"status" json:"status" binding:"omitempty,oneof=draft published archived"`
}
