package post

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"webapi/internal/app/post"
	"webapi/internal/db/model"
	"webapi/internal/http/requests"

	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPostApp struct {
	mock.Mock
	post.PostApp
}

func (m *MockPostApp) CreatePostWithTags(ctx context.Context, p *model.Post, tags []uuid.UUID) error {
	args := m.Called(ctx, p, tags)
	return args.Error(0)
}
func (m *MockPostApp) UpdatePostWithTags(ctx context.Context, p *model.Post, tags []uuid.UUID) error {
	args := m.Called(ctx, p, tags)
	return args.Error(0)
}
func (m *MockPostApp) DeletePost(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockPostApp) GetPostByIDWithContents(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Post), args.Error(1)
}
func (m *MockPostApp) GetAllPostsWithContents(ctx context.Context) ([]*model.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Post), args.Error(1)
}

func TestCreatePost_ValidationError(t *testing.T) {
	app := fiber.New()
	mockApp := new(MockPostApp)
	h := NewPostHttpHandler(mockApp)
	app.Post("/posts", h.CreatePost)

	// Invalid request (missing required fields)
	reqBody := `{}`
	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreatePost_Success(t *testing.T) {
	app := fiber.New()
	mockApp := new(MockPostApp)
	h := NewPostHttpHandler(mockApp)
	app.Post("/posts", func(c *fiber.Ctx) error {
		// Set user_id in context
		userID := uuid.New()
		c.Locals("user_id", userID)
		return h.CreatePost(c)
	})

	req := requests.CreatePostRequest{
		Slug:           "test-post",
		Title:          "Test Post",
		Type:           "article",
		Tags:           []string{},
		Description:    "desc",
		Language:       "en",
		RecordOrdering: 1,
	}
	mockApp.On("CreatePostWithTags", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(req)
	reqHttp := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
	reqHttp.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(reqHttp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestGetPostByID_NotFound(t *testing.T) {
	app := fiber.New()
	mockApp := new(MockPostApp)
	h := NewPostHttpHandler(mockApp)
	app.Get("/posts/:id", h.GetPostByID)

	mockApp.On("GetPostByIDWithContents", mock.Anything, mock.Anything).Return(&model.Post{}, errors.New("not found"))

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/posts/"+id, nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetAllPosts_Success(t *testing.T) {
	app := fiber.New()
	mockApp := new(MockPostApp)
	h := NewPostHttpHandler(mockApp)
	app.Get("/posts", h.GetAllPosts)

	mockApp.On("GetAllPostsWithContents", mock.Anything).Return([]*model.Post{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdatePost_ValidationError(t *testing.T) {
	app := fiber.New()
	mockApp := new(MockPostApp)
	h := NewPostHttpHandler(mockApp)
	app.Put("/posts/:id", h.UpdatePost)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/posts/"+id, bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeletePost_Success(t *testing.T) {
	app := fiber.New()
	mockApp := new(MockPostApp)
	h := NewPostHttpHandler(mockApp)
	app.Delete("/posts/:id", h.DeletePost)

	mockApp.On("DeletePost", mock.Anything, mock.Anything).Return(nil)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/posts/"+id, nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
