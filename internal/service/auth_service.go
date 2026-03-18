package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"go-rest/internal/model"
	"go-rest/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
)

type AuthService struct {
	users *repository.UserRepository
	jwt   *JWTManager
}

func NewAuthService(users *repository.UserRepository, jwtm *JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwtm}
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
	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": s.jwt.issuer,
		"sub": u.ID,
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(s.jwt.ttlMinutes) * time.Minute).Unix(),
		"email": u.Email,
		"name":  u.Name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.jwt.privateKey)
	if err != nil {
		return "", nil, err
	}
	return signed, u, nil
}

