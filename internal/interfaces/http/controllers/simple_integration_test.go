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
	"webapi/internal/infrastructure/adapters"
	"webapi/internal/interfaces/http/controllers"
	"webapi/internal/interfaces/http/requests"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestAuthController_SimpleIntegration demonstrates testing controllers with real dependencies
// This test uses mock repositories but real services to test the integration
func TestAuthController_SimpleIntegration(t *testing.T) {
	// Create mock repositories (in a real test, you'd use test database)
	mockUserRepo := &MockUserRepository{}

	// Create real services with mock repositories
	passwordService := adapters.NewBcryptPasswordService()
	emailService := adapters.NewSmtpEmailService(nil) // nil email client for testing

	// Create auth service with real dependencies
	authService := services.NewAuthService(
		mockUserRepo,
		passwordService,
		emailService,
	)

	// Create auth controller
	authController := controllers.NewAuthController(authService)

	// Create Fiber app
	app := fiber.New()
	app.Post("/register", authController.Register)

	// Test registration with valid data
	t.Run("successful registration", func(t *testing.T) {
		// Setup mock expectations
		mockUserRepo.On("ExistsByEmail", "test@example.com").Return(false, nil)
		mockUserRepo.On("ExistsByUsername", "testuser").Return(false, nil)
		mockUserRepo.On("ExistsByPhone", "+1234567890").Return(false, nil)
		mockUserRepo.On("Create", mock.Anything).Return(nil)

		// Create request body
		requestBody := requests.RegisterRequest{
			Username:        "testuser",
			Email:           "test@example.com",
			Phone:           "+1234567890",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert status code
		assert.Equal(t, 201, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Created", response["message"])
		assert.NotEmpty(t, response["data"])

		// Verify mock expectations
		mockUserRepo.AssertExpectations(t)
	})

	// Test registration with invalid data
	t.Run("registration with invalid email", func(t *testing.T) {
		// Create request body with invalid email
		requestBody := requests.RegisterRequest{
			Username:        "testuser2",
			Email:           "invalid-email",
			Phone:           "+1234567891",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert status code
		assert.Equal(t, 400, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["message"])
	})
}

// MockUserRepository is a mock implementation for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*entities.User, error) {
	args := m.Called(ctx, phone)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

// MockRoleRepository is a mock implementation for testing
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Role), args.Error(1)
}

func (m *MockRoleRepository) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(*entities.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Role), args.Error(1)
}

func (m *MockRoleRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
