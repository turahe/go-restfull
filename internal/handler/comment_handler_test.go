package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/internal/middleware"
	"go-rest/internal/model"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCommentService struct{ mock.Mock }

func (m *mockCommentService) Create(ctx context.Context, postID uint, userID uint, content string, tagIDs []uint) (*model.Comment, error) {
	args := m.Called(ctx, postID, userID, content, tagIDs)
	c, _ := args.Get(0).(*model.Comment)
	return c, args.Error(1)
}
func (m *mockCommentService) List(ctx context.Context, postID uint, limit int) ([]model.Comment, error) {
	args := m.Called(ctx, postID, limit)
	return args.Get(0).([]model.Comment), args.Error(1)
}

func decodeEnvCmt(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func withAuthAny() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_claims", middleware.AuthClaims{Role: "user", UserID: 1, SessionID: "s"})
		c.Next()
	}
}

func TestCommentHandler_List_InvalidPostID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &mockCommentService{}
	h := NewCommentHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/posts/:id/comments", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/bad/comments", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvCmt(t, rr)
	assert.Equal(t, "invalid post id", env.Message)
	svc.AssertExpectations(t)
}

func TestCommentHandler_Create_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &mockCommentService{}
	h := NewCommentHandler(svc, nil)
	r := gin.New()
	r.POST("/api/v1/posts/:id/comments", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewBufferString(`{"content":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	svc.AssertExpectations(t)
}

func TestCommentHandler_Create_PostMissing(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &mockCommentService{}
	svc.On("Create", mock.Anything, uint(1), uint(1), "hi", ([]uint)(nil)).
		Return((*model.Comment)(nil), service.ErrPostMissing).Once()

	h := NewCommentHandler(svc, nil)
	r := gin.New()
	r.Use(withAuthAny())
	r.POST("/api/v1/posts/:id/comments", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewBufferString(`{"content":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	env := decodeEnvCmt(t, rr)
	assert.Equal(t, "not found", env.Message)
	svc.AssertExpectations(t)
}

