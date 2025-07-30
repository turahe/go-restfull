package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/pkg/exception"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommentService implements ports.CommentService for testing
type MockCommentService struct {
	mock.Mock
}

func (m *MockCommentService) CreateComment(ctx context.Context, content string, postID, userID uuid.UUID, parentID *uuid.UUID, status string) (*entities.Comment, error) {
	args := m.Called(ctx, content, postID, userID, parentID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetCommentByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetCommentsByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	args := m.Called(ctx, postID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetCommentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetCommentReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	args := m.Called(ctx, parentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetAllComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetApprovedComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentService) GetPendingComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentService) UpdateComment(ctx context.Context, id uuid.UUID, content, status string) (*entities.Comment, error) {
	args := m.Called(ctx, id, content, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Comment), args.Error(1)
}

func (m *MockCommentService) DeleteComment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentService) ApproveComment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentService) RejectComment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentService) GetCommentCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentService) GetCommentCountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentService) GetCommentCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentService) GetPendingCommentCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Helper function to create a test comment
func createTestComment() *entities.Comment {
	now := time.Now()
	return &entities.Comment{
		ID:        uuid.New(),
		Content:   "Test comment content",
		PostID:    uuid.New(),
		UserID:    uuid.New(),
		ParentID:  nil,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper function to create a test app with authenticated user
func createTestAppWithAuth(controller *controllers.CommentController) *fiber.App {
	app := fiber.New()

	// Add middleware to set user_id in context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})

	// Register routes
	app.Get("/comments", controller.GetComments)
	app.Get("/comments/:id", controller.GetCommentByID)
	app.Post("/comments", controller.CreateComment)
	app.Put("/comments/:id", controller.UpdateComment)
	app.Delete("/comments/:id", controller.DeleteComment)
	app.Put("/comments/:id/approve", controller.ApproveComment)
	app.Put("/comments/:id/reject", controller.RejectComment)

	return app
}

func TestGetComments_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	// Mock service response
	testComments := []*entities.Comment{createTestComment()}
	mockService.On("GetApprovedComments", mock.Anything, 10, 0).Return(testComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetComments_ByPostID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	postID := uuid.New()
	testComments := []*entities.Comment{createTestComment()}
	mockService.On("GetCommentsByPostID", mock.Anything, postID, 10, 0).Return(testComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments?post_id="+postID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetComments_ByUserID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	userID := uuid.New()
	testComments := []*entities.Comment{createTestComment()}
	mockService.On("GetCommentsByUserID", mock.Anything, userID, 10, 0).Return(testComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments?user_id="+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetComments_ByParentID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	parentID := uuid.New()
	testComments := []*entities.Comment{createTestComment()}
	mockService.On("GetCommentReplies", mock.Anything, parentID, 10, 0).Return(testComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments?parent_id="+parentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetComments_ApprovedStatus(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	testComments := []*entities.Comment{createTestComment()}
	mockService.On("GetApprovedComments", mock.Anything, 10, 0).Return(testComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments?status=approved", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetComments_PendingStatus(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	testComments := []*entities.Comment{createTestComment()}
	mockService.On("GetPendingComments", mock.Anything, 10, 0).Return(testComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments?status=pending", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetComments_ServiceError(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	mockService.On("GetApprovedComments", mock.Anything, 10, 0).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/comments", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCommentByID_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	testComment := createTestComment()
	testComment.ID = commentID

	mockService.On("GetCommentByID", mock.Anything, commentID).Return(testComment, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments/"+commentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCommentByID_InvalidUUID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	req := httptest.NewRequest(http.MethodGet, "/comments/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetCommentByID_NotFound(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("GetCommentByID", mock.Anything, commentID).Return(nil, exception.DataNotFoundError)

	req := httptest.NewRequest(http.MethodGet, "/comments/"+commentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCommentByID_ServiceError(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("GetCommentByID", mock.Anything, commentID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/comments/"+commentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCreateComment_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	requestBody := requests.CreateCommentRequest{
		Content: "Test comment content",
		PostID:  uuid.New(),
	}

	testComment := createTestComment()
	mockService.On("CreateComment", mock.Anything, requestBody.Content, requestBody.PostID, mock.Anything, (*uuid.UUID)(nil), "pending").Return(testComment, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCreateComment_WithParentID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	parentID := uuid.New()
	requestBody := requests.CreateCommentRequest{
		Content:  "Test reply content",
		PostID:   uuid.New(),
		ParentID: &parentID,
	}

	testComment := createTestComment()
	mockService.On("CreateComment", mock.Anything, requestBody.Content, requestBody.PostID, mock.Anything, &parentID, "pending").Return(testComment, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCreateComment_InvalidRequestBody(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateComment_ServiceError(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	requestBody := requests.CreateCommentRequest{
		Content: "Test comment content",
		PostID:  uuid.New(),
	}

	mockService.On("CreateComment", mock.Anything, requestBody.Content, requestBody.PostID, mock.Anything, (*uuid.UUID)(nil), "pending").Return(nil, errors.New("database error"))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateComment_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	requestBody := requests.UpdateCommentRequest{
		Content: "Updated comment content",
	}

	testComment := createTestComment()
	testComment.ID = commentID
	testComment.Content = requestBody.Content

	mockService.On("UpdateComment", mock.Anything, commentID, requestBody.Content, "").Return(testComment, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateComment_InvalidUUID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	requestBody := requests.UpdateCommentRequest{
		Content: "Updated comment content",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/comments/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateComment_NotFound(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	requestBody := requests.UpdateCommentRequest{
		Content: "Updated comment content",
	}

	mockService.On("UpdateComment", mock.Anything, commentID, requestBody.Content, "").Return(nil, exception.DataNotFoundError)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteComment_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("DeleteComment", mock.Anything, commentID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/comments/"+commentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteComment_InvalidUUID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	req := httptest.NewRequest(http.MethodDelete, "/comments/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteComment_NotFound(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("DeleteComment", mock.Anything, commentID).Return(exception.DataNotFoundError)

	req := httptest.NewRequest(http.MethodDelete, "/comments/"+commentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestApproveComment_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("ApproveComment", mock.Anything, commentID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String()+"/approve", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestApproveComment_InvalidUUID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	req := httptest.NewRequest(http.MethodPut, "/comments/invalid-uuid/approve", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestApproveComment_NotFound(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("ApproveComment", mock.Anything, commentID).Return(exception.DataNotFoundError)

	req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String()+"/approve", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestRejectComment_Success(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("RejectComment", mock.Anything, commentID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String()+"/reject", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestRejectComment_InvalidUUID(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	req := httptest.NewRequest(http.MethodPut, "/comments/invalid-uuid/reject", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRejectComment_NotFound(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := createTestAppWithAuth(controller)

	commentID := uuid.New()
	mockService.On("RejectComment", mock.Anything, commentID).Return(exception.DataNotFoundError)

	req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String()+"/reject", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCommentController_Unauthenticated(t *testing.T) {
	mockService := new(MockCommentService)
	controller := controllers.NewCommentController(mockService)
	app := fiber.New()

	// Register routes without auth middleware
	app.Post("/comments", controller.CreateComment)
	app.Put("/comments/:id", controller.UpdateComment)
	app.Delete("/comments/:id", controller.DeleteComment)
	app.Put("/comments/:id/approve", controller.ApproveComment)
	app.Put("/comments/:id/reject", controller.RejectComment)

	// Test CreateComment without authentication
	requestBody := requests.CreateCommentRequest{
		Content: "Test comment content",
		PostID:  uuid.New(),
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCommentQueryParams_SetDefaults(t *testing.T) {
	params := &requests.CommentQueryParams{}
	params.SetDefaults()

	assert.Equal(t, 10, params.Limit)
	assert.Equal(t, 0, params.Offset)
	assert.Equal(t, "approved", params.Status)

	// Test with existing values
	params = &requests.CommentQueryParams{
		Limit:  20,
		Offset: 5,
		Status: "pending",
	}
	params.SetDefaults()

	assert.Equal(t, 20, params.Limit)
	assert.Equal(t, 5, params.Offset)
	assert.Equal(t, "pending", params.Status)
}
