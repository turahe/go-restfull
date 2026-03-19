package handler

import (
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

type mockCategoryService struct{ mock.Mock }

func (m *mockCategoryService) List(ctx context.Context, limit int) ([]model.Category, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]model.Category), args.Error(1)
}
func (m *mockCategoryService) GetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	args := m.Called(ctx, slug)
	c, _ := args.Get(0).(*model.Category)
	return c, args.Error(1)
}
func (m *mockCategoryService) Create(ctx context.Context, actorUserID uint, name string) (*model.Category, error) {
	args := m.Called(ctx, actorUserID, name)
	c, _ := args.Get(0).(*model.Category)
	return c, args.Error(1)
}
func (m *mockCategoryService) Update(ctx context.Context, id uint, actorUserID uint, name string) (*model.Category, error) {
	args := m.Called(ctx, id, actorUserID, name)
	c, _ := args.Get(0).(*model.Category)
	return c, args.Error(1)
}
func (m *mockCategoryService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	return m.Called(ctx, id, actorUserID).Error(0)
}

func decodeEnvCat(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func withAuthRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_claims", middleware.AuthClaims{Role: role, UserID: 1, SessionID: "s"})
		c.Next()
	}
}

func TestCategoryHandler_List_InvalidLimit(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &mockCategoryService{}
	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/categories", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?limit=bad", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvCat(t, rr)
	assert.Equal(t, "invalid limit", env.Message)
	svc.AssertExpectations(t)
}

func TestCategoryHandler_GetBySlug_NotFound(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &mockCategoryService{}
	svc.On("GetBySlug", mock.Anything, "tech").Return((*model.Category)(nil), service.ErrCategoryNotFound).Once()

	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/categories/:slug", h.GetBySlug)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/tech", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	env := decodeEnvCat(t, rr)
	assert.Equal(t, "not found", env.Message)
	svc.AssertExpectations(t)
}

func TestCategoryHandler_Create_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := &mockCategoryService{}
	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.POST("/api/v1/categories", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	svc.AssertExpectations(t)
}

