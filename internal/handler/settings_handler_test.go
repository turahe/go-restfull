package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSettingsHandler_Get(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := service.NewSettingsService(nil)
	h := NewSettingsHandler(svc, zap.NewNop())

	r := gin.New()
	r.GET("/settings", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var env response.Envelope
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &env))
	assert.Equal(t, "Successfully retrieved settings", env.Message)

	raw, err := json.Marshal(env.Data)
	require.NoError(t, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(raw, &got))

	assert.Equal(t, "en", got["defaultLocale"])
	assert.Equal(t, "false", got["maintenanceMode"])
	assert.Equal(t, "Blog API powered by Go, Gin, and GORM.", got["siteDescription"])
	assert.Equal(t, "Go REST Blog", got["siteTitle"])
}
