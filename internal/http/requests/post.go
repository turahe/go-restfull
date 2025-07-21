package requests

type CreatePostRequest struct {
	Slug           string   `json:"slug" validate:"omitempty,min=3,max=100"`
	Title          string   `json:"title" validate:"required,min=3,max=255"`
	Subtitle       string   `json:"subtitle" validate:"max=255"`
	Description    string   `json:"description" validate:"max=1000"`
	Content        string   `json:"content" validate:"required"`
	Type           string   `json:"type" validate:"required,oneof=blog article book"`
	IsSticky       bool     `json:"is_sticky"`
	PublishedAt    int64    `json:"published_at"`
	Language       string   `json:"language" validate:"required"`
	Layout         string   `json:"layout" validate:"max=100"`
	RecordOrdering int64    `json:"record_ordering"`
	Tags           []string `json:"tags" validate:"dive,uuid4"`
}

type UpdatePostRequest struct {
	Slug           string   `json:"slug" validate:"omitempty,min=3,max=100"`
	Title          string   `json:"title" validate:"required,min=3,max=255"`
	Subtitle       string   `json:"subtitle" validate:"max=255"`
	Description    string   `json:"description" validate:"max=1000"`
	Content        string   `json:"content" validate:"required"`
	Type           string   `json:"type" validate:"required,oneof=blog article book"`
	IsSticky       bool     `json:"is_sticky"`
	PublishedAt    int64    `json:"published_at"`
	Language       string   `json:"language" validate:"required"`
	Layout         string   `json:"layout" validate:"max=100"`
	RecordOrdering int64    `json:"record_ordering"`
	Tags           []string `json:"tags" validate:"dive,uuid4"`
}
