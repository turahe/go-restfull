package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service/dto"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TwoFactorService struct {
	repo   *repository.TwoFactorRepository
	encKey []byte
	issuer string
	log    *zap.Logger
}

func NewTwoFactorService(repo *repository.TwoFactorRepository, encKey []byte, issuer string, log *zap.Logger) *TwoFactorService {
	return &TwoFactorService{
		repo:   repo,
		encKey: encKey,
		issuer: issuer,
		log:    log,
	}
}

func (s *TwoFactorService) IsEnabled(ctx context.Context, userID uint) (bool, error) {
	cfg, err := s.repo.GetUserConfig(ctx, userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.log.Error("failed to get user config", zap.Error(err))
		return false, err
	}
	if err != nil {
		s.log.Error("failed to get user config", zap.Error(err))
		return false, err
	}
	return cfg != nil && cfg.Enabled, nil
}

func (s *TwoFactorService) Setup(ctx context.Context, userID uint, email string) (dto.SetupResult, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		s.log.Error("failed to encrypt secret", zap.Error(err))
		return dto.SetupResult{}, err
	}
	now := time.Now()
	enc, err := s.encrypt([]byte(key.Secret()))
	if err != nil {
		s.log.Error("failed to decrypt secret", zap.Error(err))
		return dto.SetupResult{}, err
	}
	cfg := &model.UserTwoFactor{
		UserID:    userID,
		SecretEnc: enc,
		Enabled:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.UpsertUserConfig(ctx, cfg); err != nil {
		s.log.Error("failed to upsert user config", zap.Error(err))
		return dto.SetupResult{}, err
	}
	return dto.SetupResult{
		Secret:     key.Secret(),
		OtpauthURL: key.URL(),
	}, nil
}

func (s *TwoFactorService) Enable(ctx context.Context, userID uint, code string) error {
	cfg, err := s.repo.GetUserConfig(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user config", zap.Error(err))
		return err
	}
	if cfg == nil {
		s.log.Error("2fa not initialized")
		return errors.New("2fa not initialized")
	}
	secret, err := s.decrypt(cfg.SecretEnc)
	if err != nil {
		s.log.Error("failed to decrypt secret", zap.Error(err))
		return err
	}
	if !totp.Validate(code, string(secret)) {
		s.log.Error("invalid 2fa code", zap.String("code", code))
		return errors.New("invalid 2fa code")
	}
	now := time.Now()
	cfg.Enabled = true
	cfg.UpdatedAt = now
	cfg.VerifiedAt = &now
	if err := s.repo.UpsertUserConfig(ctx, cfg); err != nil {
		s.log.Error("failed to upsert user config", zap.Error(err))
		return err
	}
	return nil
}

func (s *TwoFactorService) NewLoginChallenge(ctx context.Context, userID uint, deviceID string, ttl time.Duration) (string, time.Time, error) {
	id, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return "", time.Time{}, errors.New("failed to generate new uuid")
	}
	now := time.Now()
	ch := &model.TwoFactorChallenge{
		ID:        id,
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
	if err := s.repo.CreateChallenge(ctx, ch); err != nil {
		s.log.Error("failed to create challenge", zap.Error(err))
		return "", time.Time{}, err
	}
	return id, ch.ExpiresAt, nil
}

func (s *TwoFactorService) VerifyChallenge(ctx context.Context, challengeID string, deviceID string, code string, maxAttempts int) (uint, error) {
	now := time.Now()
	ch, err := s.repo.FindValidChallenge(ctx, challengeID, now, maxAttempts)
	if err != nil {
		s.log.Error("failed to find valid challenge", zap.Error(err))
		return 0, errors.New("invalid or expired challenge")
	}
	if ch.DeviceID != deviceID {
		_ = s.repo.IncrementAttempts(ctx, ch.ID)
		s.log.Error("invalid challenge", zap.String("challenge_id", challengeID), zap.String("device_id", deviceID))
		return 0, errors.New("invalid challenge")
	}

	cfg, err := s.repo.GetUserConfig(ctx, ch.UserID)
	if err != nil {
		return 0, err
	}
	if cfg == nil || !cfg.Enabled {
		return 0, errors.New("2fa not enabled")
	}
	secret, err := s.decrypt(cfg.SecretEnc)
	if err != nil {
		return 0, err
	}
	if !totp.Validate(code, string(secret)) {
		_ = s.repo.IncrementAttempts(ctx, ch.ID)
		return 0, errors.New("invalid 2fa code")
	}
	if err := s.repo.MarkChallengeUsed(ctx, ch.ID, now); err != nil {
		return 0, err
	}
	return ch.UserID, nil
}

// encrypt/decrypt helpers (AES-GCM, base64-encoded)

func (s *TwoFactorService) encrypt(plain []byte) (string, error) {
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		s.log.Error("failed to new cipher", zap.Error(err))
		return "", errors.New("failed to new cipher")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.log.Error("failed to new cipher", zap.Error(err))
		return "", errors.New("failed to new cipher")
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		s.log.Error("failed to read full", zap.Error(err))
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *TwoFactorService) decrypt(enc string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		s.log.Error("failed to new cipher", zap.Error(err))
		return nil, err
	}
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		s.log.Error("failed to new cipher", zap.Error(err))
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.log.Error("failed to encrypt", zap.Error(err))
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		s.log.Error("failed to decrypt", zap.Error(err))
		return nil, err
	}
	return plain, nil
}
