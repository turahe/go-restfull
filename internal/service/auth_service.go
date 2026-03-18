package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-rest/internal/model"
	"go-rest/internal/repository"
	svcresp "go-rest/internal/service/response"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
)

type AuthService struct {
	users *repository.UserRepository
	auth  *repository.AuthRepository
	jwt   *JWTService
	audit *repository.AuditRepository
	rbac  *RBACService
	twoFA *TwoFactorService

	accessTTL        time.Duration
	refreshTTLDays   int
	impersonateTTL   time.Duration
	refreshPepper    string
}

type TwoFactorSetupResult struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}

func NewAuthService(users *repository.UserRepository, authRepo *repository.AuthRepository, auditRepo *repository.AuditRepository, rbacSvc *RBACService, jwtm *JWTService, twoFA *TwoFactorService, accessTTLMinutes int, refreshTTLDays int, impersonationTTLMinutes int, refreshPepper string) *AuthService {
	return &AuthService{
		users:            users,
		auth:             authRepo,
		audit:            auditRepo,
		rbac:             rbacSvc,
		twoFA:            twoFA,
		jwt:              jwtm,
		accessTTL:        time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTLDays:   refreshTTLDays,
		impersonateTTL:   time.Duration(impersonationTTLMinutes) * time.Minute,
		refreshPepper:    refreshPepper,
	}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || name == "" || password == "" {
		return nil, errors.New("name, email, password are required")
	}

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Name:     strings.TrimSpace(name),
		Email:    email,
		Password: string(hash),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	// Assign default RBAC role.
	if s.rbac != nil {
		if _, err := s.rbac.AssignRole(ctx, u.ID, "user"); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (s *AuthService) Profile(ctx context.Context, userID uint) (svcresp.AuthUser, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return svcresp.AuthUser{}, err
	}
	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		return svcresp.AuthUser{}, err
	}
	return svcresp.AuthUser{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Role:        role,
		Permissions: perms,
	}, nil
}

func (s *AuthService) SetupTwoFA(ctx context.Context, userID uint, email string) (TwoFactorSetupResult, error) {
	if s.twoFA == nil {
		return TwoFactorSetupResult{}, errors.New("2fa service not configured")
	}
	setup, err := s.twoFA.Setup(ctx, userID, email)
	if err != nil {
		return TwoFactorSetupResult{}, err
	}
	return TwoFactorSetupResult{
		Secret:     setup.Secret,
		OtpauthURL: setup.OtpauthURL,
	}, nil
}

func (s *AuthService) EnableTwoFA(ctx context.Context, userID uint, code string) error {
	if s.twoFA == nil {
		return errors.New("2fa service not configured")
	}
	return s.twoFA.Enable(ctx, userID, code)
}

func (s *AuthService) VerifyTwoFAChallenge(ctx context.Context, challengeID string, deviceID string, code string) (svcresp.LoginResult, error) {
	if s.twoFA == nil {
		return svcresp.LoginResult{}, errors.New("2fa service not configured")
	}
	userID, err := s.twoFA.VerifyChallenge(ctx, challengeID, deviceID, code, 5)
	if err != nil {
		return svcresp.LoginResult{}, err
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return svcresp.LoginResult{}, err
	}

	// Create session
	sessionID, err := newUUIDLike()
	if err != nil {
		return svcresp.LoginResult{}, err
	}
	now := time.Now()
	sess := &model.AuthSession{
		ID:        sessionID,
		UserID:    u.ID,
		DeviceID:  deviceID,
		IPAddress: "",
		UserAgent: "",
		LastSeenAt: now,
	}
	if err := s.auth.CreateSession(ctx, sess); err != nil {
		return svcresp.LoginResult{}, err
	}

	accessJTI, err := newUUIDLike()
	if err != nil {
		return svcresp.LoginResult{}, err
	}
	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		return svcresp.LoginResult{}, err
	}
	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", u.ID), s.accessTTL)
	rc.ID = accessJTI
	claims := AccessClaims{
		RegisteredClaims: rc,
		UserID:           u.ID,
		Role:             role,
		Permissions:      perms,
		SessionID:        sessionID,
		DeviceID:         deviceID,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		return svcresp.LoginResult{}, err
	}
	refreshToken, rtModel, err := s.issueRefreshToken(ctx, u.ID, sessionID, nil)
	if err != nil {
		return svcresp.LoginResult{}, err
	}

	return svcresp.LoginResult{
		TwoFactorRequired: false,
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		ExpiresAt:         rtModel.ExpiresAt,
		SessionID:         sessionID,
		User: svcresp.AuthUser{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			Role:        role,
			Permissions: perms,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, meta svcresp.LoginMeta) (svcresp.LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return svcresp.LoginResult{}, ErrInvalidCredentials
		}
		return svcresp.LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return svcresp.LoginResult{}, ErrInvalidCredentials
	}

	if meta.DeviceID == "" {
		return svcresp.LoginResult{}, errors.New("deviceId is required")
	}

	// Create session
	sessionID, err := newUUIDLike()
	if err != nil {
		return svcresp.LoginResult{}, err
	}
	now := time.Now()
	sess := &model.AuthSession{
		ID:        sessionID,
		UserID:    u.ID,
		DeviceID:  meta.DeviceID,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		LastSeenAt: now,
	}
	if err := s.auth.CreateSession(ctx, sess); err != nil {
		return svcresp.LoginResult{}, err
	}

	// If 2FA is enabled, create a challenge and return without tokens.
	if s.twoFA != nil {
		enabled, err := s.twoFA.IsEnabled(ctx, u.ID)
		if err != nil {
			return svcresp.LoginResult{}, err
		}
		if enabled {
			chID, exp, err := s.twoFA.NewLoginChallenge(ctx, u.ID, meta.DeviceID, 5*time.Minute)
			if err != nil {
				return svcresp.LoginResult{}, err
			}
			return svcresp.LoginResult{
				TwoFactorRequired: true,
				ChallengeID:       chID,
				ExpiresAt:         exp,
				SessionID:         sessionID,
				User: svcresp.AuthUser{
					ID:    u.ID,
					Name:  u.Name,
					Email: u.Email,
				},
			}, nil
		}
	}

	accessJTI, err := newUUIDLike()
	if err != nil {
		return svcresp.LoginResult{}, err
	}

	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		return svcresp.LoginResult{}, err
	}

	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", u.ID), s.accessTTL)
	rc.ID = accessJTI
	claims := AccessClaims{
		RegisteredClaims: rc,
		UserID:           u.ID,
		Role:             role,
		Permissions:      perms,
		SessionID:        sessionID,
		DeviceID:         meta.DeviceID,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		return svcresp.LoginResult{}, err
	}

	refreshToken, rtModel, err := s.issueRefreshToken(ctx, u.ID, sessionID, nil)
	if err != nil {
		return svcresp.LoginResult{}, err
	}

	return svcresp.LoginResult{
		TwoFactorRequired: false,
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		ExpiresAt:         rtModel.ExpiresAt,
		SessionID:         sessionID,
		User: svcresp.AuthUser{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			Role:        role,
			Permissions: perms,
		},
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta svcresp.LoginMeta) (svcresp.RefreshResult, error) {
	if refreshToken == "" {
		return svcresp.RefreshResult{}, errors.New("refresh_token is required")
	}
	hash := hashRefreshToken(refreshToken, s.refreshPepper)
	rt, err := s.auth.FindRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return svcresp.RefreshResult{}, ErrInvalidCredentials
		}
		return svcresp.RefreshResult{}, err
	}

	now := time.Now()
	if rt.RevokedAt != nil || rt.ExpiresAt.Before(now) {
		return svcresp.RefreshResult{}, ErrInvalidCredentials
	}
	if rt.UsedAt != nil {
		// Refresh token reuse detected -> revoke family + session
		_ = s.auth.RevokeRefreshFamily(ctx, rt.TokenFamily, "refresh reuse detected")
		_ = s.auth.RevokeSession(ctx, rt.SessionID, nil)
		return svcresp.RefreshResult{}, ErrInvalidCredentials
	}

	active, err := s.auth.SessionActive(ctx, rt.SessionID)
	if err != nil {
		return svcresp.RefreshResult{}, err
	}
	if !active {
		return svcresp.RefreshResult{}, ErrInvalidCredentials
	}

	// Mark used and rotate
	if err := s.auth.MarkRefreshTokenUsed(ctx, rt.ID, now); err != nil {
		return svcresp.RefreshResult{}, err
	}

	refreshOut, _, err := s.issueRefreshToken(ctx, rt.UserID, rt.SessionID, &rt.ID)
	if err != nil {
		return svcresp.RefreshResult{}, err
	}

	u, err := s.users.FindByID(ctx, rt.UserID)
	if err != nil {
		return svcresp.RefreshResult{}, err
	}

	role, perms, err := s.loadRoleAndPerms(ctx, u.ID)
	if err != nil {
		return svcresp.RefreshResult{}, err
	}

	jti, err := newUUIDLike()
	if err != nil {
		return svcresp.RefreshResult{}, err
	}
	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", u.ID), s.accessTTL)
	rc.ID = jti
	claims := AccessClaims{
		RegisteredClaims: rc,
		UserID:           u.ID,
		Role:             role,
		Permissions:      perms,
		SessionID:        rt.SessionID,
		DeviceID:         meta.DeviceID,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		return svcresp.RefreshResult{}, err
	}
	return svcresp.RefreshResult{
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
	raw, err := newUUIDLike()
	if err != nil {
		return "", nil, err
	}
	family := sessionID
	if rotatedFrom != nil {
		family = sessionID // keep one family per session for simplicity
	}
	hash := hashRefreshToken(raw, s.refreshPepper)
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

func (s *AuthService) Impersonate(ctx context.Context, impersonatorID uint, targetUserID uint, reason string, meta svcresp.LoginMeta) (svcresp.ImpersonationResult, error) {
	impRole, _, err := s.loadRoleAndPerms(ctx, impersonatorID)
	if err != nil {
		return svcresp.ImpersonationResult{}, err
	}
	if impRole != "admin" && impRole != "support" {
		return svcresp.ImpersonationResult{}, errors.New("forbidden")
	}
	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return svcresp.ImpersonationResult{}, errors.New("target user not found")
		}
		return svcresp.ImpersonationResult{}, err
	}

	role, perms, err := s.loadRoleAndPerms(ctx, target.ID)
	if err != nil {
		return svcresp.ImpersonationResult{}, err
	}

	sessionID, err := newUUIDLike()
	if err != nil {
		return svcresp.ImpersonationResult{}, err
	}
	now := time.Now()
	sess := &model.AuthSession{
		ID:        sessionID,
		UserID:    target.ID,
		DeviceID:  meta.DeviceID,
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		LastSeenAt: now,
		RevokedBy: &impersonatorID, // can be used to track admin-driven session
	}
	if err := s.auth.CreateSession(ctx, sess); err != nil {
		return svcresp.ImpersonationResult{}, err
	}

	jti, err := newUUIDLike()
	if err != nil {
		return svcresp.ImpersonationResult{}, err
	}
	rc := s.jwt.DefaultRegistered(fmt.Sprintf("%d", target.ID), s.impersonateTTL)
	rc.ID = jti
	claims := AccessClaims{
		RegisteredClaims:       rc,
		UserID:                target.ID,
		Role:                  role,
		Permissions:           perms,
		SessionID:             sessionID,
		DeviceID:              meta.DeviceID,
		Impersonation:         true,
		ImpersonatedUserID:    &target.ID,
		ImpersonatorID:        &impersonatorID,
		ImpersonationReason:   reason,
	}
	accessToken, err := s.jwt.IssueAccessToken(claims)
	if err != nil {
		return svcresp.ImpersonationResult{}, err
	}

	if s.audit != nil {
		_ = s.audit.CreateImpersonation(ctx, &model.ImpersonationAudit{
			ImpersonatorID:     impersonatorID,
			ImpersonatedUserID: target.ID,
			Reason:             reason,
			IPAddress:          meta.IPAddress,
			UserAgent:          meta.UserAgent,
		})
	}

	return svcresp.ImpersonationResult{AccessToken: accessToken, ExpiresAt: rc.ExpiresAt.Time}, nil
}

func (s *AuthService) loadRoleAndPerms(ctx context.Context, userID uint) (string, []string, error) {
	// Default role if RBAC is not configured.
	if s.rbac == nil {
		return "user", []string{}, nil
	}
	roles, err := s.rbac.RolesForUser(ctx, userID)
	if err != nil {
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

