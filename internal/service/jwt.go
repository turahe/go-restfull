package service

import (
	"crypto/rsa"
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	ttlMinutes int
}

func NewJWTManager(privateKeyPath, publicKeyPath, issuer string, ttlMinutes int) (*JWTManager, error) {
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
	if issuer == "" {
		return nil, errors.New("JWT issuer is required")
	}
	if ttlMinutes <= 0 {
		return nil, errors.New("JWT ttlMinutes must be > 0")
	}
	return &JWTManager{
		privateKey: privKey,
		publicKey:  pubKey,
		issuer:     issuer,
		ttlMinutes: ttlMinutes,
	}, nil
}

func (m *JWTManager) PublicKey() *rsa.PublicKey {
	return m.publicKey
}

