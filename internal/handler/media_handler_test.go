package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/usecase"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMediaService struct{ mock.Mock }

func (m *mockMediaService) Upload(ctx context.Context, actorUserID uint, fh *multipart.FileHeader, parentID *uint) (*model.Media, error) {
	args := m.Called(ctx, actorUserID, fh, parentID)
	med, _ := args.Get(0).(*model.Media)
	return med, args.Error(1)
}
func (m *mockMediaService) CreateFolderRoot(ctx context.Context, actorUserID uint, name string) (*model.Media, error) {
	args := m.Called(ctx, actorUserID, name)
	med, _ := args.Get(0).(*model.Media)
	return med, args.Error(1)
}
func (m *mockMediaService) CreateFolderChild(ctx context.Context, actorUserID uint, parentID uint, name string) (*model.Media, error) {
	args := m.Called(ctx, actorUserID, parentID, name)
	med, _ := args.Get(0).(*model.Media)
	return med, args.Error(1)
}
func (m *mockMediaService) GetTree(ctx context.Context, actorUserID uint) ([]usecase.MediaTreeNode, error) {
	args := m.Called(ctx, actorUserID)
	return args.Get(0).([]usecase.MediaTreeNode), args.Error(1)
}
func (m *mockMediaService) GetSubtree(ctx context.Context, actorUserID uint, mediaID uint) ([]usecase.MediaTreeNode, error) {
	args := m.Called(ctx, actorUserID, mediaID)
	return args.Get(0).([]usecase.MediaTreeNode), args.Error(1)
}
func (m *mockMediaService) Update(ctx context.Context, actorUserID uint, id uint, name string) (*model.Media, error) {
	args := m.Called(ctx, actorUserID, id, name)
	med, _ := args.Get(0).(*model.Media)
	return med, args.Error(1)
}
func (m *mockMediaService) List(ctx context.Context, actorUserID uint, req request.MediaListRequest) (repository.CursorPage, error) {
	args := m.Called(ctx, actorUserID, req)
	return args.Get(0).(repository.CursorPage), args.Error(1)
}
func (m *mockMediaService) GetByID(ctx context.Context, actorUserID, id uint) (*model.Media, error) {
	args := m.Called(ctx, actorUserID, id)
	med, _ := args.Get(0).(*model.Media)
	return med, args.Error(1)
}
func (m *mockMediaService) Delete(ctx context.Context, actorUserID, id uint) error {
	return m.Called(ctx, actorUserID, id).Error(0)
}
func (m *mockMediaService) PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectKey, expiry)
	return args.String(0), args.Error(1)
}

func decodeEnvMedia(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func withAuthUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_claims", middleware.AuthClaims{Role: "user", UserID: 1, SessionID: "s"})
		c.Next()
	}
}

func TestMediaHandler_List_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &mockMediaService{}
	h := NewMediaHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/media", h.ListMedia)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	svc.AssertExpectations(t)
}

func TestMediaHandler_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	svc := &mockMediaService{}
	svc.On("GetByID", mock.Anything, uint(1), uint(10)).Return((*model.Media)(nil), usecase.ErrMediaNotFound).Once()

	h := NewMediaHandler(svc, nil)
	r := gin.New()
	r.Use(withAuthUser())
	r.GET("/api/v1/media/:id", h.GetMediaByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/10", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	env := decodeEnvMedia(t, rr)
	assert.Equal(t, "not found", env.Message)
	svc.AssertExpectations(t)
}

func TestMediaHandler_Upload_MissingFile(t *testing.T) {
	t.Parallel()

	svc := &mockMediaService{}
	h := NewMediaHandler(svc, nil)
	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/media", h.UploadMedia)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvMedia(t, rr)
	assert.Equal(t, "invalid request", env.Message)
	svc.AssertExpectations(t)
}
