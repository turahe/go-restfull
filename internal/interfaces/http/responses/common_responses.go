package responses

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// PaginationResponse represents a paginated API response
type PaginationResponse struct {
	Status       string      `json:"status"`
	Data         interface{} `json:"data"`
	TotalCount   int64       `json:"total_count"`
	TotalPage    int         `json:"total_page"`
	CurrentPage  int         `json:"current_page"`
	LastPage     int         `json:"last_page"`
	PerPage      int         `json:"per_page"`
	NextPage     int         `json:"next_page"`
	PreviousPage int         `json:"previous_page"`
	Path         string      `json:"path"`
} 