package controllers

import (
	"github.com/gofiber/fiber/v2"
	"webapi/internal/http/response"
	"webapi/pkg/exception"
)

// Centralized error handler for all routes
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Retrieve necessary details
	// Status code defaults to 500
	responseCode := fiber.StatusInternalServerError
	responseMessage := err.Error()
	requestID := c.Locals("requestid").(string)

	var cErrs *exception.ExceptionErrors

	// Use response code from ExceptionError
	cErrs, ok := err.(*exception.ExceptionErrors)
	if ok {
		responseCode = cErrs.HttpStatusCode
	}

	// Handle 500 error
	return c.Status(responseCode).JSON(
		&response.CommonResponse{
			ResponseCode:    responseCode,
			ResponseMessage: responseMessage,
			Errors:          cErrs,
			RequestID:       requestID,
		},
	)
} 