package config

import "testing"

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "root")
	t.Setenv("DB_NAME", "blog")
	t.Setenv("JWT_ISSUER", "iss")
	t.Setenv("JWT_AUDIENCE", "aud")
	t.Setenv("JWT_KEY_ID", "k1")
	t.Setenv("CASBIN_MODEL_PATH", "configs/casbin_model.conf")
	t.Setenv("TWO_FACTOR_ENC_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("TWO_FACTOR_ISSUER", "iss")
	// Media requires object storage (S3 or GCS).
	t.Setenv("S3_ENDPOINT", "localhost:9000")
	t.Setenv("S3_ACCESS_KEY", "a")
	t.Setenv("S3_SECRET_KEY", "b")
	t.Setenv("S3_BUCKET", "media")
}

func TestLoad_ServerPort_UsesPORTWhenSet(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("SERVER_PORT", "8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServerPort != "9090" {
		t.Fatalf("ServerPort = %q, want %q", cfg.ServerPort, "9090")
	}
}

func TestLoad_ServerPort_FallsBackToServerPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PORT", "")
	t.Setenv("SERVER_PORT", "8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServerPort != "8081" {
		t.Fatalf("ServerPort = %q, want %q", cfg.ServerPort, "8081")
	}
}

func TestLoad_MediaStorage_InfersS3(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("MEDIA_STORAGE", "")
	t.Setenv("GCS_BUCKET", "")
	t.Setenv("S3_ENDPOINT", "localhost:9000")
	t.Setenv("S3_ACCESS_KEY", "a")
	t.Setenv("S3_SECRET_KEY", "b")
	t.Setenv("S3_BUCKET", "media")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MediaStorage != "s3" {
		t.Fatalf("MediaStorage = %q, want s3", cfg.MediaStorage)
	}
}

func TestLoad_MediaStorage_InfersGCS(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("MEDIA_STORAGE", "")
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("S3_ACCESS_KEY", "")
	t.Setenv("S3_SECRET_KEY", "")
	t.Setenv("S3_BUCKET", "")
	t.Setenv("GCS_BUCKET", "my-bucket")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MediaStorage != "gcs" {
		t.Fatalf("MediaStorage = %q, want gcs", cfg.MediaStorage)
	}
}

func TestLoad_MediaStorage_ConflictRequiresExplicit(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("MEDIA_STORAGE", "")
	t.Setenv("S3_ENDPOINT", "localhost:9000")
	t.Setenv("S3_ACCESS_KEY", "a")
	t.Setenv("S3_SECRET_KEY", "b")
	t.Setenv("S3_BUCKET", "media")
	t.Setenv("GCS_BUCKET", "other")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error when both S3 and GCS configured without MEDIA_STORAGE")
	}
}

func TestLoad_MediaStorage_LocalRejected(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("MEDIA_STORAGE", "local")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for MEDIA_STORAGE=local")
	}
}

func TestLoad_MediaStorage_NoBackendRejected(t *testing.T) {
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "root")
	t.Setenv("DB_NAME", "blog")
	t.Setenv("JWT_ISSUER", "iss")
	t.Setenv("JWT_AUDIENCE", "aud")
	t.Setenv("JWT_KEY_ID", "k1")
	t.Setenv("CASBIN_MODEL_PATH", "configs/casbin_model.conf")
	t.Setenv("TWO_FACTOR_ENC_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("TWO_FACTOR_ISSUER", "iss")
	t.Setenv("MEDIA_STORAGE", "")
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("S3_ACCESS_KEY", "")
	t.Setenv("S3_SECRET_KEY", "")
	t.Setenv("S3_BUCKET", "")
	t.Setenv("GCS_BUCKET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error when no S3 or GCS configured")
	}
}
