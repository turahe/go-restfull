package pagination

import (
	"fmt"
	"math"
	"strconv"

	"github.com/turahe/go-restfull/internal/logger"

	"go.uber.org/zap"
)

// PaginationRequest represents the pagination parameters from HTTP requests
type PaginationRequest struct {
	Page     int    `json:"page" query:"page"`
	PerPage  int    `json:"per_page" query:"per_page"`
	SortBy   string `json:"sort_by" query:"sort_by"`
	SortDesc bool   `json:"sort_desc" query:"sort_desc"`
	Search   string `json:"search" query:"search"`
}

// PaginationResponse represents the pagination metadata in API responses
type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     int   `json:"next_page,omitempty"`
	PreviousPage int   `json:"previous_page,omitempty"`
	From         int   `json:"from"`
	To           int   `json:"to"`
}

// PaginatedResult represents a paginated response with data and metadata
type PaginatedResult struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationDTI struct {
	Page     string `json:"page" validate:"required"`
	PerPage  string `json:"perPage" validate:"required"`
	SortBy   string `json:"sortBy" validate:"required"`
	SortDesc string `json:"sortDesc" validate:"required"`
}

type PaginationDTO struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Page    int  `json:"page"`
	HasMore bool `json:"has_more"`
}

// NewPaginationRequest creates a new pagination request with default values
func NewPaginationRequest() *PaginationRequest {
	return &PaginationRequest{
		Page:     1,
		PerPage:  10,
		SortBy:   "created_at",
		SortDesc: true,
	}
}

// ParseFromQuery parses pagination parameters from query string
func ParseFromQuery(pageStr, perPageStr, sortBy, sortDescStr, search string) *PaginationRequest {
	req := NewPaginationRequest()

	// Parse page
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	// Parse per_page
	if perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 {
			req.PerPage = perPage
		}
	}

	// Parse sort_by
	if sortBy != "" {
		req.SortBy = sortBy
	}

	// Parse sort_desc
	if sortDescStr != "" {
		req.SortDesc = sortDescStr == "true" || sortDescStr == "1"
	}

	// Parse search
	if search != "" {
		req.Search = search
	}

	return req
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLimit returns the limit for database queries
func (p *PaginationRequest) GetLimit() int {
	return p.PerPage
}

// GetOrderBy generates the ORDER BY clause for SQL queries
func (p *PaginationRequest) GetOrderBy() string {
	if p.SortBy == "" {
		return "ORDER BY created_at DESC"
	}

	order := "ORDER BY " + p.SortBy
	if p.SortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	return order
}

// CreatePaginationResponse creates a pagination response from request and total count
func CreatePaginationResponse(req *PaginationRequest, totalItems int64) PaginationResponse {
	totalPages := int(math.Ceil(float64(totalItems) / float64(req.PerPage)))

	// Ensure current page doesn't exceed total pages
	if req.Page > totalPages && totalPages > 0 {
		req.Page = totalPages
	}

	from := (req.Page-1)*req.PerPage + 1
	to := req.Page * req.PerPage
	if to > int(totalItems) {
		to = int(totalItems)
	}
	if totalItems == 0 {
		from = 0
		to = 0
	}

	hasNextPage := req.Page < totalPages
	hasPrevPage := req.Page > 1

	var nextPage, prevPage int
	if hasNextPage {
		nextPage = req.Page + 1
	}
	if hasPrevPage {
		prevPage = req.Page - 1
	}

	return PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.PerPage,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		HasNextPage:  hasNextPage,
		HasPrevPage:  hasPrevPage,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		From:         from,
		To:           to,
	}
}

// CreatePaginatedResult creates a complete paginated result
func CreatePaginatedResult(data interface{}, req *PaginationRequest, totalItems int64) *PaginatedResult {
	return &PaginatedResult{
		Data:       data,
		Pagination: CreatePaginationResponse(req, totalItems),
	}
}

func ConvertPaginationToStrSql(pag *PaginationDTI) (string, error) {
	resultSql := ""

	//set s
	if pag.SortBy != "" {
		resultSql += " ORDER BY " + pag.SortBy
		if pag.SortDesc == "true" {
			resultSql += " DESC"
		}
	}

	if pag.Page != "" && pag.PerPage != "" {
		limit, err := strconv.Atoi(pag.PerPage)
		if err != nil {
			logger.Log.Error("convert string to int error.", zap.Error(err))
			return "", err
		}

		page, err := strconv.Atoi(pag.Page)
		if err != nil {
			logger.Log.Error("convert string to int error.", zap.Error(err))
			return "", err
		}

		offset := (page - 1) * limit
		resultSql += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)
	}

	return resultSql, nil
}

func GetResponsePagination(pagDTI *PaginationDTI, total int) (PaginationDTO, error) {

	var resPag PaginationDTO

	resPag.Total = total
	if pagDTI.Page != "" && pagDTI.PerPage != "" {
		page, err := strconv.Atoi(pagDTI.Page)
		if err != nil {
			logger.Log.Error("convert string to int error.", zap.Error(err))
			return resPag, err
		}
		perPage, err := strconv.Atoi(pagDTI.PerPage)
		if err != nil {
			logger.Log.Error("convert string to int error.", zap.Error(err))
			return resPag, err
		}
		resPag.Page = page
		resPag.Limit = perPage
		resPag.HasMore = isHasMore(page, perPage, total)
	}
	return resPag, nil
}

func isHasMore(page int, limit int, total int) bool {
	return total > (page * limit)
}
