package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/turahe/go-restfull/pkg/ids"

	"go.uber.org/zap"
)

func hashRefreshToken(token, pepper string, log *zap.Logger) (string, error) {
	if token == "" || pepper == "" {
		log.Error("token or pepper is empty")
		return "", errors.New("token or pepper is empty")
	}
	h := sha256.Sum256([]byte(pepper + ":" + token))
	return hex.EncodeToString(h[:]), nil
}

func newUUIDLike(log *zap.Logger) (string, error) {
	id, err := ids.New()
	if err != nil {
		log.Error("failed to generate new uuid", zap.Error(err))
		return "", errors.New("failed to generate new uuid")
	}
	return id, nil
}
