package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-rest/internal/model"
	"go-rest/internal/service"
	"go-rest/internal/service/dto"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	args := m.Called(ctx, name, email, password)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}
func (m *mockAuthService) Login(ctx context.Context, email, password string, meta dto.LoginMeta) (dto.LoginResult, error) {
	args := m.Called(ctx, email, password, meta)
	return args.Get(0).(dto.LoginResult), args.Error(1)
}
func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string, meta dto.LoginMeta) (dto.RefreshResult, error) {
	args := m.Called(ctx, refreshToken, meta)
	return args.Get(0).(dto.RefreshResult), args.Error(1)
}
func (m *mockAuthService) Profile(ctx context.Context, userID uint) (dto.AuthUser, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(dto.AuthUser), args.Error(1)
}
func (m *mockAuthService) SetupTwoFA(ctx context.Context, userID uint, email string) (dto.TwoFactorSetupResult, error) {
	args := m.Called(ctx, userID, email)
	return args.Get(0).(dto.TwoFactorSetupResult), args.Error(1)
}
func (m *mockAuthService) EnableTwoFA(ctx context.Context, userID uint, code string) error {
	return m.Called(ctx, userID, code).Error(0)
}
func (m *mockAuthService) VerifyTwoFAChallenge(ctx context.Context, challengeID string, deviceID string, code string) (dto.LoginResult, error) {
	args := m.Called(ctx, challengeID, deviceID, code)
	return args.Get(0).(dto.LoginResult), args.Error(1)
}
func (m *mockAuthService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	return m.Called(ctx, userID, currentPassword, newPassword).Error(0)
}
func (m *mockAuthService) ChangeEmail(ctx context.Context, userID uint, currentPassword, newEmail string) error {
	return m.Called(ctx, userID, currentPassword, newEmail).Error(0)
}
func (m *mockAuthService) Logout(ctx context.Context, sessionID string, accessJTI string, accessExp time.Time, userID uint) error {
	return m.Called(ctx, sessionID, accessJTI, accessExp, userID).Error(0)
}
func (m *mockAuthService) Impersonate(ctx context.Context, impersonatorID uint, targetUserID uint, reason string, meta dto.LoginMeta) (dto.ImpersonationResult, error) {
	args := m.Called(ctx, impersonatorID, targetUserID, reason, meta)
	return args.Get(0).(dto.ImpersonationResult), args.Error(1)
}

func decodeEnv(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, rr.Body.String())
	}
	return env
}

func TestAuthHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		setupMock  func(s *mockAuthService)
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "invalid json",
			body:       "{",
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid request",
		},
		{
			name:       "validation error",
			body:       `{"name":"a","email":"not-email","password":"short"}`,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "validation failed",
		},
		{
			name: "email taken",
			body: `{"name":"abcd","email":"a@b.com","password":"12345678"}`,
			setupMock: func(s *mockAuthService) {
				s.On("Register", mock.Anything, "abcd", "a@b.com", "12345678").Return((*model.User)(nil), service.ErrEmailTaken).Once()
			},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "email already registered",
		},
		{
			name: "success",
			body: `{"name":"abcd","email":"a@b.com","password":"12345678"}`,
			setupMock: func(s *mockAuthService) {
				s.On("Register", mock.Anything, "abcd", "a@b.com", "12345678").Return(&model.User{ID: 1, Name: "abcd", Email: "a@b.com"}, nil).Once()
			},
			wantStatus: http.StatusCreated,
			wantMsg:    "Successfully registered user",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockAuthService{}
			if tc.setupMock != nil {
				tc.setupMock(svc)
			}
			h := NewAuthHandler(svc, nil)

			r := gin.New()
			r.POST("/api/v1/auth/register", h.Register)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			env := decodeEnv(t, rr)
			assert.Equal(t, tc.wantMsg, env.Message)
			svc.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMock  func(s *mockAuthService)
		body       string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "validation error",
			body:       `{"email":"bad","password":"12345678","deviceId":"dev1"}`,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "validation failed",
		},
		{
			name: "invalid credentials",
			body: `{"email":"a@b.com","password":"12345678","deviceId":"dev1"}`,
			setupMock: func(s *mockAuthService) {
				s.On("Login", mock.Anything, "a@b.com", "12345678", mock.AnythingOfType("dto.LoginMeta")).
					Return(dto.LoginResult{}, service.ErrInvalidCredentials).Once()
			},
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "invalid credentials",
		},
		{
			name: "success",
			body: `{"email":"a@b.com","password":"12345678","deviceId":"dev1"}`,
			setupMock: func(s *mockAuthService) {
				out := dto.LoginResult{
					TwoFactorRequired: false,
					AccessToken:       "a",
					RefreshToken:      "r",
					ExpiresAt:         time.Now(),
					SessionID:         "s",
					User:              dto.AuthUser{ID: 1, Email: "a@b.com"},
				}
				s.On("Login", mock.Anything, "a@b.com", "12345678", mock.AnythingOfType("dto.LoginMeta")).Return(out, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantMsg:    "ok",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockAuthService{}
			if tc.setupMock != nil {
				tc.setupMock(svc)
			}
			h := NewAuthHandler(svc, nil)

			r := gin.New()
			r.POST("/api/v1/auth/login", h.Login)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			env := decodeEnv(t, rr)
			assert.Equal(t, tc.wantMsg, env.Message)
			svc.AssertExpectations(t)
		})
	}
}

