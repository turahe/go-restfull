package responses

// Response codes
const (
	SYSTEM_OPERATION_SUCCESS = 200
	BAD_REQUEST              = 400
	NOT_FOUND                = 404
	INTERNAL_SERVER_ERROR    = 500
)

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

// CommonResponse represents a common API response structure
type CommonResponse struct {
	ResponseCode    int         `json:"response_code"`
	ResponseMessage string      `json:"response_message"`
	Data            interface{} `json:"data,omitempty"`
	Errors          interface{} `json:"errors,omitempty"`
	RequestID       string      `json:"request_id,omitempty"`
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

// CreatePaginatedResult creates a paginated result
func CreatePaginatedResult(data interface{}, page, perPage int, total int64) *PaginatedResult {
	return &PaginatedResult{
		Data:       data,
		Pagination: CreatePaginationResponse(page, perPage, total),
	}
}

// PaginatedResult represents a paginated result
type PaginatedResult struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// CreatePaginationResponse creates a pagination response
func CreatePaginationResponse(page, perPage int, total int64) PaginationResponse {
	totalPages := (int(total) + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = 0
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 0
	}

	return PaginationResponse{
		Status:       "success",
		Data:         nil,
		TotalCount:   total,
		TotalPage:    totalPages,
		CurrentPage:  page,
		LastPage:     totalPages,
		PerPage:      perPage,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		Path:         "",
	}
}
