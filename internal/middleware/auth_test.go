package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/internal/service/dto"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantTok string
		wantOK  bool
	}{
		{"empty", "", "", false},
		{"no bearer", "Basic xyz", "", false},
		{"bearer empty", "Bearer ", "", false},
		{"bearer with token", "Bearer abc123", "abc123", true},
		{"bearer case insensitive", "BEARER tok", "tok", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, ok := bearerToken(tt.header)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantTok, tok)
		})
	}
}

func TestGetAuth(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, ok := GetAuth(c)
	assert.False(t, ok)

	ac := AuthClaims{UserID: 1, Role: "admin"}
	c.Set(ctxAuthKey, ac)
	got, ok := GetAuth(c)
	assert.True(t, ok)
	assert.Equal(t, uint(1), got.UserID)
	assert.Equal(t, "admin", got.Role)
}

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	t.Run("missing or invalid authorization header", func(t *testing.T) {
		jwtSvc, authRepo := stubAuthDeps(t)
		r := gin.New()
		r.Use(JWTAuth(jwtSvc, authRepo, log))
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		jwtSvc, authRepo := stubAuthDeps(t)
		r := gin.New()
		r.Use(JWTAuth(jwtSvc, authRepo, log))
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid token sets claims", func(t *testing.T) {
		jwtSvc, authRepo := stubAuthDeps(t)
		sessionID := uuid.New().String()
		ctx := context.Background()
		require.NoError(t, authRepo.CreateSession(ctx, &model.AuthSession{
			ID:         sessionID,
			UserID:     99,
			DeviceID:   "dev1",
			IPAddress:  "127.0.0.1",
			UserAgent:  "test",
			LastSeenAt: time.Now(),
		}))

		claims := dto.AccessClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "jti-1",
				Subject:   "99",
				Issuer:    "test",
				Audience:  jwt.ClaimStrings{"test"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			UserID:    99,
			Role:      "user",
			SessionID: sessionID,
			DeviceID:  "dev1",
		}
		tok, err := jwtSvc.IssueAccessToken(claims)
		require.NoError(t, err)

		r := gin.New()
		r.Use(JWTAuth(jwtSvc, authRepo, log))
		r.GET("/", func(c *gin.Context) {
			ac, ok := GetAuth(c)
			require.True(t, ok)
			assert.Equal(t, uint(99), ac.UserID)
			assert.Equal(t, "user", ac.Role)
			assert.Equal(t, sessionID, ac.SessionID)
			c.String(200, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}

func stubAuthDeps(t *testing.T) (*service.JWTService, *repository.AuthRepository) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pub := &priv.PublicKey
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	jwtSvc, err := service.NewJWTService(string(privPEM), string(pubPEM), "test", "test", "k1", zap.NewNop())
	require.NoError(t, err)

	db := openAuthTestDB(t)
	authRepo := repository.NewAuthRepository(db, zap.NewNop())
	return jwtSvc, authRepo
}

func openAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:auth_middleware?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{}))
	return db
}
