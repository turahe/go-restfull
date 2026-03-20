package handler

import (
	"net/http"

	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SettingsHandler struct {
	BaseHandler
	settings *service.SettingsService
}

func NewSettingsHandler(settings *service.SettingsService, log *zap.Logger) *SettingsHandler {
	return &SettingsHandler{
		BaseHandler: BaseHandler{Log: log},
		settings:    settings,
	}
}

// Get godoc
// @Summary      Public application settings
// @Description  Returns public DB-backed application settings (rows where `is_public=true`).
// @Tags         Settings
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Router       /api/v1/settings [get]
func (h *SettingsHandler) Get(c *gin.Context) {
	data, err := h.settings.Public(c.Request.Context())
	if err != nil {
		h.internalError(c, response.ServiceCodeSettings, err, "settings load failed")
		return
	}
	response.OK(
		c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodeSettings, response.CaseCodeRetrieved),
		"Successfully retrieved settings",
		data,
	)
}
