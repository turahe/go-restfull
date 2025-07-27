package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
	Search   string `json:"search"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
}

// PaginationMiddleware extracts and validates pagination parameters
func PaginationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse pagination parameters from query string
		pageStr := c.Query("page", "1")
		perPageStr := c.Query("per_page", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "created_at")
		sortDescStr := c.Query("sort_desc", "true")

		// Convert to integers with defaults
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		perPage, err := strconv.Atoi(perPageStr)
		if err != nil || perPage < 1 {
			perPage = 10
		}

		// Ensure reasonable limits
		if perPage > 100 {
			perPage = 100
		}

		sortDesc := sortDescStr == "true" || sortDescStr == "1"
		offset := (page - 1) * perPage

		// Store pagination parameters in context
		pagination := PaginationParams{
			Page:     page,
			PerPage:  perPage,
			Search:   search,
			SortBy:   sortBy,
			SortDesc: sortDesc,
		}

		c.Locals("pagination", pagination)
		c.Locals("offset", offset)

		return c.Next()
	}
}

// GetPaginationParams retrieves pagination parameters from context
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	pagination := c.Locals("pagination")
	if pagination == nil {
		return PaginationParams{
			Page:     1,
			PerPage:  10,
			SortBy:   "created_at",
			SortDesc: true,
		}
	}
	return pagination.(PaginationParams)
}

// GetOffset retrieves offset from context
func GetOffset(c *fiber.Ctx) int {
	offset := c.Locals("offset")
	if offset == nil {
		return 0
	}
	return offset.(int)
}
