package model

type Content struct {
	ID          string `json:"id"`
	ModelType   string `json:"model_type"`
	ModelID     string `json:"model_id"`
	ContentRaw  string `json:"content_raw"`
	ContentHTML string `json:"content_html"`
	CreatedBy   string `json:"created_by"`
	UpdatedBy   string `json:"updated_by"`
}
