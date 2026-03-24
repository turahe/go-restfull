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

	// SwaggerEnabled registers /swagger when true. If SWAGGER_ENABLED is unset, defaults to true only when APP_ENV is "local".
	SwaggerEnabled bool

	ServerPort string

	RateLimitRPS   float64
	RateLimitBurst int

	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	RateLimitKeyPrefix string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	// DBDriver selects the MySQL connection implementation:
	// - "mysql": direct TCP host/port DSN
	// - "mysql-cloud": use Cloud SQL Go Connector (cloud.google.com/go/cloudsqlconn)
	DBDriver string

	// Cloud SQL fields (used when DBDriver="mysql-cloud")
	CloudSQLInstanceConnectionName string
	CloudSQLPrivateIP              bool

	JWTPublicKey  string
	JWTPrivateKey string
	JWTIssuer     string
	JWTAudience   string
	JWTKeyID      string

	AccessTokenTTLMinutes   int
	RefreshTokenTTLDays     int
	ImpersonationTTLMinutes int
	RefreshTokenPepper      string

	CasbinModelPath string

	TwoFactorEncKey string
	TwoFactorIssuer string

	MediaMaxUploadBytes int64

	// MediaStorage selects where uploads go: s3 (S3 API: AWS S3, MinIO, etc.) or gcs (Google Cloud Storage).
	MediaStorage string

	// S3-compatible object storage (AWS S3, MinIO, etc.). MINIO_* env vars are merged into these when S3_* are unset.
	S3Endpoint  string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool

	// Google Cloud Storage (native JSON API).
	GCSBucket string
}

func Load() (Config, error) {
	// Best-effort load for local development; ignore if missing.
	_ = godotenv.Load()

	cfg := Config{
		Env: strings.TrimSpace(os.Getenv("APP_ENV")),
		// Cloud Run injects PORT; prefer it over SERVER_PORT when present.
		ServerPort:                     strings.TrimSpace(getEnvFirstNonEmpty("PORT", "SERVER_PORT", "8080")),
		RateLimitRPS:                   getEnvFloatDefault("RATE_LIMIT_RPS", 5),
		RateLimitBurst:                 getEnvIntDefault("RATE_LIMIT_BURST", 10),
		RedisAddr:                      strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		RedisPassword:                  os.Getenv("REDIS_PASSWORD"),
		RedisDB:                        getEnvIntDefault("REDIS_DB", 0),
		RateLimitKeyPrefix:             strings.TrimSpace(getEnvDefault("RATE_LIMIT_KEY_PREFIX", "rl:ip:")),
		DBHost:                         strings.TrimSpace(getEnvDefault("DB_HOST", "127.0.0.1")),
		DBPort:                         strings.TrimSpace(getEnvDefault("DB_PORT", "3306")),
		DBUser:                         strings.TrimSpace(getEnvDefault("DB_USER", "root")),
		DBPassword:                     os.Getenv("DB_PASSWORD"),
		DBName:                         strings.TrimSpace(getEnvDefault("DB_NAME", "blog")),
		DBDriver:                       strings.TrimSpace(getEnvDefault("DB_DRIVER", "mysql")),
		CloudSQLInstanceConnectionName: strings.TrimSpace(os.Getenv("INSTANCE_CONNECTION_NAME")),
		CloudSQLPrivateIP:              getEnvBoolDefault("PRIVATE_IP", false),
		JWTPublicKey:                   strings.TrimSpace(os.Getenv("JWT_PUBLIC_KEY")),
		JWTPrivateKey:                  strings.TrimSpace(os.Getenv("JWT_PRIVATE_KEY")),
		JWTIssuer:                      strings.TrimSpace(getEnvDefault("JWT_ISSUER", "go-rest-blog")),
		JWTAudience:                    strings.TrimSpace(getEnvDefault("JWT_AUDIENCE", "blog-api")),
		JWTKeyID:                       strings.TrimSpace(getEnvDefault("JWT_KEY_ID", "k1")),

		AccessTokenTTLMinutes:   getEnvIntDefault("ACCESS_TOKEN_TTL_MINUTES", 10),
		RefreshTokenTTLDays:     getEnvIntDefault("REFRESH_TOKEN_TTL_DAYS", 30),
		ImpersonationTTLMinutes: getEnvIntDefault("IMPERSONATION_TTL_MINUTES", 5),
		RefreshTokenPepper:      os.Getenv("REFRESH_TOKEN_PEPPER"),
		CasbinModelPath:         strings.TrimSpace(getEnvDefault("CASBIN_MODEL_PATH", "configs/casbin_model.conf")),
		TwoFactorEncKey:         strings.TrimSpace(os.Getenv("TWO_FACTOR_ENC_KEY")),
		TwoFactorIssuer:         strings.TrimSpace(getEnvDefault("TWO_FACTOR_ISSUER", "")),
		MediaMaxUploadBytes:     getEnvInt64Default("MEDIA_MAX_UPLOAD_BYTES", 10*1024*1024),

		S3Endpoint:  strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
		S3Region:    strings.TrimSpace(os.Getenv("S3_REGION")),
		S3AccessKey: strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3Bucket:    strings.TrimSpace(getEnvDefault("S3_BUCKET", "")),
		S3UseSSL:    getEnvBoolDefault("S3_USE_SSL", false),

		GCSBucket: strings.TrimSpace(os.Getenv("GCS_BUCKET")),
	}

	// Merge legacy MINIO_* into S3 when S3_* are unset (MinIO is S3-compatible).
	if cfg.S3Endpoint == "" {
		cfg.S3Endpoint = strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	}
	if cfg.S3AccessKey == "" {
		cfg.S3AccessKey = strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	}
	if cfg.S3SecretKey == "" {
		cfg.S3SecretKey = os.Getenv("MINIO_SECRET_KEY")
	}
	if cfg.S3Bucket == "" {
		cfg.S3Bucket = strings.TrimSpace(getEnvDefault("MINIO_BUCKET", ""))
	}
	if cfg.S3Bucket == "" && cfg.S3Endpoint != "" && cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		cfg.S3Bucket = "media"
	}
	if strings.TrimSpace(os.Getenv("S3_USE_SSL")) == "" {
		cfg.S3UseSSL = getEnvBoolDefault("MINIO_USE_SSL", false)
	}

	if cfg.DBName == "" || cfg.DBUser == "" || cfg.DBHost == "" || cfg.DBPort == "" {
		return Config{}, errors.New("missing required DB configuration")
	}
	switch strings.ToLower(cfg.DBDriver) {
	case "mysql", "mysql-cloud", "":
		// ok (empty treated as defaulted)
	default:
		return Config{}, errors.New("DB_DRIVER must be 'mysql' or 'mysql-cloud'")
	}
	if strings.ToLower(cfg.DBDriver) == "mysql-cloud" && cfg.CloudSQLInstanceConnectionName == "" {
		return Config{}, errors.New("INSTANCE_CONNECTION_NAME is required when DB_DRIVER=mysql-cloud")
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
	if cfg.MediaMaxUploadBytes <= 0 {
		return Config{}, errors.New("MEDIA_MAX_UPLOAD_BYTES must be > 0")
	}

	ms := strings.ToLower(strings.TrimSpace(os.Getenv("MEDIA_STORAGE")))
	s3Ok := cfg.S3Endpoint != "" && cfg.S3AccessKey != "" && cfg.S3SecretKey != "" && cfg.S3Bucket != ""
	gcsOk := cfg.GCSBucket != ""
	if ms == "" {
		if s3Ok && gcsOk {
			return Config{}, errors.New("both S3 and GCS are configured; set MEDIA_STORAGE to s3 or gcs")
		}
		if s3Ok {
			ms = "s3"
		} else if gcsOk {
			ms = "gcs"
		} else {
			return Config{}, errors.New("media storage requires S3-compatible config (S3_* or MINIO_*) or GCS (GCS_BUCKET), or set MEDIA_STORAGE to s3 or gcs")
		}
	} else {
		switch ms {
		case "local", "filesystem", "fs", "disk":
			return Config{}, errors.New("MEDIA_STORAGE local filesystem is not supported; use s3 or gcs")
		case "minio", "aws":
			ms = "s3"
		case "google":
			ms = "gcs"
		}
	}
	if ms != "s3" && ms != "gcs" {
		return Config{}, errors.New("MEDIA_STORAGE must be s3 or gcs")
	}
	cfg.MediaStorage = ms

	if cfg.MediaStorage == "s3" {
		if cfg.S3Endpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" || cfg.S3Bucket == "" {
			return Config{}, errors.New("S3 storage requires S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY, and S3_BUCKET (or legacy MINIO_* equivalents)")
		}
	}
	if cfg.MediaStorage == "gcs" {
		if cfg.GCSBucket == "" {
			return Config{}, errors.New("GCS storage requires GCS_BUCKET")
		}
	}

	if strings.TrimSpace(os.Getenv("SWAGGER_ENABLED")) != "" {
		cfg.SwaggerEnabled = getEnvBoolDefault("SWAGGER_ENABLED", false)
	} else {
		cfg.SwaggerEnabled = strings.EqualFold(cfg.Env, "local")
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

func getEnvFirstNonEmpty(keys ...string) string {
	for _, k := range keys[:len(keys)-1] {
		v := strings.TrimSpace(os.Getenv(k))
		if v != "" {
			return v
		}
	}
	return keys[len(keys)-1]
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
