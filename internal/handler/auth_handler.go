package handler

import (
	"context"
	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/model"
	"go-rest/internal/service"
	"go-rest/internal/service/dto"
	"go-rest/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	BaseHandler
	auth AuthService
}

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string, meta dto.LoginMeta) (dto.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string, meta dto.LoginMeta) (dto.RefreshResult, error)
	Profile(ctx context.Context, userID uint) (dto.AuthUser, error)
	SetupTwoFA(ctx context.Context, userID uint, email string) (dto.TwoFactorSetupResult, error)
	EnableTwoFA(ctx context.Context, userID uint, code string) error
	VerifyTwoFAChallenge(ctx context.Context, challengeID string, deviceID string, code string) (dto.LoginResult, error)
	ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error
	ChangeEmail(ctx context.Context, userID uint, currentPassword, newEmail string) error
	Logout(ctx context.Context, sessionID string, accessJTI string, accessExp time.Time, userID uint) error
	Impersonate(ctx context.Context, impersonatorID uint, targetUserID uint, reason string, meta dto.LoginMeta) (dto.ImpersonationResult, error)
}

func NewAuthHandler(auth AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{BaseHandler: BaseHandler{Log: log}, auth: auth}
}

// Register godoc
// @Summary      Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.RegisterRequest  true  "Register payload"
// @Success      201   {object}  response.Created
// @Failure      400   {object}  response.BadRequest
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}

	u, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if err == service.ErrEmailTaken {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeAuth, response.CaseCodeDuplicateEntry), "email already registered", "email taken")
			return
		}
		h.internalError(c, response.ServiceCodeAuth, err, "register failed")
		return
	}

	response.Created(c,
		response.BuildResponseCode(http.StatusCreated, response.ServiceCodeAuth, response.CaseCodeCreated),
		"Successfully registered user",
		gin.H{
			"id":    u.ID,
			"name":  u.Name,
			"email": u.Email,
		})
}

// Login godoc
// @Summary      Login and get JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.LoginRequest  true  "Login payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}

	res, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, dto.LoginMeta{
		DeviceID:  req.DeviceID,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	})
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeInvalidCredentials), "invalid credentials", "invalid credentials")
			return
		}
		h.internalError(c, response.ServiceCodeAuth, err, "login failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeLoginSuccess), "Successfully logged in", res)
}

// TwoFASetup godoc
// @Summary      Initialize TOTP 2FA for current user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  response.OK
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/2fa/setup [post]
func (h *AuthHandler) TwoFASetup(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	profile, err := h.auth.Profile(c.Request.Context(), auth.UserID)
	if err != nil {
		h.internalError(c, response.ServiceCodeAuth, err, "2fa setup failed")
		return
	}
	res, err := h.auth.SetupTwoFA(c.Request.Context(), auth.UserID, profile.Email)
	if err != nil {
		h.internalError(c, response.ServiceCodeAuth, err, "2fa setup failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "Successfully initialized 2FA", res)
}

// TwoFAEnable godoc
// @Summary      Enable TOTP 2FA for current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.TwoFAEnableRequest  true  "Enable 2FA payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/2fa/enable [post]
func (h *AuthHandler) TwoFAEnable(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.TwoFAEnableRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}
	if err := h.auth.EnableTwoFA(c.Request.Context(), auth.UserID, req.Code); err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeAuth, response.CaseCodeInvalidValue), "invalid 2fa code", err.Error())
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "Successfully enabled 2FA", nil)
}

// TwoFAVerify godoc
// @Summary      Verify 2FA challenge and issue tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.TwoFAVerifyRequest  true  "Verify 2FA payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/2fa/verify [post]
func (h *AuthHandler) TwoFAVerify(c *gin.Context) {
	var req request.TwoFAVerifyRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}

	res, err := h.auth.VerifyTwoFAChallenge(c.Request.Context(), req.ChallengeID, req.DeviceID, req.Code)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeAuth, response.CaseCodeInvalidValue), "invalid 2fa verification", err.Error())
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "Successfully verified 2FA challenge", res)
}

// Refresh godoc
// @Summary      Rotate refresh token and get new access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.RefreshRequest  true  "Refresh payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}
	res, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken, dto.LoginMeta{
		DeviceID:  req.DeviceID,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	})
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeInvalidCredentials), "invalid credentials", "invalid refresh token")
			return
		}
		h.internalError(c, response.ServiceCodeAuth, err, "refresh failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "Successfully refreshed token", res)
}

// Profile godoc
// @Summary      Get current user profile
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  response.OK
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	profile, err := h.auth.Profile(c.Request.Context(), auth.UserID)
	if err != nil {
		h.internalError(c, response.ServiceCodeAuth, err, "profile failed")
		return
	}
	response.OK(c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeRetrieved),
		"Successfully retrieved profile",
		profile)
}

// ChangePassword godoc
// @Summary      Change password for current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ChangePasswordRequest  true  "Change password payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/password/change [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	var req request.ChangePasswordRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}

	if err := h.auth.ChangePassword(c.Request.Context(), auth.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == service.ErrInvalidCurrentPass {
			response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeInvalidCredentials), "invalid password", "invalid current password")
			return
		}
		h.internalError(c, response.ServiceCodeAuth, err, "change password failed")
		return
	}

	_ = h.auth.Logout(c.Request.Context(), auth.SessionID, "", time.Time{}, auth.UserID)
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeUpdated), "Successfully changed password", nil)
}

// ChangeEmail godoc
// @Summary      Change email for current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ChangeEmailRequest  true  "Change email payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/email/change [post]
func (h *AuthHandler) ChangeEmail(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	var req request.ChangeEmailRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}

	if err := h.auth.ChangeEmail(c.Request.Context(), auth.UserID, req.CurrentPassword, req.NewEmail); err != nil {
		if err == service.ErrInvalidCurrentPass {
			response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeInvalidCredentials), "invalid password", "invalid current password")
			return
		}
		if err == service.ErrEmailTaken {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeAuth, response.CaseCodeDuplicateEntry), "email already registered", "email taken")
			return
		}
		h.internalError(c, response.ServiceCodeAuth, err, "change email failed")
		return
	}

	_ = h.auth.Logout(c.Request.Context(), auth.SessionID, "", time.Time{}, auth.UserID)
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeUpdated), "Successfully changed email", nil)
}

// Impersonate godoc
// @Summary      Admin/support impersonate a user (short-lived)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ImpersonateRequest  true  "Impersonate payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      403   {object}  response.Forbidden
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/auth/impersonate [post]
func (h *AuthHandler) Impersonate(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.ImpersonateRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}
	res, err := h.auth.Impersonate(c.Request.Context(), auth.UserID, req.UserID, req.Reason, dto.LoginMeta{
		DeviceID:  req.DeviceID,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	})
	if err != nil {
		if err.Error() == "forbidden" {
			response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeAuth, response.CaseCodePermissionDenied), "forbidden", "not allowed")
			return
		}
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeAuth, response.CaseCodeInvalidValue), "invalid request", err.Error())
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "Successfully impersonated user", res)
}
