package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env string

	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTPublicKeyPath  string
	JWTPrivateKeyPath string
	JWTIssuer         string
	JWTTTLMinutes     int
}

func Load() (Config, error) {
	// Best-effort load for local development; ignore if missing.
	_ = godotenv.Load()

	cfg := Config{
		Env:                strings.TrimSpace(os.Getenv("APP_ENV")),
		ServerPort:         strings.TrimSpace(getEnvDefault("SERVER_PORT", "8080")),
		DBHost:             strings.TrimSpace(getEnvDefault("DB_HOST", "127.0.0.1")),
		DBPort:             strings.TrimSpace(getEnvDefault("DB_PORT", "3306")),
		DBUser:             strings.TrimSpace(getEnvDefault("DB_USER", "root")),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             strings.TrimSpace(getEnvDefault("DB_NAME", "blog")),
		JWTPublicKeyPath:   strings.TrimSpace(getEnvDefault("JWT_PUBLIC_KEY_PATH", "keys/jwtRS256.key.pub")),
		JWTPrivateKeyPath:  strings.TrimSpace(getEnvDefault("JWT_PRIVATE_KEY_PATH", "keys/jwtRS256.key")),
		JWTIssuer:          strings.TrimSpace(getEnvDefault("JWT_ISSUER", "go-rest-blog")),
		JWTTTLMinutes:      getEnvIntDefault("JWT_TTL_MINUTES", 120),
	}

	if cfg.DBName == "" || cfg.DBUser == "" || cfg.DBHost == "" || cfg.DBPort == "" {
		return Config{}, errors.New("missing required DB configuration")
	}
	if cfg.JWTTTLMinutes <= 0 {
		return Config{}, errors.New("JWT_TTL_MINUTES must be > 0")
	}
	return cfg, nil
}

func getEnvDefault(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getEnvIntDefault(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

