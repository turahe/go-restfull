package auth

import (
	"net/http"
	"webapi/internal/app/user"
	"webapi/internal/db/model"
	"webapi/internal/helper/utils"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"
	"webapi/internal/logger"
	"webapi/pkg/email"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHttpHandler struct {
	app          user.UserApp
	otpService   *utils.OTPService
	emailService *email.EmailService
}

func NewAuthHttpHandler(app user.UserApp) *AuthHttpHandler {
	return &AuthHttpHandler{
		app:          app,
		otpService:   utils.NewOTPService(),
		emailService: email.NewEmailService(),
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body requests.AuthLoginRequest true "Login credentials"
// @Success 200 {object} response.CommonResponse{data=utils.TokenPair}
// @Failure 400 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /v1/auth/login [post]
func (h *AuthHttpHandler) Login(c *fiber.Ctx) error {
	var req requests.AuthLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Process the business logic
	userDto, err := h.app.Login(c.Context(), requests.AuthLoginRequest{
		UserName: req.UserName,
		Password: req.Password,
	})

	if err != nil {
		return err
	}

	// Generate JWT tokens
	tokenPair, err := utils.GenerateTokenPair(userDto.ID, userDto.UserName, userDto.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to generate tokens",
		})
	}

	return c.Status(http.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Login successful",
		Data:            tokenPair,
	})
}

// Register godoc
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body requests.AuthRegisterRequest true "User registration information"
// @Success 201 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/auth/register [post]
func (h *AuthHttpHandler) Register(c *fiber.Ctx) error {
	var req requests.AuthRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	validator := requests.XValidator{}
	errs := validator.Validate(&req)
	if len(errs) > 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"validation_errors": errs})
	}

	userModel := model.User{
		UserName: req.UserName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	}
	dto, err := h.app.CreateUser(c.Context(), userModel)
	if err != nil {
		return err
	}

	// Send welcome email after successful registration
	templateData := struct {
		Name  string
		Email string
	}{
		Name:  req.UserName,
		Email: req.Email,
	}
	h.emailService.SendEmailTemplate(
		req.Email,
		"Welcome to Our Platform",
		"pkg/template/email/welcome.html",
		templateData,
		true,
	)

	return c.Status(http.StatusCreated).JSON(response.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Registration successful",
		Data:            dto,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body requests.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.CommonResponse{data=utils.TokenPair}
// @Failure 400 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /v1/auth/refresh [post]
func (h *AuthHttpHandler) RefreshToken(c *fiber.Ctx) error {
	var req requests.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request body",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Refresh token is required",
		})
	}

	// Generate new token pair using refresh token
	tokenPair, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(response.CommonResponse{
			ResponseCode:    http.StatusUnauthorized,
			ResponseMessage: "Invalid refresh token",
			Data:            err.Error(),
		})
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Token refreshed successfully",
		Data:            tokenPair,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user (client should discard tokens)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /v1/auth/logout [post]
func (h *AuthHttpHandler) Logout(c *fiber.Ctx) error {
	// In a stateless JWT system, logout is handled client-side
	// The client should discard the tokens
	// Optionally, you could implement a blacklist for revoked tokens

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Logout successful",
	})
}

// ForgetPassword godoc
// @Summary Forget password (send OTP or reset link)
// @Description Send an OTP to phone or reset link to email for password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param data body requests.ForgetPasswordRequest true "Identity (email or phone)"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /v1/auth/forget-password [post]
func (h *AuthHttpHandler) ForgetPassword(c *fiber.Ctx) error {
	var req requests.ForgetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request body",
		})
	}

	if req.Identity == "" {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Identity is required",
		})
	}

	email, phone := utils.ParseIdentity(req.Identity)

	if email != "" {
		// Send reset link for email
		_, err := h.otpService.GenerateAndStoreResetLink(c.Context(), req.Identity, 0)
		if err != nil {
			logger.Log.Error("Failed to generate reset link", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{
				ResponseCode:    http.StatusInternalServerError,
				ResponseMessage: "Failed to generate reset link",
			})
		}

		return c.JSON(response.CommonResponse{
			ResponseCode:    http.StatusOK,
			ResponseMessage: "Reset link sent to email successfully",
		})
	} else {
		// Send OTP for phone
		otp, err := h.otpService.GenerateAndStoreOTP(c.Context(), req.Identity, 6, 0)
		if err != nil {
			logger.Log.Error("Failed to generate OTP", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{
				ResponseCode:    http.StatusInternalServerError,
				ResponseMessage: "Failed to generate OTP",
			})
		}

		// Simulate sending OTP to phone
		logger.Log.Info("[SIMULATED] Sent OTP to phone", zap.String("otp", otp), zap.String("phone", phone))

		return c.JSON(response.CommonResponse{
			ResponseCode:    http.StatusOK,
			ResponseMessage: "OTP sent to phone successfully",
		})
	}
}

// ValidateOTP godoc
// @Summary Validate OTP or reset link
// @Description Validate the OTP sent to phone or reset link sent to email
// @Tags auth
// @Accept json
// @Produce json
// @Param data body requests.ValidateOTPRequest true "Validation data (OTP for phone or token for email)"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /v1/auth/validate-otp [post]
func (h *AuthHttpHandler) ValidateOTP(c *fiber.Ctx) error {
	var req requests.ValidateOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request body",
		})
	}

	if req.Identity == "" {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Identity is required",
		})
	}

	email, phone := utils.ParseIdentity(req.Identity)

	if email != "" {
		// Validate reset link for email
		if req.Token == "" {
			return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
				ResponseCode:    http.StatusBadRequest,
				ResponseMessage: "Token is required for email validation",
			})
		}

		isValid, err := h.otpService.ValidateResetLink(c.Context(), req.Identity, req.Token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(response.CommonResponse{
				ResponseCode:    http.StatusUnauthorized,
				ResponseMessage: "Invalid or expired reset link",
			})
		}

		if !isValid {
			return c.Status(http.StatusUnauthorized).JSON(response.CommonResponse{
				ResponseCode:    http.StatusUnauthorized,
				ResponseMessage: "Invalid reset link",
			})
		}

		return c.JSON(response.CommonResponse{
			ResponseCode:    http.StatusOK,
			ResponseMessage: "Reset link validated successfully",
		})
	} else if phone != "" {
		// Validate OTP for phone
		if req.OTP == "" {
			return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
				ResponseCode:    http.StatusBadRequest,
				ResponseMessage: "OTP is required for phone validation",
			})
		}

		isValid, err := h.otpService.ValidateOTP(c.Context(), req.Identity, req.OTP)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(response.CommonResponse{
				ResponseCode:    http.StatusUnauthorized,
				ResponseMessage: "Invalid or expired OTP",
			})
		}

		if !isValid {
			return c.Status(http.StatusUnauthorized).JSON(response.CommonResponse{
				ResponseCode:    http.StatusUnauthorized,
				ResponseMessage: "Invalid OTP",
			})
		}

		return c.JSON(response.CommonResponse{
			ResponseCode:    http.StatusOK,
			ResponseMessage: "OTP validated successfully",
		})
	}

	return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{
		ResponseCode:    http.StatusBadRequest,
		ResponseMessage: "Invalid identity format",
	})
}
