package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/utils"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService is a mock implementation of ports.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(ctx context.Context, username, email, phone, password string) (*utils.TokenPair, *entities.User, error) {
	args := m.Called(ctx, username, email, phone, password)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*entities.User), args.Error(2)
	}
	return args.Get(0).(*utils.TokenPair), args.Get(1).(*entities.User), args.Error(2)
}

func (m *MockAuthService) LoginUser(ctx context.Context, username, password string) (*utils.TokenPair, *entities.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*entities.User), args.Error(2)
	}
	return args.Get(0).(*utils.TokenPair), args.Get(1).(*entities.User), args.Error(2)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*utils.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.TokenPair), args.Error(1)
}

func (m *MockAuthService) LogoutUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) ForgetPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthService) ResetPassword(ctx context.Context, email, otp, password string) error {
	args := m.Called(ctx, email, otp, password)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*utils.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.TokenClaims), args.Error(1)
}

func setupAuthControllerTest(t *testing.T) (*fiber.App, *MockAuthService, *controllers.AuthController) {
	// Create mock auth service
	mockAuthService := new(MockAuthService)

	// Create auth controller
	authController := controllers.NewAuthController(mockAuthService)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Register routes
	app.Post("/auth/register", authController.Register)
	app.Post("/auth/login", authController.Login)
	app.Post("/auth/refresh", authController.Refresh)
	app.Post("/auth/logout", authController.Logout)
	app.Post("/auth/forget-password", authController.ForgetPassword)
	app.Post("/auth/reset-password", authController.ResetPassword)

	return app, mockAuthService, authController
}

func TestAuthController_Register(t *testing.T) {
	app, mockAuthService, _ := setupAuthControllerTest(t)

	tests := []struct {
		name           string
		requestBody    requests.RegisterRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful registration",
			requestBody: requests.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Phone:           "+1234567890",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockSetup: func() {
				user, _ := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
				tokenPair := &utils.TokenPair{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					ExpiresIn:    3600,
				}
				mockAuthService.On("RegisterUser", mock.Anything, "testuser", "test@example.com", "+1234567890", "password123").
					Return(tokenPair, user, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "invalid request body",
			requestBody: requests.RegisterRequest{
				Username:        "",
				Email:           "invalid-email",
				Phone:           "123",
				Password:        "short",
				ConfirmPassword: "short",
			},
			mockSetup:      func() {}, // No mock setup needed for validation error
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "service error",
			requestBody: requests.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Phone:           "+1234567890",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockSetup: func() {
				mockAuthService.On("RegisterUser", mock.Anything, "testuser", "test@example.com", "+1234567890", "password123").
					Return(nil, (*entities.User)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuthService.ExpectedCalls = nil
			mockAuthService.Calls = nil

			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Make request
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, "error", response["status"])
				assert.NotEmpty(t, response["message"])
			} else {
				assert.Equal(t, "success", response["status"])
				assert.NotNil(t, response["data"])
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	app, mockAuthService, _ := setupAuthControllerTest(t)

	tests := []struct {
		name           string
		requestBody    requests.LoginRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful login",
			requestBody: requests.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				user, _ := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
				tokenPair := &utils.TokenPair{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					ExpiresIn:    3600,
				}
				mockAuthService.On("LoginUser", mock.Anything, "testuser", "password123").
					Return(tokenPair, user, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid credentials",
			requestBody: requests.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockAuthService.On("LoginUser", mock.Anything, "testuser", "wrongpassword").
					Return(nil, (*entities.User)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuthService.ExpectedCalls = nil
			mockAuthService.Calls = nil

			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Make request
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, "error", response["status"])
			} else {
				assert.Equal(t, "success", response["status"])
				assert.NotNil(t, response["data"])
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_Refresh(t *testing.T) {
	app, mockAuthService, _ := setupAuthControllerTest(t)

	tests := []struct {
		name           string
		requestBody    requests.RefreshTokenRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful token refresh",
			requestBody: requests.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			mockSetup: func() {
				tokenPair := &utils.TokenPair{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					ExpiresIn:    3600,
				}
				mockAuthService.On("RefreshToken", mock.Anything, "valid_refresh_token").
					Return(tokenPair, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid refresh token",
			requestBody: requests.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			mockSetup: func() {
				mockAuthService.On("RefreshToken", mock.Anything, "invalid_token").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuthService.ExpectedCalls = nil
			mockAuthService.Calls = nil

			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Make request
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, "error", response["status"])
			} else {
				assert.Equal(t, "success", response["status"])
				assert.NotNil(t, response["data"])
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_ForgetPassword(t *testing.T) {
	app, mockAuthService, _ := setupAuthControllerTest(t)

	tests := []struct {
		name           string
		requestBody    requests.ForgetPasswordRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful forget password",
			requestBody: requests.ForgetPasswordRequest{
				Email: "test@example.com",
			},
			mockSetup: func() {
				mockAuthService.On("ForgetPassword", mock.Anything, "test@example.com").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "invalid email",
			requestBody: requests.ForgetPasswordRequest{
				Email: "invalid-email",
			},
			mockSetup:      func() {}, // No mock setup needed for validation error
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuthService.ExpectedCalls = nil
			mockAuthService.Calls = nil

			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/forget-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Make request
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, "error", response["status"])
			} else {
				assert.Equal(t, "success", response["status"])
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_ResetPassword(t *testing.T) {
	app, mockAuthService, _ := setupAuthControllerTest(t)

	tests := []struct {
		name           string
		requestBody    requests.ResetPasswordRequest
		mockSetup      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful password reset",
			requestBody: requests.ResetPasswordRequest{
				Email:           "test@example.com",
				OTP:             "123456",
				Password:        "newpassword123",
				ConfirmPassword: "newpassword123",
			},
			mockSetup: func() {
				mockAuthService.On("ResetPassword", mock.Anything, "test@example.com", "123456", "newpassword123").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "password mismatch",
			requestBody: requests.ResetPasswordRequest{
				Email:           "test@example.com",
				OTP:             "123456",
				Password:        "newpassword123",
				ConfirmPassword: "differentpassword",
			},
			mockSetup:      func() {}, // No mock setup needed for validation error
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuthService.ExpectedCalls = nil
			mockAuthService.Calls = nil

			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Make request
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, "error", response["status"])
			} else {
				assert.Equal(t, "success", response["status"])
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}
