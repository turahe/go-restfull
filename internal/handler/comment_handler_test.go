package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/model"
	"go-rest/internal/service"
	"go-rest/internal/repository"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCommentService struct{ mock.Mock }

func (m *mockCommentService) Create(ctx context.Context, postID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error) {
	args := m.Called(ctx, postID, userID, req)
	c, _ := args.Get(0).(*model.Comment)
	return c, args.Error(1)
}
func (m *mockCommentService) List(ctx context.Context, req request.CommentListRequest) (repository.CursorPage, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(repository.CursorPage), args.Error(1)
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

	svc := &mockCommentService{}
	svc.On("Create", mock.Anything, uint(1), uint(1), request.CreateCommentRequest{Content: "hi", TagIDs: nil}).
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

