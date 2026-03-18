package handler

import (
	"go-rest/internal/handler/request"
	"go-rest/internal/service"
	"go-rest/pkg/response"
	"net/http"

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

	token, u, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeInvalidCredentials), "invalid credentials", "invalid credentials")
			return
		}
		h.internalError(c, response.ServiceCodeAuth, err, "login failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeAuth, response.CaseCodeLoginSuccess), "ok", gin.H{
		"token": token,
		"user": gin.H{
			"id":    u.ID,
			"name":  u.Name,
			"email": u.Email,
		},
	})
}
