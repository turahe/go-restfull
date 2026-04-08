package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCategoryService struct{ mock.Mock }

func (m *mockCategoryService) CreateRoot(ctx context.Context, name string, actorUserID uint) (*model.CategoryModel, error) {
	args := m.Called(ctx, name, actorUserID)
	c, _ := args.Get(0).(*model.CategoryModel)
	return c, args.Error(1)
}

func (m *mockCategoryService) CreateChild(ctx context.Context, parentID uint, name string, actorUserID uint) (*model.CategoryModel, error) {
	args := m.Called(ctx, parentID, name, actorUserID)
	c, _ := args.Get(0).(*model.CategoryModel)
	return c, args.Error(1)
}

func (m *mockCategoryService) List(ctx context.Context, req request.CategoryListRequest) (repository.CursorPage, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(repository.CursorPage), args.Error(1)
}

func (m *mockCategoryService) GetTree(ctx context.Context) ([]service.CategoryTreeNode, error) {
	args := m.Called(ctx)
	var v []service.CategoryTreeNode
	if args.Get(0) != nil {
		v = args.Get(0).([]service.CategoryTreeNode)
	}
	return v, args.Error(1)
}

func (m *mockCategoryService) GetSubtree(ctx context.Context, categoryID uint) ([]service.CategoryTreeNode, error) {
	args := m.Called(ctx, categoryID)
	var v []service.CategoryTreeNode
	if args.Get(0) != nil {
		v = args.Get(0).([]service.CategoryTreeNode)
	}
	return v, args.Error(1)
}

func (m *mockCategoryService) Update(ctx context.Context, id uint, name string, actorUserID uint) (*model.CategoryModel, error) {
	args := m.Called(ctx, id, name, actorUserID)
	c, _ := args.Get(0).(*model.CategoryModel)
	return c, args.Error(1)
}

func (m *mockCategoryService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	args := m.Called(ctx, id, actorUserID)
	return args.Error(0)
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

func TestCategoryHandler_List_OK(t *testing.T) {
	t.Parallel()

	svc := &mockCategoryService{}
	svc.On("List", mock.Anything, mock.AnythingOfType("request.CategoryListRequest")).Return(repository.CursorPage{
		Items: []model.CategoryModel{{ID: 1, Name: "A", Slug: "a", Lft: 1, Rgt: 2, Depth: 0}},
	}, nil).Once()

	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/categories", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestCategoryHandler_GetTree_OK(t *testing.T) {
	t.Parallel()

	svc := &mockCategoryService{}
	svc.On("GetTree", mock.Anything).Return([]service.CategoryTreeNode{
		{ID: 1, Name: "Root", Children: []service.CategoryTreeNode{}},
	}, nil).Once()

	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/categories/tree", h.GetTree)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/tree", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestCategoryHandler_GetSubtree_NotFound(t *testing.T) {
	t.Parallel()

	svc := &mockCategoryService{}
	svc.On("GetSubtree", mock.Anything, uint(1)).Return(nil, service.ErrCategoryNotFound).Once()

	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/categories/:id/subtree", h.GetSubtree)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/1/subtree", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	env := decodeEnvCat(t, rr)
	assert.Equal(t, "not found", env.Message)
	svc.AssertExpectations(t)
}

func TestCategoryHandler_CreateRoot_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &mockCategoryService{}
	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.POST("/api/v1/categories/root", h.CreateRoot)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/root", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	svc.AssertExpectations(t)
}

func TestCategoryHandler_CreateRoot_OK(t *testing.T) {
	t.Parallel()

	svc := &mockCategoryService{}
	svc.On("CreateRoot", mock.Anything, "Books", uint(1)).Return(&model.CategoryModel{ID: 1, Name: "Books", Slug: "books", Lft: 1, Rgt: 2, Depth: 0}, nil).Once()

	h := NewCategoryHandler(svc, nil)
	r := gin.New()
	r.POST("/api/v1/categories/root", withAuthRole("admin"), h.CreateRoot)

	body := `{"name":"Books"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/root", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	svc.AssertExpectations(t)
}
