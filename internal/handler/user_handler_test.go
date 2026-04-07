package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) List(ctx context.Context, req request.UserListRequest) (repository.CursorPage, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(repository.CursorPage), args.Error(1)
}

func (m *mockUserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func withAuth(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_claims", middleware.AuthClaims{Role: role, UserID: 1})
		c.Next()
	}
}

func decodeEnvelope(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func TestUserHandler_List(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth("admin"))
	r.GET("/api/v1/users", h.List)

	svc.On("List", mock.Anything, mock.Anything).Return(repository.CursorPage{
		Items:      []model.User{{ID: 1}},
		NextCursor: nil,
		PrevCursor: nil,
	}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	env := decodeEnvelope(t, rr)
	assert.Equal(t, "Successfully retrieved users", env.Message)
	svc.AssertExpectations(t)
}

func TestUserHandler_List_Forbidden(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth("user"))
	r.GET("/api/v1/users", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth("admin"))
	r.GET("/api/v1/users/:id", h.GetByID)

	svc.On("GetByID", mock.Anything, uint(10)).Return(&model.User{ID: 10}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/10", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	env := decodeEnvelope(t, rr)
	assert.Equal(t, "Successfully retrieved user", env.Message)
	svc.AssertExpectations(t)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth("admin"))
	r.GET("/api/v1/users/:id", h.GetByID)

	svc.On("GetByID", mock.Anything, uint(10)).Return((*model.User)(nil), errors.New("user not found")).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/10", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// handler maps only service.ErrUserNotFound to 404; generic error -> 500
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
