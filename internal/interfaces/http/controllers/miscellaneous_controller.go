package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/turahe/go-restfull/pkg/exception"
)

type MiscellaneousHTTPHandler struct{}

func NewMiscellaneousHTTPHandler() *MiscellaneousHTTPHandler {
	return &MiscellaneousHTTPHandler{}
}

func (m *MiscellaneousHTTPHandler) NotFound(c *fiber.Ctx) error {
	return exception.ApiNotFoundError
}
