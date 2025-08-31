package controllers

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/helper/pagination"
	"github.com/turahe/go-restfull/pkg/logger"
	"go.uber.org/zap")

// SearchController handles search-related HTTP requests
type SearchController struct {
	hybridSearchService ports.HybridSearchService
}

// NewSearchController creates a new search controller
func NewSearchController(searchService ports.HybridSearchService) *SearchController {
	return &SearchController{
		hybridSearchService: searchService,
	}
}

// SearchRequest represents the search request structure
type SearchRequest struct {
	Query    string `json:"query" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// SearchResponse represents the search response structure
type SearchResponse struct {
	Query      string                        `json:"query"`
	Type       string                        `json:"type"`
	Results    interface{}                   `json:"results"`
	Pagination pagination.PaginationResponse `json:"pagination"`
	Engine     string                        `json:"search_engine"`
}

// GetSearchStatus returns the current search service status
func (c *SearchController) GetSearchStatus(ctx *fiber.Ctx) error {
	searchInfo := c.hybridSearchService.GetSearchEngineInfo()

	status := fiber.Map{
		"search_engine": searchInfo,
		"endpoints": fiber.Map{
			"unified_search": "/api/v1/search",
			"type_search":    "/api/v1/search/:type",
			"status":         "/api/v1/search",
		},
		"supported_types": []string{
			"posts", "users", "organizations", "tags", "taxonomies",
			"media", "menus", "roles", "content", "addresses", "comments", "all",
		},
		"features": fiber.Map{
			"meilisearch":    searchInfo["meilisearch_available"].(bool),
			"sql_fallback":   true,
			"pagination":     true,
			"type_filtering": true,
		},
	}

	return ctx.JSON(status)
}

// Search performs a search across all supported types
func (c *SearchController) Search(ctx *fiber.Ctx) error {
	var req SearchRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize

	// Perform search based on type
	var results interface{}

	switch req.Type {
	case "posts":
		posts, err := c.hybridSearchService.SearchPosts(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search posts",
			})
	}
		results = posts

	case "users":
		users, err := c.hybridSearchService.SearchUsers(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search users",
			})
	}
		results = users

	case "organizations":
		orgs, err := c.hybridSearchService.SearchOrganizations(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search organizations",
			})
	}
		results = orgs

	case "tags":
		tags, err := c.hybridSearchService.SearchTags(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search tags",
			})
	}
		results = tags

	case "taxonomies":
		taxonomies, err := c.hybridSearchService.SearchTaxonomies(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search taxonomies",
			})
	}
		results = taxonomies

	case "media":
		media, err := c.hybridSearchService.SearchMedia(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search media",
			})
	}
		results = media

	case "menus":
		menus, err := c.hybridSearchService.SearchMenus(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search menus",
			})
	}
		results = menus

	case "roles":
		roles, err := c.hybridSearchService.SearchRoles(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search roles",
			})
	}
		results = roles

	case "content":
		content, err := c.hybridSearchService.SearchContent(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search content",
			})
	}
		results = content

	case "addresses":
		addresses, err := c.hybridSearchService.SearchAddresses(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search addresses",
			})
	}
		results = addresses

	case "comments":
		comments, err := c.hybridSearchService.SearchComments(ctx.Context(), req.Query, req.PageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search comments",
			})
	}
		results = comments

	case "all", "":
		// Search across all types
		allResults := make(map[string]interface{})

		// Search posts
		if posts, err := c.hybridSearchService.SearchPosts(ctx.Context(), req.Query, 5, 0); err == nil {
			allResults["posts"] = posts
		}

		// Search users
		if users, err := c.hybridSearchService.SearchUsers(ctx.Context(), req.Query, 5, 0); err == nil {
			allResults["users"] = users
		}

		// Search organizations
		if orgs, err := c.hybridSearchService.SearchOrganizations(ctx.Context(), req.Query, 5, 0); err == nil {
			allResults["organizations"] = orgs
		}

		// Search tags
		if tags, err := c.hybridSearchService.SearchTags(ctx.Context(), req.Query, 5, 0); err == nil {
			allResults["tags"] = tags
		}

		// Search taxonomies
		if taxonomies, err := c.hybridSearchService.SearchTaxonomies(ctx.Context(), req.Query, 5, 0); err == nil {
			allResults["taxonomies"] = taxonomies
		}

		results = allResults

	default:
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid search type",
		})
	}

	// Create pagination response
	paginationReq := &pagination.PaginationRequest{
		Page:    req.Page,
		PerPage: req.PageSize,
	}

	// For now, we'll use a placeholder total count
	// In a real implementation, you'd get the actual total from the search service
	totalItems := int64(100) // Placeholder

	paginationResp := pagination.CreatePaginationResponse(paginationReq, totalItems)

	response := SearchResponse{
		Query:      req.Query,
		Type:       req.Type,
		Results:    results,
		Pagination: paginationResp,
		Engine:     "hybrid", // Indicates hybrid search (Meilisearch + SQL fallback)
	}

	return ctx.JSON(response)
}

// SearchByType performs a search for a specific type using query parameters
func (c *SearchController) SearchByType(ctx *fiber.Ctx) error {
	searchType := ctx.Params("type")
	query := ctx.Query("q")
	pageStr := ctx.Query("page", "1")
	pageSizeStr := ctx.Query("page_size", "10")

	if query == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var results interface{}

	switch searchType {
	case "posts":
		posts, err := c.hybridSearchService.SearchPosts(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search posts",
			})
	}
		results = posts

	case "users":
		users, err := c.hybridSearchService.SearchUsers(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search users",
			})
	}
		results = users

	case "organizations":
		orgs, err := c.hybridSearchService.SearchOrganizations(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search organizations",
			})
	}
		results = orgs

	case "tags":
		tags, err := c.hybridSearchService.SearchTags(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search tags",
			})
	}
		results = tags

	case "taxonomies":
		taxonomies, err := c.hybridSearchService.SearchTaxonomies(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search taxonomies",
			})
	}
		results = taxonomies

	case "media":
		media, err := c.hybridSearchService.SearchMedia(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search media",
			})
	}
		results = media

	case "menus":
		menus, err := c.hybridSearchService.SearchMenus(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search menus",
			})
	}
		results = menus

	case "roles":
		roles, err := c.hybridSearchService.SearchRoles(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search roles",
			})
	}
		results = roles

	case "content":
		content, err := c.hybridSearchService.SearchContent(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search content",
			})
	}
		results = content

	case "addresses":
		addresses, err := c.hybridSearchService.SearchAddresses(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search addresses",
			})
	}
		results = addresses

	case "comments":
		comments, err := c.hybridSearchService.SearchComments(ctx.Context(), query, pageSize, offset)
		if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search comments",
			})
	}
		results = comments

	default:
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid search type",
		})
	}

	// Create pagination response
	paginationReq := &pagination.PaginationRequest{
		Page:    page,
		PerPage: pageSize,
	}

	// For now, we'll use a placeholder total count
	// In a real implementation, you'd get the actual total from the search service
	totalItems := int64(100) // Placeholder

	paginationResp := pagination.CreatePaginationResponse(paginationReq, totalItems)

	response := SearchResponse{
		Query:      query,
		Type:       searchType,
		Results:    results,
		Pagination: paginationResp,
		Engine:     "hybrid", // Indicates hybrid search (Meilisearch + SQL fallback)
	}

	return ctx.JSON(response)
}
