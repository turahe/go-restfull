package middleware

import (
	"crypto/rsa"
	"errors"
	"strconv"
	"strings"

	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthClaims struct {
	UserID uint
	Email  string
	Name   string
}

const ctxAuthKey = "auth_claims"

func JWTAuth(publicKey *rsa.PublicKey, issuer string, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			response.Unauthorized(c, 4010101, "missing or invalid authorization header", "missing bearer token")
			c.Abort()
			return
		}

		parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return publicKey, nil
		}, jwt.WithIssuer(issuer), jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
		if err != nil || !parsed.Valid {
			log.Warn("jwt invalid", zap.Error(err))
			response.Unauthorized(c, 4010102, "invalid token", "invalid token")
			c.Abort()
			return
		}

		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, 4010103, "invalid token claims", "invalid claims")
			c.Abort()
			return
		}

		sub, ok := claims["sub"]
		if !ok {
			response.Unauthorized(c, 4010104, "invalid token subject", "missing sub")
			c.Abort()
			return
		}

		var userID uint64
		switch v := sub.(type) {
		case float64:
			userID = uint64(v)
		case string:
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				response.Unauthorized(c, 4010104, "invalid token subject", "bad sub")
				c.Abort()
				return
			}
			userID = n
		default:
			response.Unauthorized(c, 4010104, "invalid token subject", "bad sub type")
			c.Abort()
			return
		}

		ac := AuthClaims{
			UserID: uint(userID),
		}
		if v, ok := claims["email"].(string); ok {
			ac.Email = v
		}
		if v, ok := claims["name"].(string); ok {
			ac.Name = v
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

