package service

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/service/dto"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type mockAuthUserRepo struct{ mock.Mock }

func (m *mockAuthUserRepo) Create(ctx context.Context, u *model.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockAuthUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}
func (m *mockAuthUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}
func (m *mockAuthUserRepo) UpdatePassword(ctx context.Context, userID uint, newHash string) error {
	return m.Called(ctx, userID, newHash).Error(0)
}
func (m *mockAuthUserRepo) UpdateEmail(ctx context.Context, userID uint, newEmail string) error {
	return m.Called(ctx, userID, newEmail).Error(0)
}

type mockAuthRepo struct{ mock.Mock }

func (m *mockAuthRepo) CreateSession(ctx context.Context, s *model.AuthSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockAuthRepo) SessionActive(ctx context.Context, sessionID string) (bool, error) {
	args := m.Called(ctx, sessionID)
	return args.Bool(0), args.Error(1)
}
func (m *mockAuthRepo) CreateRefreshToken(ctx context.Context, t *model.RefreshToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockAuthRepo) FindRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	args := m.Called(ctx, hash)
	rt, _ := args.Get(0).(*model.RefreshToken)
	return rt, args.Error(1)
}
func (m *mockAuthRepo) MarkRefreshTokenUsed(ctx context.Context, refreshTokenID uint, usedAt time.Time) error {
	return m.Called(ctx, refreshTokenID, usedAt).Error(0)
}
func (m *mockAuthRepo) RevokeRefreshFamily(ctx context.Context, family string, reason string) error {
	return m.Called(ctx, family, reason).Error(0)
}
func (m *mockAuthRepo) RevokeSession(ctx context.Context, sessionID string, revokedBy *uint) error {
	return m.Called(ctx, sessionID, revokedBy).Error(0)
}
func (m *mockAuthRepo) RevokeRefreshBySessionID(ctx context.Context, sessionID string, reason string) error {
	return m.Called(ctx, sessionID, reason).Error(0)
}
func (m *mockAuthRepo) CreateRevokedJTI(ctx context.Context, r *model.RevokedJTI) error {
	return m.Called(ctx, r).Error(0)
}

type mockJWT struct{ mock.Mock }

func (m *mockJWT) DefaultRegistered(subject string, ttl time.Duration) jwt.RegisteredClaims {
	args := m.Called(subject, ttl)
	return args.Get(0).(jwt.RegisteredClaims)
}
func (m *mockJWT) IssueAccessToken(cl dto.AccessClaims) (string, error) {
	args := m.Called(cl)
	return args.String(0), args.Error(1)
}

type mockRBAC struct{ mock.Mock }

func (m *mockRBAC) AssignRole(ctx context.Context, userID uint, role string) (bool, error) {
	args := m.Called(ctx, userID, role)
	return args.Bool(0), args.Error(1)
}
func (m *mockRBAC) RolesForUser(ctx context.Context, userID uint) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockRBAC) PermissionsForUser(ctx context.Context, userID uint) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

type mockTwoFA struct{ mock.Mock }

func (m *mockTwoFA) IsEnabled(ctx context.Context, userID uint) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}
func (m *mockTwoFA) Setup(ctx context.Context, userID uint, email string) (dto.SetupResult, error) {
	args := m.Called(ctx, userID, email)
	return args.Get(0).(dto.SetupResult), args.Error(1)
}
func (m *mockTwoFA) Enable(ctx context.Context, userID uint, code string) error {
	return m.Called(ctx, userID, code).Error(0)
}
func (m *mockTwoFA) NewLoginChallenge(ctx context.Context, userID uint, deviceID string, ttl time.Duration) (string, time.Time, error) {
	args := m.Called(ctx, userID, deviceID, ttl)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}
func (m *mockTwoFA) VerifyChallenge(ctx context.Context, challengeID string, deviceID string, code string, maxAttempts int) (uint, error) {
	args := m.Called(ctx, challengeID, deviceID, code, maxAttempts)
	return uint(args.Int(0)), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("email taken", func(t *testing.T) {
		t.Parallel()
		users := &mockAuthUserRepo{}
		users.On("FindByEmail", mock.Anything, "a@b.com").Return(&model.User{ID: 1}, nil).Once()

		s := NewAuthService(users, &mockAuthRepo{}, nil, nil, &mockJWT{}, nil, nil, 10, 30, 5, "pepper", zap.NewNop())
		_, err := s.Register(ctx, "n", "A@B.com", "pass")
		assert.ErrorIs(t, err, ErrEmailTaken)
		users.AssertExpectations(t)
	})

	t.Run("success assigns default role when rbac enabled", func(t *testing.T) {
		t.Parallel()
		users := &mockAuthUserRepo{}
		rbac := &mockRBAC{}

		users.On("FindByEmail", mock.Anything, "a@b.com").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()
		users.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Run(func(args mock.Arguments) {
			u := args.Get(1).(*model.User)
			u.ID = 99
		}).Once()
		rbac.On("AssignRole", mock.Anything, uint(99), entities.RoleUser).Return(true, nil).Once()

		s := NewAuthService(users, &mockAuthRepo{}, nil, rbac, &mockJWT{}, nil, nil, 10, 30, 5, "pepper", zap.NewNop())
		u, err := s.Register(ctx, " Name ", "A@B.com", "password")
		assert.NoError(t, err)
		assert.Equal(t, uint(99), u.ID)
		assert.Equal(t, "a@b.com", u.Email)
		assert.NotEmpty(t, u.Password) // hashed

		users.AssertExpectations(t)
		rbac.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	hash, _ := bcryptHash("12345678")

	t.Run("user not found -> invalid credentials", func(t *testing.T) {
		t.Parallel()
		users := &mockAuthUserRepo{}
		users.On("FindByEmail", mock.Anything, "a@b.com").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()
		s := NewAuthService(users, &mockAuthRepo{}, nil, nil, &mockJWT{}, nil, nil, 10, 30, 5, "pepper", zap.NewNop())

		_, err := s.Login(ctx, "a@b.com", "12345678", dto.LoginMeta{DeviceID: "dev1"})
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		users.AssertExpectations(t)
	})

	t.Run("deviceId required", func(t *testing.T) {
		t.Parallel()
		users := &mockAuthUserRepo{}
		users.On("FindByEmail", mock.Anything, "a@b.com").Return(&model.User{ID: 1, Email: "a@b.com", Password: hash}, nil).Once()
		s := NewAuthService(users, &mockAuthRepo{}, nil, nil, &mockJWT{}, nil, nil, 10, 30, 5, "pepper", zap.NewNop())

		_, err := s.Login(ctx, "a@b.com", "12345678", dto.LoginMeta{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deviceId is required")
		users.AssertExpectations(t)
	})

	t.Run("2FA enabled returns challenge without tokens", func(t *testing.T) {
		t.Parallel()
		users := &mockAuthUserRepo{}
		authRepo := &mockAuthRepo{}
		twoFA := &mockTwoFA{}

		users.On("FindByEmail", mock.Anything, "a@b.com").Return(&model.User{ID: 1, Name: "A", Email: "a@b.com", Password: hash}, nil).Once()
		authRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*model.AuthSession")).Return(nil).Once()
		twoFA.On("IsEnabled", mock.Anything, uint(1)).Return(true, nil).Once()
		exp := time.Now().Add(5 * time.Minute)
		twoFA.On("NewLoginChallenge", mock.Anything, uint(1), "dev1", 5*time.Minute).Return("ch", exp, nil).Once()

		s := NewAuthService(users, authRepo, nil, nil, &mockJWT{}, twoFA, nil, 10, 30, 5, "pepper", zap.NewNop())
		res, err := s.Login(ctx, "a@b.com", "12345678", dto.LoginMeta{DeviceID: "dev1"})
		assert.NoError(t, err)
		assert.True(t, res.TwoFactorRequired)
		assert.Equal(t, "ch", res.ChallengeID)
		assert.Equal(t, "a@b.com", res.User.Email)
		assert.Empty(t, res.AccessToken)
		assert.Empty(t, res.RefreshToken)

		users.AssertExpectations(t)
		authRepo.AssertExpectations(t)
		twoFA.AssertExpectations(t)
	})

	t.Run("success without 2FA issues tokens", func(t *testing.T) {
		t.Parallel()
		users := &mockAuthUserRepo{}
		authRepo := &mockAuthRepo{}
		j := &mockJWT{}
		rbac := &mockRBAC{}

		users.On("FindByEmail", mock.Anything, "a@b.com").Return(&model.User{ID: 7, Name: "A", Email: "a@b.com", Password: hash}, nil).Once()
		authRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*model.AuthSession")).Return(nil).Once()
		rbac.On("RolesForUser", mock.Anything, uint(7)).Return([]string{"admin"}, nil).Once()
		rbac.On("PermissionsForUser", mock.Anything, uint(7)).Return([]string{"posts:read"}, nil).Once()

		rc := jwt.RegisteredClaims{
			Subject:   "7",
			Issuer:    "iss",
			Audience:  []string{"aud"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		}
		j.On("DefaultRegistered", "7", 10*time.Minute).Return(rc).Once()
		j.On("IssueAccessToken", mock.AnythingOfType("dto.AccessClaims")).Return("access", nil).Once()
		authRepo.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(nil).Once()

		s := NewAuthService(users, authRepo, nil, rbac, j, nil, nil, 10, 30, 5, "pepper", zap.NewNop())
		res, err := s.Login(ctx, "a@b.com", "12345678", dto.LoginMeta{DeviceID: "dev1"})
		assert.NoError(t, err)
		assert.False(t, res.TwoFactorRequired)
		assert.Equal(t, "access", res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.Equal(t, "admin", res.User.Role)
		assert.Equal(t, []string{"posts:read"}, res.User.Permissions)

		users.AssertExpectations(t)
		authRepo.AssertExpectations(t)
		j.AssertExpectations(t)
		rbac.AssertExpectations(t)
	})
}

func bcryptHash(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TestAuthService_Register_ConcurrentSameEmail runs concurrent Register with the same email
// using a real UserRepository (in-memory SQLite). Expect exactly one success; others get ErrEmailTaken or DB duplicate.
// Run with: go test -race -run TestAuthService_Register_ConcurrentSameEmail ./internal/service/...
func TestAuthService_Register_ConcurrentSameEmail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Real in-memory DB so unique constraint is enforced
	db := openAuthServiceTestDB(t)
	userRepo := newAuthServiceUserRepoFromDB(db)
	// No RBAC so Register only does FindByEmail + Create
	svc := NewAuthService(userRepo, &mockAuthRepo{}, nil, nil, &mockJWT{}, nil, nil, 10, 30, 5, "pepper", zap.NewNop())

	const concurrency = 15
	email := "concurrent-register@example.com"
	results := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			_, err := svc.Register(ctx, "User", email, "password123")
			results <- err
		}()
	}
	var successCount int
	for i := 0; i < concurrency; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount, "expected exactly one successful Register with same email")
}

// openAuthServiceTestDB opens in-memory SQLite for auth service tests that need a real DB.
func openAuthServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + url.QueryEscape(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newAuthServiceUserRepoFromDB(db *gorm.DB) AuthUserRepo {
	return &authServiceUserRepoAdapter{db: db}
}

type authServiceUserRepoAdapter struct {
	db *gorm.DB
}

func (a *authServiceUserRepoAdapter) Create(ctx context.Context, u *model.User) error {
	return a.db.WithContext(ctx).Create(u).Error
}
func (a *authServiceUserRepoAdapter) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := a.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
func (a *authServiceUserRepoAdapter) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	err := a.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
func (a *authServiceUserRepoAdapter) UpdatePassword(ctx context.Context, userID uint, newHash string) error {
	return a.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("password", newHash).Error
}
func (a *authServiceUserRepoAdapter) UpdateEmail(ctx context.Context, userID uint, newEmail string) error {
	return a.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("email", newEmail).Error
}
