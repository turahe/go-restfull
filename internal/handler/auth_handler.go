package handler

import (
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	auth *service.AuthService
	log  *zap.Logger
}

func NewAuthHandler(auth *service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{auth: auth, log: log}
}

type registerReq struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=190"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// Register godoc
// @Summary      Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      registerReq  true  "Register payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, 4000101, "invalid request", err.Error())
		return
	}

	u, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if err == service.ErrEmailTaken {
			response.BadRequest(c, 4000102, "email already registered", "email taken")
			return
		}
		h.log.Error("register failed", zap.Error(err))
		response.InternalServerError(c, 5000201, "internal error", "register failed")
		return
	}

	response.Created(c, 2010101, "registered", gin.H{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	})
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email,max=190"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// Login godoc
// @Summary      Login and get JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginReq  true  "Login payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, 4000103, "invalid request", err.Error())
		return
	}

	token, u, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, 4010201, "invalid credentials", "invalid credentials")
			return
		}
		h.log.Error("login failed", zap.Error(err))
		response.InternalServerError(c, 5000202, "internal error", "login failed")
		return
	}

	response.OK(c, 2000101, "ok", gin.H{
		"token": token,
		"user": gin.H{
			"id":    u.ID,
			"name":  u.Name,
			"email": u.Email,
		},
	})
}

