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

	"go-rest/internal/model"
	"go-rest/internal/repository"
	"go-rest/internal/service/dto"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TwoFactorService struct {
	repo   *repository.TwoFactorRepository
	encKey []byte
	issuer string
}

func NewTwoFactorService(repo *repository.TwoFactorRepository, encKey []byte, issuer string) *TwoFactorService {
	return &TwoFactorService{
		repo:   repo,
		encKey: encKey,
		issuer: issuer,
	}
}

func (s *TwoFactorService) IsEnabled(ctx context.Context, userID uint) (bool, error) {
	cfg, err := s.repo.GetUserConfig(ctx, userID)
	if err != nil {
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
		return dto.SetupResult{}, err
	}
	now := time.Now()
	enc, err := s.encrypt([]byte(key.Secret()))
	if err != nil {
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
		return err
	}
	if cfg == nil {
		return errors.New("2fa not initialized")
	}
	secret, err := s.decrypt(cfg.SecretEnc)
	if err != nil {
		return err
	}
	if !totp.Validate(code, string(secret)) {
		return errors.New("invalid 2fa code")
	}
	now := time.Now()
	cfg.Enabled = true
	cfg.UpdatedAt = now
	cfg.VerifiedAt = &now
	return s.repo.UpsertUserConfig(ctx, cfg)
}

func (s *TwoFactorService) NewLoginChallenge(ctx context.Context, userID uint, deviceID string, ttl time.Duration) (string, time.Time, error) {
	id, err := newUUIDLike()
	if err != nil {
		return "", time.Time{}, err
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
		return "", time.Time{}, err
	}
	return id, ch.ExpiresAt, nil
}

func (s *TwoFactorService) VerifyChallenge(ctx context.Context, challengeID string, deviceID string, code string, maxAttempts int) (uint, error) {
	now := time.Now()
	ch, err := s.repo.FindValidChallenge(ctx, challengeID, now, maxAttempts)
	if err != nil {
		return 0, errors.New("invalid or expired challenge")
	}
	if ch.DeviceID != deviceID {
		_ = s.repo.IncrementAttempts(ctx, ch.ID)
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
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *TwoFactorService) decrypt(enc string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

