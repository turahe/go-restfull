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
