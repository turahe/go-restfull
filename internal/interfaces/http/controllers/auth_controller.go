package controllers

import (
	"net/http"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	authService ports.AuthService
	userRepo    repositories.UserRepository
}

// NewAuthController creates a new auth controller instance
func NewAuthController(authService ports.AuthService, userRepo repositories.UserRepository) *AuthController {
	return &AuthController{
		authService: authService,
		userRepo:    userRepo,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param user body requests.RegisterRequest true "User registration request"
// @Success 201 {object} responses.AuthResourceResponse "User registered successfully"
// @Failure 400 {object} responses.ValidationErrorResponse "Bad request - Validation errors"
// @Failure 409 {object} responses.ValidationErrorResponse "Conflict - User already exists"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var req requests.RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Errors:  responses.ValidationErrors{},
		})
	}

	// Validate request with database uniqueness check
	validationErrors, err := req.ValidateWithDatabase(ctx.Context(), c.userRepo)
	if err != nil {
		// Check if it's a uniqueness error (409 Conflict)
		if validationErrors.HasErrors() {
			// Check for uniqueness errors
			hasUniquenessError := false
			for _, err := range validationErrors.GetErrors() {
				if err.Field == "username" || err.Field == "email" || err.Field == "phone" {
					hasUniquenessError = true
					break
				}
			}

			if hasUniquenessError {
				return ctx.Status(http.StatusConflict).JSON(responses.ValidationErrorResponse{
					Status:  "error",
					Message: "The given data was invalid.",
					Errors:  validationErrors.GetErrors(),
				})
			}

			// Other validation errors (400 Bad Request)
			return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
				Status:  "error",
				Message: "The given data was invalid.",
				Errors:  validationErrors.GetErrors(),
			})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to validate request",
		})
	}

	// Get normalized phone number
	normalizedPhone, err := req.GetNormalizedPhone()
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid phone number format",
		})
	}

	// Register user with normalized phone number
	tokenPair, _, err := c.authService.RegisterUser(ctx.Context(), req.Username, req.Email, normalizedPhone, req.Password)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Create auth response
	// Note: We need to get the user entity to create a proper auth response
	// For now, we'll return a token response since we don't have the user entity here
	response := responses.NewTokenResourceResponse(tokenPair)
	response.ResponseCode = http.StatusCreated
	response.ResponseMessage = "User registered successfully"
	return ctx.Status(http.StatusCreated).JSON(response)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body requests.LoginRequest true "Login credentials"
// @Success 200 {object} responses.TokenResourceResponse "Login successful"
// @Failure 400 {object} responses.ValidationErrorResponse "Bad request - Validation errors"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - Invalid credentials"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req requests.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Errors:  responses.ValidationErrors{},
		})
	}

	// Validate request
	validationErrors, err := req.Validate()
	if err != nil {
		// If validation failed, return the validation errors
		if validationErrors != nil && validationErrors.HasErrors() {
			return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
				Status:  "error",
				Message: "The given data was invalid.",
				Errors:  validationErrors.GetErrors(),
			})
		}
		// If no validation errors but general error, return generic error
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Login user
	tokenPair, _, err := c.authService.LoginUser(ctx.Context(), req.Identity, req.Password)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
	}

	// Create auth response
	return ctx.JSON(responses.NewTokenResource(tokenPair))
}

// Refresh godoc
// @Summary Refresh access token
// @Description Refresh an access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param token body requests.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} responses.TokenResourceResponse "Token refreshed successfully"
// @Failure 400 {object} responses.ValidationErrorResponse "Bad request - Validation errors"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - Invalid refresh token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	var req requests.RefreshTokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Errors:  responses.ValidationErrors{},
		})
	}

	// Validate request
	validationErrors, err := req.Validate()
	if err != nil {
		// If validation failed, return the validation errors
		if validationErrors != nil && validationErrors.HasErrors() {
			return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
				Status:  "error",
				Message: "The given data was invalid.",
				Errors:  validationErrors.GetErrors(),
			})
		}
		// If no validation errors but general error, return generic error
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Refresh token
	tokenPair, err := c.authService.RefreshToken(ctx.Context(), req.RefreshToken)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Create token response
	return ctx.JSON(responses.NewTokenResourceResponse(tokenPair))
}

// Logout godoc
// @Summary User logout
// @Description Logout user (client should discard tokens)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} responses.ErrorResponse "Logout successful"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// Get user ID from context (set by JWT middleware)
	userID := ctx.Locals("user_id").(string)

	// Logout user
	err := c.authService.LogoutUser(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to logout",
		})
	}

	return ctx.JSON(responses.ErrorResponse{
		Status:  "success",
		Message: "Logout successful",
	})
}

// ForgetPassword godoc
// @Summary Request password reset
// @Description Send a password reset email with OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.ForgetPasswordRequest true "Password reset request"
// @Success 200 {object} responses.ErrorResponse "Password reset email sent"
// @Failure 400 {object} responses.ValidationErrorResponse "Bad request - Validation errors"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/forget-password [post]
func (c *AuthController) ForgetPassword(ctx *fiber.Ctx) error {
	var req requests.ForgetPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Errors:  responses.ValidationErrors{},
		})
	}

	// Validate request
	validationErrors, err := req.Validate()
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "The given data was invalid.",
			Errors:  validationErrors.GetErrors(),
		})
	}

	// if identifier as username or email, send reset password link to email
	// if identifier as phone number, send OTP
	err = c.authService.ForgetPassword(ctx.Context(), req.Identifier)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to send password reset email",
		})
	}

	return ctx.JSON(responses.ErrorResponse{
		Status:  "success",
		Message: "Password reset email sent",
	})
}

// ResetPassword godoc
// @Summary Reset password with OTP
// @Description Reset password using email and OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.ResetPasswordRequest true "Password reset with OTP request"
// @Success 200 {object} responses.ErrorResponse "Password reset successful"
// @Failure 400 {object} responses.ValidationErrorResponse "Bad request - Validation errors"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - Invalid OTP"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *fiber.Ctx) error {
	var req requests.ResetPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Errors:  responses.ValidationErrors{},
		})
	}

	// Validate request
	validationErrors, err := req.Validate()
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ValidationErrorResponse{
			Status:  "error",
			Message: "The given data was invalid.",
			Errors:  validationErrors.GetErrors(),
		})
	}

	// Reset password
	err = c.authService.ResetPassword(ctx.Context(), req.Email, req.OTP, req.Password)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid email or OTP",
		})
	}

	return ctx.JSON(responses.ErrorResponse{
		Status:  "success",
		Message: "Password reset successful",
	})
}
