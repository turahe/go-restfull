package handler

import (
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/pkg/response"
	"strconv"
	"strings"

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
		reqLog := h.Log
		if id, ok := middleware.GetRequestID(c); ok {
			reqLog = reqLog.With(zap.String("request_id", id))
		}
		reqLog.Error(message, zap.Error(err))
	}
	response.InternalServerError(c,
		response.BuildResponseCode(500, serviceCode, response.CaseCodeInternalError),
		"internal error",
		message,
	)
}

func (h BaseHandler) ParseIntDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func (h BaseHandler) ParseUintParam(c *gin.Context, name string) (uint, error) {
	s := strings.TrimSpace(c.Param(name))
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}
