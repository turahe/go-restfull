package ids

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a random 16-byte hex string (32 chars).
func New() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

