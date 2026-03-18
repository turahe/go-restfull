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

	RateLimitRPS   float64
RateLimitBurst int

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RateLimitKeyPrefix string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTPublicKeyPath  string
	JWTPrivateKeyPath string
	JWTIssuer         string
	JWTAudience       string
	JWTKeyID          string

	AccessTokenTTLMinutes        int
	RefreshTokenTTLDays          int
	ImpersonationTTLMinutes      int
	RefreshTokenPepper           string

	CasbinModelPath string

	TwoFactorEncKey string
	TwoFactorIssuer string

	MediaUploadDir       string
	MediaMaxUploadBytes int64

	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucket     string
	MinioUseSSL      bool
}

func Load() (Config, error) {
	// Best-effort load for local development; ignore if missing.
	_ = godotenv.Load()

	cfg := Config{
		Env:                strings.TrimSpace(os.Getenv("APP_ENV")),
		ServerPort:         strings.TrimSpace(getEnvDefault("SERVER_PORT", "8080")),
		RateLimitRPS:       getEnvFloatDefault("RATE_LIMIT_RPS", 5),
		RateLimitBurst:     getEnvIntDefault("RATE_LIMIT_BURST", 10),
		RedisAddr:          strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:            getEnvIntDefault("REDIS_DB", 0),
		RateLimitKeyPrefix: strings.TrimSpace(getEnvDefault("RATE_LIMIT_KEY_PREFIX", "rl:ip:")),
		DBHost:             strings.TrimSpace(getEnvDefault("DB_HOST", "127.0.0.1")),
		DBPort:             strings.TrimSpace(getEnvDefault("DB_PORT", "3306")),
		DBUser:             strings.TrimSpace(getEnvDefault("DB_USER", "root")),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             strings.TrimSpace(getEnvDefault("DB_NAME", "blog")),
		JWTPublicKeyPath:   strings.TrimSpace(getEnvDefault("JWT_PUBLIC_KEY_PATH", "keys/jwtRS256.key.pub")),
		JWTPrivateKeyPath:  strings.TrimSpace(getEnvDefault("JWT_PRIVATE_KEY_PATH", "keys/jwtRS256.key")),
		JWTIssuer:          strings.TrimSpace(getEnvDefault("JWT_ISSUER", "go-rest-blog")),
		JWTAudience:        strings.TrimSpace(getEnvDefault("JWT_AUDIENCE", "blog-api")),
		JWTKeyID:           strings.TrimSpace(getEnvDefault("JWT_KEY_ID", "k1")),

		AccessTokenTTLMinutes:   getEnvIntDefault("ACCESS_TOKEN_TTL_MINUTES", 10),
		RefreshTokenTTLDays:     getEnvIntDefault("REFRESH_TOKEN_TTL_DAYS", 30),
		ImpersonationTTLMinutes: getEnvIntDefault("IMPERSONATION_TTL_MINUTES", 5),
		RefreshTokenPepper:      os.Getenv("REFRESH_TOKEN_PEPPER"),
		CasbinModelPath:         strings.TrimSpace(getEnvDefault("CASBIN_MODEL_PATH", "configs/casbin_model.conf")),
		TwoFactorEncKey:         strings.TrimSpace(os.Getenv("TWO_FACTOR_ENC_KEY")),
		TwoFactorIssuer:         strings.TrimSpace(getEnvDefault("TWO_FACTOR_ISSUER", "")),
		MediaUploadDir:          strings.TrimSpace(getEnvDefault("MEDIA_UPLOAD_DIR", "uploads")),
		MediaMaxUploadBytes:     getEnvInt64Default("MEDIA_MAX_UPLOAD_BYTES", 10*1024*1024),

		MinioEndpoint:  strings.TrimSpace(os.Getenv("MINIO_ENDPOINT")),
		MinioAccessKey: strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY")),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    strings.TrimSpace(getEnvDefault("MINIO_BUCKET", "media")),
		MinioUseSSL:     getEnvBoolDefault("MINIO_USE_SSL", false),
	}

	if cfg.DBName == "" || cfg.DBUser == "" || cfg.DBHost == "" || cfg.DBPort == "" {
		return Config{}, errors.New("missing required DB configuration")
	}
	if cfg.AccessTokenTTLMinutes < 5 || cfg.AccessTokenTTLMinutes > 15 {
		return Config{}, errors.New("ACCESS_TOKEN_TTL_MINUTES must be between 5 and 15")
	}
	if cfg.RefreshTokenTTLDays < 1 || cfg.RefreshTokenTTLDays > 365 {
		return Config{}, errors.New("REFRESH_TOKEN_TTL_DAYS must be between 1 and 365")
	}
	if cfg.ImpersonationTTLMinutes < 1 || cfg.ImpersonationTTLMinutes > 10 {
		return Config{}, errors.New("IMPERSONATION_TTL_MINUTES must be between 1 and 10")
	}
	if cfg.JWTIssuer == "" || cfg.JWTAudience == "" || cfg.JWTKeyID == "" {
		return Config{}, errors.New("JWT_ISSUER, JWT_AUDIENCE, JWT_KEY_ID are required")
	}
	if cfg.CasbinModelPath == "" {
		return Config{}, errors.New("CASBIN_MODEL_PATH is required")
	}
	if cfg.RateLimitRPS < 0 {
		return Config{}, errors.New("RATE_LIMIT_RPS must be >= 0")
	}
	if cfg.RateLimitBurst < 0 {
		return Config{}, errors.New("RATE_LIMIT_BURST must be >= 0")
	}
	if cfg.RedisDB < 0 {
		return Config{}, errors.New("REDIS_DB must be >= 0")
	}
	if cfg.TwoFactorEncKey == "" {
		return Config{}, errors.New("TWO_FACTOR_ENC_KEY is required")
	}
	if cfg.TwoFactorIssuer == "" {
		cfg.TwoFactorIssuer = cfg.JWTIssuer
	}
	if cfg.MediaUploadDir == "" {
		cfg.MediaUploadDir = "uploads"
	}
	if cfg.MediaMaxUploadBytes <= 0 {
		return Config{}, errors.New("MEDIA_MAX_UPLOAD_BYTES must be > 0")
	}

	// If MinIO is partially configured, treat it as disabled to avoid breaking local dev.
	// (We only enable when endpoint + access + secret are all set.)
	if cfg.MinioEndpoint == "" || cfg.MinioAccessKey == "" || cfg.MinioSecretKey == "" {
		cfg.MinioEndpoint = ""
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

func getEnvInt64Default(key string, def int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func getEnvBoolDefault(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func getEnvFloatDefault(key string, def float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}

