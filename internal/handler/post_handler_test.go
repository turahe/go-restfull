package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/internal/model"
	"go-rest/internal/repository"
	"go-rest/internal/handler/request"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostService struct {
	mock.Mock
}

func (m *mockPostService) List(ctx context.Context, req request.PostListRequest) (repository.CursorPage, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(repository.CursorPage), args.Error(1)
}
func (m *mockPostService) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	args := m.Called(ctx, slug)
	p, _ := args.Get(0).(*model.Post)
	return p, args.Error(1)
}

func (m *mockPostService) Create(ctx context.Context, userID uint, req request.CreatePostRequest) (*model.Post, error) {
	args := m.Called(ctx, userID, req)
	p, _ := args.Get(0).(*model.Post)
	return p, args.Error(1)
}

func (m *mockPostService) Update(ctx context.Context, id uint, actorUserID uint, req request.UpdatePostRequest) (*model.Post, error) {
	args := m.Called(ctx, id, actorUserID, req)
	p, _ := args.Get(0).(*model.Post)
	return p, args.Error(1)
}
func (m *mockPostService) Delete(ctx context.Context, id uint, actorUserID uint) error {
	return m.Called(ctx, id, actorUserID).Error(0)
}

func decodeEnvelopePost(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func TestPostHandler_List_InvalidCursor(t *testing.T) {
	t.Parallel()

	svc := &mockPostService{}
	h := NewPostHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/posts", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?limit=bad", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	env := decodeEnvelopePost(t, rr)
	assert.Equal(t, "invalid request", env.Message)
	svc.AssertExpectations(t)
}

func TestPostHandler_List_Success_Defaults(t *testing.T) {
	t.Parallel()

	svc := &mockPostService{}
	h := NewPostHandler(svc, nil)
	r := gin.New()
	r.GET("/api/v1/posts", h.List)

	page := repository.CursorPage{
		Items:      []model.Post{{ID: 1}},
		NextCursor: nil,
		PrevCursor: nil,
	}
	svc.On("List", mock.Anything, mock.Anything).Return(page, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	env := decodeEnvelopePost(t, rr)
	assert.Equal(t, "Successfully retrieved posts", env.Message)
	svc.AssertExpectations(t)
}

func TestPostHandler_GetBySlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		slug       string
		setupMock  func(s *mockPostService)
		wantStatus int
		wantMsg    string
	}{
		{
			name: "not found",
			slug: "x",
			setupMock: func(s *mockPostService) {
				s.On("GetBySlug", mock.Anything, "x").Return((*model.Post)(nil), service.ErrPostNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantMsg:    "not found",
		},
		{
			name: "invalid slug maps to bad request",
			slug: "%20",
			setupMock: func(s *mockPostService) {
				s.On("GetBySlug", mock.Anything, " ").Return((*model.Post)(nil), service.ErrInvalidSlug).Once()
			},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid request",
		},
		{
			name: "success",
			slug: "hello",
			setupMock: func(s *mockPostService) {
				s.On("GetBySlug", mock.Anything, "hello").Return(&model.Post{ID: 1, Slug: "hello"}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully retrieved post by slug",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockPostService{}
			tc.setupMock(svc)
			h := NewPostHandler(svc, nil)
			r := gin.New()
			r.GET("/api/v1/posts/slug/:slug", h.GetBySlug)

			target := "/api/v1/posts/slug/" + tc.slug
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			env := decodeEnvelopePost(t, rr)
			assert.Equal(t, tc.wantMsg, env.Message)
			svc.AssertExpectations(t)
		})
	}
}

