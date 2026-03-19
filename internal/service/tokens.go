package service

import (
	"crypto/sha256"
	"encoding/hex"

	"go-rest/pkg/ids"
)

func hashRefreshToken(token, pepper string) string {
	h := sha256.Sum256([]byte(pepper + ":" + token))
	return hex.EncodeToString(h[:])
}

func newUUIDLike() (string, error) {
	return ids.New()
}

