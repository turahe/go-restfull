package handler

import (
	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/service"
	svcresp "go-rest/internal/service/response"
	"go-rest/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	BaseHandler
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{BaseHandler: BaseHandler{Log: log}, auth: auth}
}

// Register godoc
// @Summary      Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.RegisterRequest  true  "Register payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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

	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeAuth, response.CaseCodeCreated), "registered", gin.H{
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
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}

	res, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, svcresp.LoginMeta{
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

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeLoginSuccess), "ok", res)
}

// TwoFASetup godoc
// @Summary      Initialize TOTP 2FA for current user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "ok", res)
}

// TwoFAEnable godoc
// @Summary      Enable TOTP 2FA for current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.TwoFAEnableRequest  true  "Enable 2FA payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "enabled", nil)
}

// TwoFAVerify godoc
// @Summary      Verify 2FA challenge and issue tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.TwoFAVerifyRequest  true  "Verify 2FA payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "ok", res)
}

// Refresh godoc
// @Summary      Rotate refresh token and get new access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.RefreshRequest  true  "Refresh payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshRequest
	if !h.bindJSON(c, response.ServiceCodeAuth, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeAuth, req) {
		return
	}
	res, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken, svcresp.LoginMeta{
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "ok", res)
}

// Profile godoc
// @Summary      Get current user profile
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeRetrieved), "ok", profile)
}

// ChangePassword godoc
// @Summary      Change password for current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ChangePasswordRequest  true  "Change password payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeUpdated), "password changed", nil)
}

// ChangeEmail godoc
// @Summary      Change email for current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ChangeEmailRequest  true  "Change email payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeUpdated), "email changed", nil)
}

// Impersonate godoc
// @Summary      Admin/support impersonate a user (short-lived)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ImpersonateRequest  true  "Impersonate payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	res, err := h.auth.Impersonate(c.Request.Context(), auth.UserID, req.UserID, req.Reason, svcresp.LoginMeta{
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
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeSuccess), "ok", res)
}
