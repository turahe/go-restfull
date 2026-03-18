package service

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKeys map[string]*rsa.PublicKey // kid -> key

	issuer   string
	audience string
	keyID    string
}

type AccessClaims struct {
	jwt.RegisteredClaims

	UserID      uint     `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	DeviceID    string   `json:"device_id"`

	Impersonation           bool   `json:"impersonation,omitempty"`
	ImpersonatedUserID      *uint  `json:"impersonated_user_id,omitempty"`
	ImpersonatorID          *uint  `json:"impersonator_id,omitempty"`
	ImpersonationReason     string `json:"impersonation_reason,omitempty"`
}

func NewJWTService(privateKeyPath, publicKeyPath, issuer, audience, keyID string) (*JWTService, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, err
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, err
	}
	if issuer == "" || audience == "" || keyID == "" {
		return nil, errors.New("JWT issuer, audience, keyID are required")
	}
	return &JWTService{
		privateKey: privKey,
		publicKeys: map[string]*rsa.PublicKey{keyID: pubKey},
		issuer:     issuer,
		audience:   audience,
		keyID:      keyID,
	}, nil
}

func (s *JWTService) IssueAccessToken(cl AccessClaims) (string, error) {
	if cl.ID == "" {
		return "", errors.New("jti is required")
	}
	if cl.Subject == "" {
		return "", errors.New("sub is required")
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, cl)
	tok.Header["kid"] = s.keyID
	return tok.SignedString(s.privateKey)
}

func (s *JWTService) ParseAndValidateAccess(tokenStr string) (*AccessClaims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing kid")
		}
		pub, ok := s.publicKeys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}
		return pub, nil
	}

	var claims AccessClaims
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	now := time.Now()
	if claims.IssuedAt == nil || claims.NotBefore == nil || claims.ExpiresAt == nil {
		return nil, errors.New("missing required time claims")
	}
	if claims.IssuedAt.Time.After(now.Add(2 * time.Minute)) {
		return nil, errors.New("iat is in the future")
	}
	if claims.NotBefore.Time.After(now.Add(30 * time.Second)) {
		return nil, errors.New("nbf is in the future")
	}
	if claims.ID == "" {
		return nil, errors.New("missing jti")
	}
	if claims.SessionID == "" || claims.DeviceID == "" {
		return nil, errors.New("missing session_id or device_id")
	}
	if claims.UserID == 0 {
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

