package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/service/dto"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCurrentPass = errors.New("invalid current password")
)

type AuthUserRepo interface {
	Create(ctx context.Context, u *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	UpdatePassword(ctx context.Context, userID uint, newHash string) error
	UpdateEmail(ctx context.Context, userID uint, newEmail string) error
}

type AuthRepo interface {
	CreateSession(ctx context.Context, s *model.AuthSession) error
	SessionActive(ctx context.Context, sessionID string) (bool, error)

	CreateRefreshToken(ctx context.Context, t *model.RefreshToken) error
	FindRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, refreshTokenID uint, usedAt time.Time) error
	RevokeRefreshFamily(ctx context.Context, family string, reason string) error
	RevokeSession(ctx context.Context, sessionID string, revokedBy *uint) error
	RevokeRefreshBySessionID(ctx context.Context, sessionID string, reason string) error
	CreateRevokedJTI(ctx context.Context, r *model.RevokedJTI) error
}

type AuthJWT interface {
	DefaultRegistered(subject string, ttl time.Duration) jwt.RegisteredClaims
	IssueAccessToken(cl dto.AccessClaims) (string, error)
}

type AuthRBAC interface {
	AssignRole(ctx context.Context, userID uint, role string) (bool, error)
	RolesForUser(ctx context.Context, userID uint) ([]string, error)
	PermissionsForUser(ctx context.Context, userID uint) ([]string, error)
}

type AuthTwoFA interface {
	IsEnabled(ctx context.Context, userID uint) (bool, error)
	Setup(ctx context.Context, userID uint, email string) (dto.SetupResult, error)
	Enable(ctx context.Context, userID uint, code string) error
	NewLoginChallenge(ctx context.Context, userID uint, deviceID string, ttl time.Duration) (string, time.Time, error)
	VerifyChallenge(ctx context.Context, challengeID string, deviceID string, code string, maxAttempts int) (uint, error)
}

type AuthAudit interface {
	CreateImpersonation(ctx context.Context, a *model.ImpersonationAudit) error
}

type AuthService struct {
	log            *zap.Logger
	users          AuthUserRepo
	auth           AuthRepo
	jwt            AuthJWT
	audit          AuthAudit
	rbac           AuthRBAC
	twoFA          AuthTwoFA
	mediaSvc       *MediaService
	accessTTL      time.Duration
	refreshTTLDays int
	impersonateTTL time.Duration
	refreshPepper  string
}

func NewAuthService(users AuthUserRepo,
	authRepo AuthRepo,
	auditRepo AuthAudit,
	rbacSvc AuthRBAC,
	jwtm AuthJWT,
	twoFA AuthTwoFA,
	mediaSvc *MediaService,
	accessTTLMinutes int,
	refreshTTLDays int,
	impersonationTTLMinutes int,
	refreshPepper string,
	log *zap.Logger) *AuthService {
	return &AuthService{
		users:          users,
		auth:           authRepo,
		audit:          auditRepo,
		rbac:           rbacSvc,
		twoFA:          twoFA,
		jwt:            jwtm,
		accessTTL:      time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTLDays: refreshTTLDays,
		impersonateTTL: time.Duration(impersonationTTLMinutes) * time.Minute,
		refreshPepper:  refreshPepper,
		mediaSvc:       mediaSvc,
		log:            log,
	}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || name == "" || password == "" {
		return nil, errors.New("name, email, password are required")
	}

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		s.log.Error("email already registered", zap.String("email", email))
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to generate password hash", zap.Error(err))
		return nil, err
	}

	u := &model.User{
		Name:     strings.TrimSpace(name),
		Email:    email,
		Password: string(hash),
	}
	if err := s.users.Create(ctx, u); err != nil {
		s.log.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	// Assign default RBAC role.
	if s.rbac != nil {
		if _, err := s.rbac.AssignRole(ctx, u.ID, "user"); err != nil {
			s.log.Error("failed to assign role", zap.Error(err))
			return nil, err
		}
	}
	return u, nil
}

func (s *AuthService) Profile(ctx context.Context, userID uint) (dto.AuthUser, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		s.log.Error("failed to find by id", zap.Error(err))
		return dto.AuthUser{}, err
	}
	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		s.log.Error("failed to load role and permissions", zap.Error(err))
		return dto.AuthUser{}, err
	}
	avatar, err := s.mediaSvc.UserAvatar(ctx, u)
	if err != nil {
		s.log.Error("failed to get user avatar", zap.Error(err))
		return dto.AuthUser{}, err
	}
	return dto.AuthUser{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Role:        role,
		Permissions: perms,
		Avatar:      avatar,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		s.log.Error("failed to find by id", zap.Error(err))
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPassword)); err != nil {
		s.log.Error("invalid current password", zap.Error(err))
		return ErrInvalidCurrentPass
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to generate password hash", zap.Error(err))
		return err
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

func (s *AuthService) ChangeEmail(ctx context.Context, userID uint, currentPassword, newEmail string) error {
	newEmail = strings.TrimSpace(strings.ToLower(newEmail))
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if newEmail == "" {
		s.log.Error("new email is required")
		return errors.New("newEmail is required")
	}
	if newEmail == u.Email {
		return nil // no-op
	}
	_, err = s.users.FindByEmail(ctx, newEmail)
	if err == nil {
		s.log.Error("email already registered", zap.String("email", newEmail))
		return ErrEmailTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPassword)); err != nil {
		s.log.Error("invalid current password", zap.Error(err))
		return ErrInvalidCurrentPass
	}
	return s.users.UpdateEmail(ctx, userID, newEmail)
}

func (s *AuthService) SetupTwoFA(ctx context.Context, userID uint, email string) (dto.TwoFactorSetupResult, error) {
	if s.twoFA == nil {
		s.log.Error("2fa service not configured")
		return dto.TwoFactorSetupResult{}, errors.New("2fa service not configured")
	}
	setup, err := s.twoFA.Setup(ctx, userID, email)
	if err != nil {
		s.log.Error("failed to setup two factor", zap.Error(err))
		return dto.TwoFactorSetupResult{}, err
	}
	return dto.TwoFactorSetupResult{
		Secret:     setup.Secret,
		OtpauthURL: setup.OtpauthURL,
	}, nil
}

func (s *AuthService) EnableTwoFA(ctx context.Context, userID uint, code string) error {
	if s.twoFA == nil {
		s.log.Error("2fa service not configured")
		return errors.New("2fa service not configured")
	}
	return s.twoFA.Enable(ctx, userID, code)
}

func (s *AuthService) VerifyTwoFAChallenge(ctx context.Context, challengeID string, deviceID string, code string) (dto.LoginResult, error) {
	if s.twoFA == nil {
		s.log.Error("2fa service not configured")
		return dto.LoginResult{}, errors.New("2fa service not configured")
	}
	userID, err := s.twoFA.VerifyChallenge(ctx, challengeID, deviceID, code, 5)
	if err != nil {
		s.log.Error("failed to verify challenge", zap.Error(err))
		return dto.LoginResult{}, err
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		s.log.Error("failed to find by id", zap.Error(err))
		return dto.LoginResult{}, err
	}

	// Create session
	sessionID, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.LoginResult{}, err
	}
	now := time.Now()
	sess := &model.AuthSession{
		ID:         sessionID,
		UserID:     u.ID,
		DeviceID:   deviceID,
		IPAddress:  "",
		UserAgent:  "",
		LastSeenAt: now,
	}
	if err := s.auth.CreateSession(ctx, sess); err != nil {
		return dto.LoginResult{}, err
	}

	accessJTI, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.LoginResult{}, err
	}
	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		s.log.Error("failed to load role and permissions", zap.Error(err))
		return dto.LoginResult{}, err
	}
	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", u.ID), s.accessTTL)
	rc.ID = accessJTI
	claims := dto.AccessClaims{
		RegisteredClaims: rc,
		UserID:           u.ID,
		Role:             role,
		Permissions:      perms,
		SessionID:        sessionID,
		DeviceID:         deviceID,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		s.log.Error("failed to issue access token", zap.Error(err))
		return dto.LoginResult{}, err
	}
	refreshToken, rtModel, err := s.issueRefreshToken(ctx, u.ID, sessionID, nil)
	if err != nil {
		s.log.Error("failed to issue refresh token", zap.Error(err))
		return dto.LoginResult{}, err
	}

	return dto.LoginResult{
		TwoFactorRequired: false,
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		ExpiresAt:         rtModel.ExpiresAt,
		SessionID:         sessionID,
		User: dto.AuthUser{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			Role:        role,
			Permissions: perms,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, meta dto.LoginMeta) (dto.LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.LoginResult{}, ErrInvalidCredentials
		}
		return dto.LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		s.log.Error("invalid credentials", zap.Error(err))
		return dto.LoginResult{}, ErrInvalidCredentials
	}

	if meta.DeviceID == "" {
		s.log.Error("deviceId is required")
		return dto.LoginResult{}, errors.New("deviceId is required")
	}

	// Create session
	sessionID, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.LoginResult{}, err
	}
	now := time.Now()
	sess := &model.AuthSession{
		ID:         sessionID,
		UserID:     u.ID,
		DeviceID:   meta.DeviceID,
		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
		LastSeenAt: now,
	}
	if err := s.auth.CreateSession(ctx, sess); err != nil {
		s.log.Error("failed to create session", zap.Error(err))
		return dto.LoginResult{}, err
	}

	// If 2FA is enabled, create a challenge and return without tokens.
	if s.twoFA != nil {
		enabled, err := s.twoFA.IsEnabled(ctx, u.ID)
		if err != nil {
			s.log.Error("failed to check if 2fa is enabled", zap.Error(err))
			return dto.LoginResult{}, err
		}
		if enabled {
			chID, exp, err := s.twoFA.NewLoginChallenge(ctx, u.ID, meta.DeviceID, 5*time.Minute)
			if err != nil {
				s.log.Error("failed to create login challenge", zap.Error(err))
				return dto.LoginResult{}, err
			}
			return dto.LoginResult{
				TwoFactorRequired: true,
				ChallengeID:       chID,
				ExpiresAt:         exp,
				SessionID:         sessionID,
				User: dto.AuthUser{
					ID:    u.ID,
					Name:  u.Name,
					Email: u.Email,
				},
			}, nil
		}
	}

	accessJTI, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.LoginResult{}, err
	}

	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		s.log.Error("failed to load role and permissions", zap.Error(err))
		return dto.LoginResult{}, err
	}

	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", u.ID), s.accessTTL)
	rc.ID = accessJTI
	claims := dto.AccessClaims{
		RegisteredClaims: rc,
		UserID:           u.ID,
		Role:             role,
		Permissions:      perms,
		SessionID:        sessionID,
		DeviceID:         meta.DeviceID,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		s.log.Error("failed to issue access token", zap.Error(err))
		return dto.LoginResult{}, err
	}

	refreshToken, rtModel, err := s.issueRefreshToken(ctx, u.ID, sessionID, nil)
	if err != nil {
		s.log.Error("failed to issue refresh token", zap.Error(err))
		return dto.LoginResult{}, err
	}

	return dto.LoginResult{
		TwoFactorRequired: false,
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		ExpiresAt:         rtModel.ExpiresAt,
		SessionID:         sessionID,
		User: dto.AuthUser{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			Role:        role,
			Permissions: perms,
		},
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta dto.LoginMeta) (dto.RefreshResult, error) {
	if refreshToken == "" {
		return dto.RefreshResult{}, errors.New("refresh_token is required")
	}
	hash, err := hashRefreshToken(refreshToken, s.refreshPepper, s.log)
	if err != nil {
		s.log.Error("failed to hash refresh token", zap.Error(err))
		return dto.RefreshResult{}, err
	}
	rt, err := s.auth.FindRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.RefreshResult{}, ErrInvalidCredentials
		}
		return dto.RefreshResult{}, err
	}

	now := time.Now()
	if rt.RevokedAt != nil || rt.ExpiresAt.Before(now) {
		s.log.Error("refresh token revoked or expired")
		return dto.RefreshResult{}, ErrInvalidCredentials
	}
	if rt.UsedAt != nil {
		// Refresh token reuse detected -> revoke family + session
		_ = s.auth.RevokeRefreshFamily(ctx, rt.TokenFamily, "refresh reuse detected")
		_ = s.auth.RevokeSession(ctx, rt.SessionID, nil)
		return dto.RefreshResult{}, ErrInvalidCredentials
	}

	active, err := s.auth.SessionActive(ctx, rt.SessionID)
	if err != nil {
		return dto.RefreshResult{}, err
	}
	if !active {
		return dto.RefreshResult{}, ErrInvalidCredentials
	}

	// Mark used and rotate
	if err := s.auth.MarkRefreshTokenUsed(ctx, rt.ID, now); err != nil {
		return dto.RefreshResult{}, err
	}

	refreshOut, _, err := s.issueRefreshToken(ctx, rt.UserID, rt.SessionID, &rt.ID)
	if err != nil {
		return dto.RefreshResult{}, err
	}

	u, err := s.users.FindByID(ctx, rt.UserID)
	if err != nil {
		return dto.RefreshResult{}, err
	}

	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		return dto.RefreshResult{}, err
	}

	jti, err := newUUIDLike(s.log)
	s.log.Error("failed to generate new uuid", zap.Error(err))
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.RefreshResult{}, err
	}
	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", u.ID), s.accessTTL)
	rc.ID = jti
	claims := dto.AccessClaims{
		RegisteredClaims: rc,
		UserID:           u.ID,
		Role:             role,
		Permissions:      perms,
		SessionID:        rt.SessionID,
		DeviceID:         meta.DeviceID,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		return dto.RefreshResult{}, err
	}
	return dto.RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: refreshOut,
		ExpiresAt:    rc.ExpiresAt.Time,
		SessionID:    rt.SessionID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string, accessJTI string, accessExp time.Time, userID uint) error {
	_ = s.auth.RevokeSession(ctx, sessionID, &userID)
	_ = s.auth.RevokeRefreshBySessionID(ctx, sessionID, "logout")
	if accessJTI != "" && accessExp.After(time.Now()) {
		_ = s.auth.CreateRevokedJTI(ctx, &model.RevokedJTI{
			JTI:       accessJTI,
			UserID:    userID,
			SessionID: sessionID,
			Reason:    "logout",
			ExpiresAt: accessExp,
		})
	}
	return nil
}

func (s *AuthService) issueRefreshToken(ctx context.Context, userID uint, sessionID string, rotatedFrom *uint) (string, *model.RefreshToken, error) {
	if s.refreshPepper == "" {
		return "", nil, errors.New("REFRESH_TOKEN_PEPPER is required")
	}
	raw, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return "", nil, err
	}
	family := sessionID
	if rotatedFrom != nil {
		family = sessionID // keep one family per session for simplicity
	}
	hash, err := hashRefreshToken(raw, s.refreshPepper, s.log)
	if err != nil {
		return "", nil, err
	}
	exp := time.Now().Add(time.Duration(s.refreshTTLDays) * 24 * time.Hour)
	m := &model.RefreshToken{
		SessionID:     sessionID,
		UserID:        userID,
		TokenHash:     hash,
		TokenFamily:   family,
		RotatedFromID: rotatedFrom,
		ExpiresAt:     exp,
	}
	if err := s.auth.CreateRefreshToken(ctx, m); err != nil {
		return "", nil, err
	}
	return raw, m, nil
}

func (s *AuthService) Impersonate(ctx context.Context, impersonatorID uint, targetUserID uint, reason string, meta dto.LoginMeta) (dto.ImpersonationResult, error) {
	impRole, _, err := s.loadRoleAndPerms(ctx, impersonatorID)
	if err != nil {
		s.log.Error("failed to load role and permissions", zap.Error(err))
		return dto.ImpersonationResult{}, err
	}
	if impRole != "admin" && impRole != "support" {
		s.log.Error("impersonate role not allowed")
		return dto.ImpersonationResult{}, errors.New("forbidden")
	}
	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("target user not found")
			return dto.ImpersonationResult{}, errors.New("target user not found")
		}
		return dto.ImpersonationResult{}, err
	}

	role, perms, err := s.loadRoleAndPerms(ctx, target.ID)
	if err != nil {
		s.log.Error("failed to load role and permissions", zap.Error(err))
		return dto.ImpersonationResult{}, err
	}

	sessionID, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.ImpersonationResult{}, err
	}
	now := time.Now()
	sess := &model.AuthSession{
		ID:         sessionID,
		UserID:     target.ID,
		DeviceID:   meta.DeviceID,
		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
		LastSeenAt: now,
		RevokedBy:  &impersonatorID, // can be used to track admin-driven session
	}
	if err := s.auth.CreateSession(ctx, sess); err != nil {
		s.log.Error("failed to create session", zap.Error(err))
		return dto.ImpersonationResult{}, err
	}

	jti, err := newUUIDLike(s.log)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return dto.ImpersonationResult{}, err
	}
	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", target.ID), s.impersonateTTL)
	rc.ID = jti
	claims := dto.AccessClaims{
		RegisteredClaims:    rc,
		UserID:              target.ID,
		Role:                role,
		Permissions:         perms,
		SessionID:           sessionID,
		DeviceID:            meta.DeviceID,
		Impersonation:       true,
		ImpersonatedUserID:  &target.ID,
		ImpersonatorID:      &impersonatorID,
		ImpersonationReason: reason,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		s.log.Error("failed to issue access token", zap.Error(err))
		return dto.ImpersonationResult{}, err
	}

	if s.audit != nil {
		if err := s.audit.CreateImpersonation(ctx, &model.ImpersonationAudit{
			ImpersonatorID:     impersonatorID,
			ImpersonatedUserID: target.ID,
			Reason:             reason,
			IPAddress:          meta.IPAddress,
			UserAgent:          meta.UserAgent,
		}); err != nil {
			s.log.Error("failed to create impersonation audit", zap.Error(err))
			return dto.ImpersonationResult{}, err
		}
	}

	return dto.ImpersonationResult{AccessToken: accessToken, ExpiresAt: rc.ExpiresAt.Time}, nil
}

func (s *AuthService) loadRoleAndPerms(ctx context.Context, userID uint) (string, []string, error) {
	// Default role if RBAC is not configured.
	if s.rbac == nil {
		return "user", []string{}, nil
	}
	roles, err := s.rbac.RolesForUser(ctx, userID)
	if err != nil {
		s.log.Error("failed to generate new uuid", zap.Error(err))
		return "", nil, err
	}
	role := "user"
	if len(roles) > 0 {
		role = roles[0]
	}
	perms, err := s.rbac.PermissionsForUser(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	return role, perms, nil
}
