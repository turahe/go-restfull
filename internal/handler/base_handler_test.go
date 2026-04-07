package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testReq struct {
	Name string `json:"name" binding:"required"`
}

func decodeEnvBase(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func TestBaseHandler_bindJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := BaseHandler{}
	r := gin.New()
	r.POST("/x", func(c *gin.Context) {
		var req testReq
		_ = h.bindJSON(c, response.ServiceCodeCommon, &req)
	})

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvBase(t, rr)
	assert.Equal(t, "invalid request", env.Message)
}

func TestBaseHandler_bindJSON_ValidationError(t *testing.T) {
	t.Parallel()

	h := BaseHandler{}
	r := gin.New()
	r.POST("/x", func(c *gin.Context) {
		var req testReq
		_ = h.bindJSON(c, response.ServiceCodeCommon, &req)
	})

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvBase(t, rr)
	assert.Equal(t, "validation failed", env.Message)
}
