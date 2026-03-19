package service

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"go-rest/internal/service/dto"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKeys map[string]*rsa.PublicKey // kid -> key

	issuer   string
	audience string
	keyID    string
	log      *zap.Logger
}

func NewJWTService(privateKeyPath, publicKeyPath, issuer, audience, keyID string, log *zap.Logger) (*JWTService, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Error("failed to read private key", zap.Error(err))
		return nil, err
	}
	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Error("failed to read public key", zap.Error(err))
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		log.Error("failed to parse private key", zap.Error(err))
		return nil, err
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		log.Error("failed to parse public key", zap.Error(err))
		return nil, err
	}
	if issuer == "" || audience == "" || keyID == "" {
		log.Error("JWT issuer, audience, keyID are required")
		return nil, errors.New("JWT issuer, audience, keyID are required")
	}
	return &JWTService{
		privateKey: privKey,
		publicKeys: map[string]*rsa.PublicKey{keyID: pubKey},
		issuer:     issuer,
		audience:   audience,
		keyID:      keyID,
		log:        log,
	}, nil
}

func (s *JWTService) IssueAccessToken(cl dto.AccessClaims) (string, error) {
	if cl.ID == "" {
		s.log.Error("jti is required")
		return "", errors.New("jti is required")
	}
	if cl.Subject == "" {
		s.log.Error("sub is required")
		return "", errors.New("sub is required")
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, cl)
	tok.Header["kid"] = s.keyID
	return tok.SignedString(s.privateKey)
}

func (s *JWTService) ParseAndValidateAccess(tokenStr string) (*dto.AccessClaims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			s.log.Error("unexpected signing method")
			return nil, errors.New("unexpected signing method")
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			s.log.Error("missing kid")
			return nil, errors.New("missing kid")
		}
		pub, ok := s.publicKeys[kid]
		if !ok {
			s.log.Error("unknown kid", zap.String("kid", kid))
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}
		return pub, nil
	}

	var claims dto.AccessClaims
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		s.log.Error("failed to parse token", zap.Error(err))
		return nil, err
	}
	if !parsed.Valid {
		s.log.Error("invalid token")
		return nil, errors.New("invalid token")
	}

	now := time.Now()
	if claims.IssuedAt == nil || claims.NotBefore == nil || claims.ExpiresAt == nil {
		s.log.Error("missing required time claims")
		return nil, errors.New("missing required time claims")
	}
	if claims.IssuedAt.Time.After(now.Add(2 * time.Minute)) {
		s.log.Error("iat is in the future")
		return nil, errors.New("iat is in the future")
	}
	if claims.NotBefore.Time.After(now.Add(30 * time.Second)) {
		s.log.Error("nbf is in the future")
		return nil, errors.New("nbf is in the future")
	}
	if claims.ID == "" {
		s.log.Error("missing jti")
		return nil, errors.New("missing jti")
	}
	if claims.SessionID == "" || claims.DeviceID == "" {
		s.log.Error("missing session_id or device_id")
		return nil, errors.New("missing session_id or device_id")
	}
	if claims.UserID == 0 {
		s.log.Error("missing user_id")
		return nil, errors.New("missing user_id")
	}
	return &claims, nil
}

func (s *JWTService) DefaultRegistered(subject string, ttl time.Duration) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   subject,
		Audience:  jwt.ClaimStrings{s.audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
}
