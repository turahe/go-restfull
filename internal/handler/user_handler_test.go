package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/turahe/go-restfull/internal/domain/entities"
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

func (m *mockUserService) Create(ctx context.Context, req request.CreateUserRequest) (*service.UserCreateOutcome, error) {
	args := m.Called(ctx, req)
	out, _ := args.Get(0).(*service.UserCreateOutcome)
	return out, args.Error(1)
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

func TestUserHandler_Create(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth(entities.RoleAdmin))
	r.POST("/api/v1/users", h.Create)

	body := `{"name":"New","email":"new@example.com","password":"password1","confirmPassword":"password1"}`
	svc.On("Create", mock.Anything, mock.MatchedBy(func(req request.CreateUserRequest) bool {
		return req.Name == "New" && req.Email == "new@example.com" && req.Password == "password1" && req.ConfirmPassword == "password1"
	})).Return(&service.UserCreateOutcome{
		User:   &model.User{ID: 7, Name: "New", Email: "new@example.com"},
		RoleID: 3,
	}, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	env := decodeEnvelope(t, rr)
	assert.Equal(t, "Successfully created user", env.Message)
	data, ok := env.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, float64(3), data["roleId"])
	svc.AssertExpectations(t)
}

func TestUserHandler_Create_Forbidden(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth(entities.RoleUser))
	r.POST("/api/v1/users", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(`{"name":"New","email":"a@b.com","password":"password1","confirmPassword":"password1"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestUserHandler_Create_EmailTaken(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth(entities.RoleAdmin))
	r.POST("/api/v1/users", h.Create)

	svc.On("Create", mock.Anything, mock.Anything).Return((*service.UserCreateOutcome)(nil), service.ErrEmailTaken).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(`{"name":"New","email":"a@b.com","password":"password1","confirmPassword":"password1"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvelope(t, rr)
	assert.Equal(t, "email already registered", env.Message)
	svc.AssertExpectations(t)
}

func TestUserHandler_List(t *testing.T) {
	t.Parallel()
	svc := &mockUserService{}
	h := NewUserHandler(svc, nil)

	r := gin.New()
	r.Use(withAuth(entities.RoleAdmin))
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
	r.Use(withAuth(entities.RoleUser))
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
	r.Use(withAuth(entities.RoleAdmin))
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
	r.Use(withAuth(entities.RoleAdmin))
	r.GET("/api/v1/users/:id", h.GetByID)

	svc.On("GetByID", mock.Anything, uint(10)).Return((*model.User)(nil), service.ErrUserNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/10", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}
