package healthz

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"webapi/internal/http/response"
)

type HealthzHTTPHandler struct{}

func NewHealthzHTTPHandler() *HealthzHTTPHandler {
	return &HealthzHTTPHandler{}
}

// Healthz godoc
// @Summary Health check endpoint
// @Description Check if the API is running and healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.CommonResponse
// @Router /healthz [get]
func (h *HealthzHTTPHandler) Healthz(c *fiber.Ctx) error {
	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
	})
}
