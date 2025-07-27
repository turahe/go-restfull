package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"webapi/internal/application/services"
	"webapi/internal/domain/entities"
	"webapi/internal/interfaces/http/controllers"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/testutils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthController_Integration_Register(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create auth service with real dependencies
	authService := services.NewAuthService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create auth controller
	authController := controllers.NewAuthController(authService)

	// Create Fiber app
	app := fiber.New()
	app.Post("/register", authController.Register)

	tests := []struct {
		name           string
		requestBody    requests.RegisterRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful registration",
			requestBody: requests.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Phone:           "+1234567890",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			expectedStatus: 201,
			expectedError:  false,
		},
		{
			name: "registration with invalid email",
			requestBody: requests.RegisterRequest{
				Username:        "testuser2",
				Email:           "invalid-email",
				Phone:           "+1234567891",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name: "registration with password mismatch",
			requestBody: requests.RegisterRequest{
				Username:        "testuser3",
				Email:           "test3@example.com",
				Phone:           "+1234567892",
				Password:        "Password123!",
				ConfirmPassword: "DifferentPassword123!",
			},
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name: "registration with weak password",
			requestBody: requests.RegisterRequest{
				Username:        "testuser4",
				Email:           "test4@example.com",
				Phone:           "+1234567893",
				Password:        "weak",
				ConfirmPassword: "weak",
			},
			expectedStatus: 400,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
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
				assert.NotEmpty(t, response["message"])
			} else {
				assert.Equal(t, "Created", response["message"])
				assert.NotEmpty(t, response["data"])
			}
		})
	}
}

func TestAuthController_Integration_Login(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create auth service with real dependencies
	authService := services.NewAuthService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create auth controller
	authController := controllers.NewAuthController(authService)

	// Create Fiber app
	app := fiber.New()
	app.Post("/login", authController.Login)

	// Create a test user first
	user, err := entities.NewUser("logintest", "logintest@example.com", "+1234567890", "Password123!")
	require.NoError(t, err)

	// Hash password
	hashedPassword, err := setup.Container.PasswordService.HashPassword(user.Password)
	require.NoError(t, err)
	user.Password = hashedPassword

	// Verify email and phone
	user.VerifyEmail()
	user.VerifyPhone()

	// Save user to database
	err = setup.Container.UserRepository.Create(context.Background(), user)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    requests.LoginRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful login",
			requestBody: requests.LoginRequest{
				Username: "logintest",
				Password: "Password123!",
			},
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name: "invalid username",
			requestBody: requests.LoginRequest{
				Username: "nonexistent",
				Password: "Password123!",
			},
			expectedStatus: 401,
			expectedError:  true,
		},
		{
			name: "invalid password",
			requestBody: requests.LoginRequest{
				Username: "logintest",
				Password: "WrongPassword123!",
			},
			expectedStatus: 401,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
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
				assert.NotEmpty(t, response["message"])
			} else {
				assert.Equal(t, "OK", response["message"])
				assert.NotEmpty(t, response["data"])
			}
		})
	}
}

func TestAuthController_Integration_Refresh(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create auth service with real dependencies
	authService := services.NewAuthService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create auth controller
	authController := controllers.NewAuthController(authService)

	// Create Fiber app
	app := fiber.New()
	app.Post("/refresh", authController.Refresh)

	// Create a test user
	user, err := entities.NewUser("refreshtest", "refreshtest@example.com", "+1234567890", "Password123!")
	require.NoError(t, err)

	// Hash password
	hashedPassword, err := setup.Container.PasswordService.HashPassword(user.Password)
	require.NoError(t, err)
	user.Password = hashedPassword

	// Verify email and phone
	user.VerifyEmail()
	user.VerifyPhone()

	// Save user to database
	err = setup.Container.UserRepository.Create(context.Background(), user)
	require.NoError(t, err)

	// Generate refresh token using the auth service
	tokenPair, _, err := authService.LoginUser(context.Background(), user.UserName, "Password123!")
	require.NoError(t, err)

	tests := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful token refresh",
			refreshToken:   tokenPair.RefreshToken,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "invalid refresh token",
			refreshToken:   "invalid-token",
			expectedStatus: 401,
			expectedError:  true,
		},
		{
			name:           "empty refresh token",
			refreshToken:   "",
			expectedStatus: 400,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			requestBody := map[string]string{"refresh_token": tt.refreshToken}
			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
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
				assert.NotEmpty(t, response["message"])
			} else {
				assert.Equal(t, "OK", response["message"])
				assert.NotEmpty(t, response["data"])
			}
		})
	}
}
