package handler

import (
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BaseHandler centralizes shared handler concerns:
// - consistent bind + validation errors (Laravel-style)
// - standard response codes
// - logging for internal errors
type BaseHandler struct {
	Log *zap.Logger
}

func (h BaseHandler) bindJSON(c *gin.Context, serviceCode string, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		if payload, ok := bindError(err); ok {
			response.BadRequest(c,
				response.BuildResponseCode(400, serviceCode, response.CaseCodeValidationError),
				"validation failed",
				payload,
			)
			return false
		}
		response.BadRequest(c,
			response.BuildResponseCode(400, serviceCode, response.CaseCodeInvalidFormat),
			"invalid request",
			err.Error(),
		)
		return false
	}
	return true
}

func (h BaseHandler) validate(c *gin.Context, serviceCode string, req any) bool {
	if errs := validateStructLaravel(req); errs != nil {
		response.BadRequest(c,
			response.BuildResponseCode(400, serviceCode, response.CaseCodeValidationError),
			"validation failed",
			errs,
		)
		return false
	}
	return true
}

func (h BaseHandler) internalError(c *gin.Context, serviceCode string, err error, message string) {
	if h.Log != nil && err != nil {
		h.Log.Error(message, zap.Error(err))
	}
	response.InternalServerError(c,
		response.BuildResponseCode(500, serviceCode, response.CaseCodeInternalError),
		"internal error",
		message,
	)
}

