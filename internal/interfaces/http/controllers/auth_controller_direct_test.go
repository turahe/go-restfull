package controllers_test

import (
	"context"
	"testing"

	"webapi/internal/domain/entities"
	"webapi/internal/helper/utils"
	"webapi/internal/interfaces/http/controllers"
	"webapi/internal/interfaces/http/requests"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
)

// MockAuthService is a mock implementation of ports.AuthService
type MockAuthServiceDirect struct {
	mock.Mock
}

func (m *MockAuthServiceDirect) RegisterUser(ctx context.Context, username, email, phone, password string) (*utils.TokenPair, *entities.User, error) {
	args := m.Called(ctx, username, email, phone, password)
	return args.Get(0).(*utils.TokenPair), args.Get(1).(*entities.User), args.Error(2)
}

func (m *MockAuthServiceDirect) LoginUser(ctx context.Context, username, password string) (*utils.TokenPair, *entities.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*utils.TokenPair), args.Get(1).(*entities.User), args.Error(2)
}

func (m *MockAuthServiceDirect) RefreshToken(ctx context.Context, refreshToken string) (*utils.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*utils.TokenPair), args.Error(1)
}

func (m *MockAuthServiceDirect) LogoutUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthServiceDirect) ForgetPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthServiceDirect) ResetPassword(ctx context.Context, email, otp, password string) error {
	args := m.Called(ctx, email, otp, password)
	return args.Error(0)
}

func (m *MockAuthServiceDirect) ValidateToken(ctx context.Context, token string) (*utils.TokenClaims, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*utils.TokenClaims), args.Error(1)
}

func setupAuthControllerDirectTest(t *testing.T) (*MockAuthServiceDirect, *controllers.AuthController) {
	// Create mock auth service
	mockAuthService := new(MockAuthServiceDirect)

	// Create auth controller
	authController := controllers.NewAuthController(mockAuthService)

	return mockAuthService, authController
}

func TestAuthController_Register_Direct(t *testing.T) {
	mockAuthService, authController := setupAuthControllerDirectTest(t)

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
			expectedStatus: 201,
			expectedError:  false,
		},
		{
			name: "registration with invalid data",
			requestBody: requests.RegisterRequest{
				Username:        "",
				Email:           "invalid-email",
				Phone:           "",
				Password:        "short",
				ConfirmPassword: "different",
			},
			mockSetup: func() {
				// No mock setup needed for validation failure
			},
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name: "service error during registration",
			requestBody: requests.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Phone:           "+1234567890",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockSetup: func() {
				mockAuthService.On("RegisterUser", mock.Anything, "testuser", "test@example.com", "+1234567890", "password123").
					Return((*utils.TokenPair)(nil), (*entities.User)(nil), assert.AnError)
			},
			expectedStatus: 500,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create Fiber context
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set request body
			ctx.Request().SetBody([]byte(`{
				"username": "` + tt.requestBody.Username + `",
				"email": "` + tt.requestBody.Email + `",
				"phone": "` + tt.requestBody.Phone + `",
				"password": "` + tt.requestBody.Password + `",
				"confirm_password": "` + tt.requestBody.ConfirmPassword + `"
			}`))

			// Call controller method
			err := authController.Register(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_Login_Direct(t *testing.T) {
	mockAuthService, authController := setupAuthControllerDirectTest(t)

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
			expectedStatus: 200,
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
			expectedStatus: 500,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create Fiber context
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set request body
			ctx.Request().SetBody([]byte(`{
				"username": "` + tt.requestBody.Username + `",
				"password": "` + tt.requestBody.Password + `"
			}`))

			// Call controller method
			err := authController.Login(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_Refresh_Direct(t *testing.T) {
	mockAuthService, authController := setupAuthControllerDirectTest(t)

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
			expectedStatus: 200,
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
			expectedStatus: 500,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create Fiber context
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set request body
			ctx.Request().SetBody([]byte(`{
				"refresh_token": "` + tt.requestBody.RefreshToken + `"
			}`))

			// Call controller method
			err := authController.Refresh(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_ForgetPassword_Direct(t *testing.T) {
	mockAuthService, authController := setupAuthControllerDirectTest(t)

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
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name: "invalid email",
			requestBody: requests.ForgetPasswordRequest{
				Email: "invalid-email",
			},
			mockSetup: func() {
				// No mock setup needed for validation error
			},
			expectedStatus: 400,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create Fiber context
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set request body
			ctx.Request().SetBody([]byte(`{
				"email": "` + tt.requestBody.Email + `"
			}`))

			// Call controller method
			err := authController.ForgetPassword(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_ResetPassword_Direct(t *testing.T) {
	mockAuthService, authController := setupAuthControllerDirectTest(t)

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
			expectedStatus: 200,
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
			mockSetup: func() {
				// No mock setup needed for validation error
			},
			expectedStatus: 400,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create Fiber context
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set request body
			ctx.Request().SetBody([]byte(`{
				"email": "` + tt.requestBody.Email + `",
				"otp": "` + tt.requestBody.OTP + `",
				"password": "` + tt.requestBody.Password + `",
				"confirm_password": "` + tt.requestBody.ConfirmPassword + `"
			}`))

			// Call controller method
			err := authController.ResetPassword(ctx)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}
