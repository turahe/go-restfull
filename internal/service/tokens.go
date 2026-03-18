package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func newUUIDLike() (string, error) {
	// 16 random bytes hex-encoded (32 chars). Good enough for IDs in this local project.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashRefreshToken(token, pepper string) string {
	h := sha256.Sum256([]byte(pepper + ":" + token))
	return hex.EncodeToString(h[:])
}

