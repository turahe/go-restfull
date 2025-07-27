package services

import (
	"context"
	"webapi/internal/helper/pagination"
	"webapi/internal/http/response"
)

// PaginationService defines the interface for pagination operations
type PaginationService interface {
	// CreatePaginatedResponse creates a standardized paginated response
	CreatePaginatedResponse(ctx context.Context, data interface{}, total int64, paginationParams *pagination.PaginationRequest) *response.PaginatedResult

	// ValidatePaginationParams validates and normalizes pagination parameters
	ValidatePaginationParams(params *pagination.PaginationRequest) *pagination.PaginationRequest
}

// PaginationServiceImpl implements the PaginationService interface
type PaginationServiceImpl struct{}

// NewPaginationService creates a new pagination service instance
func NewPaginationService() PaginationService {
	return &PaginationServiceImpl{}
}

// CreatePaginatedResponse creates a standardized paginated response
func (p *PaginationServiceImpl) CreatePaginatedResponse(ctx context.Context, data interface{}, total int64, paginationParams *pagination.PaginationRequest) *response.PaginatedResult {
	// Validate and normalize pagination parameters
	validatedParams := p.ValidatePaginationParams(paginationParams)

	// Create paginated result using the existing helper function
	return response.CreatePaginatedResult(data, validatedParams.Page, validatedParams.PerPage, total)
}

// ValidatePaginationParams validates and normalizes pagination parameters
func (p *PaginationServiceImpl) ValidatePaginationParams(params *pagination.PaginationRequest) *pagination.PaginationRequest {
	if params == nil {
		params = pagination.NewPaginationRequest()
	}

	// Ensure minimum values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	// SortDesc is already a boolean, no need to normalize

	return params
}
