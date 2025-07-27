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

func TestUserController_Integration_CreateUser(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create user service with real dependencies
	userService := services.NewUserService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create user controller
	userController := controllers.NewUserController(userService, setup.Container.PaginationService)

	// Create Fiber app
	app := fiber.New()
	app.Post("/users", userController.CreateUser)

	tests := []struct {
		name           string
		requestBody    requests.CreateUserRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful user creation",
			requestBody: requests.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Phone:    "+1234567890",
				Password: "Password123!",
			},
			expectedStatus: 201,
			expectedError:  false,
		},
		{
			name: "user creation with invalid email",
			requestBody: requests.CreateUserRequest{
				Username: "testuser2",
				Email:    "invalid-email",
				Phone:    "+1234567891",
				Password: "Password123!",
			},
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name: "user creation with weak password",
			requestBody: requests.CreateUserRequest{
				Username: "testuser3",
				Email:    "test3@example.com",
				Phone:    "+1234567892",
				Password: "weak",
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
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
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

func TestUserController_Integration_GetUserByID(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create user service with real dependencies
	userService := services.NewUserService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create user controller
	userController := controllers.NewUserController(userService, setup.Container.PaginationService)

	// Create Fiber app
	app := fiber.New()
	app.Get("/users/:id", userController.GetUserByID)

	// Create a test user first
	user, err := entities.NewUser("getusertest", "getusertest@example.com", "+1234567890", "Password123!")
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
		userID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful get user by ID",
			userID:         user.ID.String(),
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "user not found",
			userID:         "00000000-0000-0000-0000-000000000000",
			expectedStatus: 404,
			expectedError:  true,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			expectedStatus: 400,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)

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

func TestUserController_Integration_GetAllUsers(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create user service with real dependencies
	userService := services.NewUserService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create user controller
	userController := controllers.NewUserController(userService, setup.Container.PaginationService)

	// Create Fiber app
	app := fiber.New()
	app.Get("/users", userController.GetUsers)

	// Create test users
	for i := 1; i <= 3; i++ {
		user, err := entities.NewUser(
			"getalluser"+string(rune(i)),
			"getalluser"+string(rune(i))+"@example.com",
			"+123456789"+string(rune(i)),
			"Password123!",
		)
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
	}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "get all users",
			query:          "",
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "get users with pagination",
			query:          "?page=1&limit=2",
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "get users with search",
			query:          "?search=getalluser1",
			expectedStatus: 200,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)

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

func TestUserController_Integration_UpdateUser(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create user service with real dependencies
	userService := services.NewUserService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create user controller
	userController := controllers.NewUserController(userService, setup.Container.PaginationService)

	// Create Fiber app
	app := fiber.New()
	app.Put("/users/:id", userController.UpdateUser)

	// Create a test user first
	user, err := entities.NewUser("updateusertest", "updateusertest@example.com", "+1234567890", "Password123!")
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
		userID         string
		requestBody    requests.UpdateUserRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "successful user update",
			userID: user.ID.String(),
			requestBody: requests.UpdateUserRequest{
				Username: "updateduser",
				Email:    "updated@example.com",
				Phone:    "+1234567891",
			},
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:   "update with invalid email",
			userID: user.ID.String(),
			requestBody: requests.UpdateUserRequest{
				Username: "updateduser2",
				Email:    "invalid-email",
				Phone:    "+1234567892",
			},
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name:           "user not found",
			userID:         "00000000-0000-0000-0000-000000000000",
			requestBody:    requests.UpdateUserRequest{},
			expectedStatus: 404,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewBuffer(body))
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
				assert.Equal(t, "Updated", response["message"])
				assert.NotEmpty(t, response["data"])
			}
		})
	}
}

func TestUserController_Integration_DeleteUser(t *testing.T) {
	// Skip if no database available
	testutils.SkipIfNoDatabase(t)

	// Setup test environment
	setup := testutils.SetupTestContainer(t)
	defer setup.Cleanup()

	// Create user service with real dependencies
	userService := services.NewUserService(
		setup.Container.UserRepository,
		setup.Container.PasswordService,
		setup.Container.EmailService,
	)

	// Create user controller
	userController := controllers.NewUserController(userService, setup.Container.PaginationService)

	// Create Fiber app
	app := fiber.New()
	app.Delete("/users/:id", userController.DeleteUser)

	// Create a test user first
	user, err := entities.NewUser("deleteusertest", "deleteusertest@example.com", "+1234567890", "Password123!")
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
		userID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful user deletion",
			userID:         user.ID.String(),
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "user not found",
			userID:         "00000000-0000-0000-0000-000000000000",
			expectedStatus: 404,
			expectedError:  true,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			expectedStatus: 400,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)

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
				assert.Equal(t, "Deleted", response["message"])
			}
		})
	}
}
