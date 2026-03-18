package middleware

import (
	"strings"

	"go-rest/internal/repository"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthClaims struct {
	UserID      uint
	Role        string
	Permissions []string
	SessionID   string
	DeviceID    string
	JTI         string

	Impersonation       bool
	ImpersonatorID      *uint
	ImpersonatedUserID  *uint
	ImpersonationReason string
}

const ctxAuthKey = "auth_claims"

func JWTAuth(jwtSvc *service.JWTService, authRepo *repository.AuthRepository, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "missing or invalid authorization header", "missing bearer token")
			c.Abort()
			return
		}

		claims, err := jwtSvc.ParseAndValidateAccess(tokenStr)
		if err != nil {
			log.Warn("jwt invalid", zap.Error(err))
			response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodeAuth, response.CaseCodeInvalidToken), "invalid token", "invalid token")
			c.Abort()
			return
		}

		revoked, err := authRepo.IsJTIRevoked(c.Request.Context(), claims.ID)
		if err != nil {
			log.Warn("jti check failed", zap.Error(err))
			response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodeAuth, response.CaseCodeInvalidToken), "invalid token", "invalid token")
			c.Abort()
			return
		}
		if revoked {
			response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodeAuth, response.CaseCodeInvalidToken), "invalid token", "revoked token")
			c.Abort()
			return
		}

		active, err := authRepo.SessionActive(c.Request.Context(), claims.SessionID)
		if err != nil {
			log.Warn("session check failed", zap.Error(err))
			response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodeAuth, response.CaseCodeInvalidToken), "invalid token", "invalid token")
			c.Abort()
			return
		}
		if !active {
			response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodeAuth, response.CaseCodeInvalidToken), "invalid token", "session revoked")
			c.Abort()
			return
		}

		ac := AuthClaims{
			UserID:      claims.UserID,
			Role:        claims.Role,
			Permissions: claims.Permissions,
			SessionID:   claims.SessionID,
			DeviceID:    claims.DeviceID,
			JTI:         claims.ID,
			Impersonation: claims.Impersonation,
			ImpersonatorID: claims.ImpersonatorID,
			ImpersonatedUserID: claims.ImpersonatedUserID,
			ImpersonationReason: claims.ImpersonationReason,
		}

		c.Set(ctxAuthKey, ac)
		c.Next()
	}
}

func GetAuth(c *gin.Context) (AuthClaims, bool) {
	v, ok := c.Get(ctxAuthKey)
	if !ok {
		return AuthClaims{}, false
	}
	claims, ok := v.(AuthClaims)
	return claims, ok
}

func bearerToken(h string) (string, bool) {
	h = strings.TrimSpace(h)
	if h == "" {
		return "", false
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	tok := strings.TrimSpace(parts[1])
	if tok == "" {
		return "", false
	}
	return tok, true
}

