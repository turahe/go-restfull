# Go REST Blog API

[![Test](https://github.com/turahe/go-restfull/actions/workflows/test.yml/badge.svg)](https://github.com/turahe/go-restfull/actions/workflows/test.yml)
![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)

## 1. Project Title

**Go REST Blog API** - a production-oriented REST backend built with Go, Gin, and GORM.

## 2. Description

This project provides a secure, modular, and testable backend for blog-style platforms.  
It includes authentication, RBAC authorization, content management (posts, categories, comments), media uploads, and operational tooling (testing and CI).

It is designed for:

- engineers building blog or CMS APIs in Go
- teams needing JWT + refresh token rotation + optional 2FA
- projects that value clean architecture, testing, and CI quality gates

## 3. Tech Stack

- **Language:** Go
- **HTTP Framework:** Gin
- **ORM:** GORM
- **Database:** MySQL
- **Cache / Rate Limiting:** Redis
- **Auth:** JWT (RS256), refresh token rotation, TOTP 2FA
- **Authorization:** Casbin (RBAC)
- **Object Storage:** MinIO (S3-compatible)
- **API Docs:** Swagger (swaggo)
- **Testing:** Go test, integration tests, race detector, benchmarks
- **CI:** GitHub Actions

## 4. Features

- RS256 JWT authentication with short-lived access tokens
- refresh token rotation with reuse detection
- revocation support for sessions and JTIs
- optional TOTP 2FA login challenge flow
- RBAC authorization (Casbin + DB-backed role/permission model)
- impersonation flow with audit trail
- blog domain CRUD: users, roles, permissions, categories, tags, posts, comments
- media upload and attachment to entities (`post`, `user`, `category`, `comment`)
- Redis-backed and in-memory rate limiting
- Swagger documentation endpoint
- strong test coverage: unit, integration, benchmark, and concurrency tests

## 5. Project Structure

```text
.
├── cmd/                  # Entrypoints and CLI commands
├── internal/
│   ├── config/           # Environment/config loading and validation
│   ├── database/         # DB/Redis connections + migrations
│   ├── handler/          # HTTP handlers and router wiring
│   ├── middleware/       # Auth, RBAC, logging, request-id, rate limit, recovery
│   ├── model/            # GORM models
│   ├── rbac/             # Casbin enforcer integration
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic layer
│   └── seeder/           # Seed data utilities
├── pkg/                  # Shared utility packages (logger, response, ids)
├── docs/                 # Generated Swagger artifacts
└── .github/workflows/    # CI pipelines
```

## 6. Installation & Setup

```bash
git clone git@github.com:turahe/go-restfull.git
cd go-restfull
go mod tidy
cp .env.example .env
go run cmd/main.go
```

Set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` in `.env` to PEM text (quoted multiline); see `.env.example`.

### Local run notes

1. Create a MySQL database (default: `blog`) or set `DB_NAME` in `.env`.
2. Configure JWT keys in `.env` as PEM text; see `.env.example`.
3. Update `.env` for DB, Redis, and object storage (S3-compatible or GCS) for media uploads.
4. Auto-migration runs on startup.
5. Swagger UI is available at `http://localhost:8080/swagger/index.html`.

Seed default RBAC and app settings (idempotent):

```bash
go run ./cmd seed rbac
go run ./cmd seed settings
```

## Configuration

Important environment variables:

- **Database:** `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **Redis:** `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`
- **JWT:** `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY` (PEM only), `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_KEY_ID`
- **Token TTLs:** `ACCESS_TOKEN_TTL_MINUTES`, `REFRESH_TOKEN_TTL_DAYS`, `IMPERSONATION_TTL_MINUTES`
- **2FA:** `TWO_FACTOR_ENC_KEY`, `TWO_FACTOR_ISSUER`
- **Media (object storage, required):** `MEDIA_STORAGE` (`s3` or `gcs`), `MEDIA_MAX_UPLOAD_BYTES`, plus either S3-compatible (`S3_*` or legacy `MINIO_*`) or `GCS_BUCKET` with Application Default Credentials.

See `.env.example` for complete defaults.

## Docker

The compose setup can run MySQL, Redis, MinIO (S3 API), and the API. **`docker-compose.yml` declares an `environment` block** for each service; **keys match the variable names in your project `.env`** (same as `.env.example`), and values are filled with **`${VAR}`** substitution from that file. Copy `.env.example` to `.env`, add JWT PEM keys and any secrets (`REFRESH_TOKEN_PEPPER`, `TWO_FACTOR_ENC_KEY`), then run compose. Media is stored only in **object storage** (S3-compatible or GCS), not on the container filesystem; configure `MINIO_*` or `S3_*` to point at the Compose MinIO service (or another bucket). For local `go run` without Docker, use `DB_HOST=127.0.0.1` and `REDIS_ADDR=127.0.0.1:6379`, and point S3/MinIO env vars at a reachable endpoint (e.g. local MinIO).

```bash
# MySQL only
docker compose up -d mysql

# Full stack
docker compose up -d --build

# Stop
docker compose down
```

## Google Cloud Run

This project supports Cloud Run container deployment.

### Why it works

- The app now honors Cloud Run's injected `PORT` variable (with fallback to `SERVER_PORT`).
- `Dockerfile` already builds a Linux container and starts the API (`/app/api serve`).

### Deploy with gcloud

```bash
gcloud run deploy go-rest-api \
  --source . \
  --region asia-southeast2 \
  --platform managed \
  --allow-unauthenticated \
  --set-env-vars "APP_ENV=prod,DB_DRIVER=mysql,DB_HOST=<db-host>,DB_PORT=3306,DB_USER=<db-user>,DB_PASSWORD=<db-pass>,DB_NAME=<db-name>,JWT_PRIVATE_KEY=<pem-private-key>,JWT_PUBLIC_KEY=<pem-public-key>,JWT_ISSUER=go-rest-blog,JWT_AUDIENCE=blog-api,JWT_KEY_ID=k1,TWO_FACTOR_ENC_KEY=<32-byte-key>,TWO_FACTOR_ISSUER=go-rest-blog,MEDIA_STORAGE=gcs,GCS_BUCKET=<bucket-name>"
```

Cloud SQL for MySQL (recommended on GCP):
```bash
gcloud run deploy go-rest-api \
  --source . \
  --region asia-southeast2 \
  --platform managed \
  --allow-unauthenticated \
  --set-env-vars "APP_ENV=prod,DB_DRIVER=mysql-cloud,INSTANCE_CONNECTION_NAME=<project:region:instance>,PRIVATE_IP=false,DB_USER=<db-user>,DB_PASSWORD=<db-pass>,DB_NAME=<db-name>,JWT_PRIVATE_KEY=<pem-private-key>,JWT_PUBLIC_KEY=<pem-public-key>,JWT_ISSUER=go-rest-blog,JWT_AUDIENCE=blog-api,JWT_KEY_ID=k1,TWO_FACTOR_ENC_KEY=<32-byte-key>,TWO_FACTOR_ISSUER=go-rest-blog,MEDIA_STORAGE=gcs,GCS_BUCKET=<bucket-name>"
```

### Deploy via Cloud Build

`cloudbuild.yaml` is included for CI/CD deployment to Cloud Run:

```bash
gcloud builds submit --config cloudbuild.yaml
```

## Makefile Commands

```bash
make docker-up
make docker-down
make swagger
make test
```

## Authentication and Security Model

### JWT and refresh lifecycle

- Access token: short-lived (`ACCESS_TOKEN_TTL_MINUTES`, 5-15 min)
- Refresh token: long-lived (`REFRESH_TOKEN_TTL_DAYS`), stored **hashed**
- On `/auth/refresh`: old token marked `used_at`, new refresh + access tokens issued
- Reuse detection revokes the full session/family on replay attempts
- Revocation supports:
  - session revocation (`auth_sessions.revoked_at`)
  - access token blacklist (`revoked_jtis`) until expiration

### 2FA (TOTP)

- Optional TOTP-based second factor.
- Login flow:
  1. user logs in with email/password + `deviceId`
  2. if 2FA disabled -> tokens returned
  3. if 2FA enabled -> challenge response returned
  4. client verifies with `/api/v1/auth/2fa/verify`

2FA management endpoints (authenticated):

- `POST /api/v1/auth/2fa/setup`
- `POST /api/v1/auth/2fa/enable`

### Impersonation

- Allowed roles: `admin`, `support`
- Issues short-lived impersonation token (default 5 min)
- Token carries:
  - `impersonation=true`
  - `impersonator_id`
  - `impersonated_user_id`
  - `impersonation_reason`
- Every impersonation action is recorded in immutable audit logs

## Public settings

Unauthenticated clients can load non-secret configuration (JWT issuer/audience/key id, token TTLs, upload size limit, rate-limit hints, feature flags):

- `GET /api/v1/settings`

Database-backed keys live in the `settings` table (`setting_key`, `value`, `is_public`). Rows with `is_public = true` are returned directly as the response `data` object as a string map (`setting_key` → `value`). Use `SettingRepository` (or raw SQL) to manage rows; admin HTTP CRUD can be added later with RBAC.

## Media Storage

- Media files are stored in **object storage only** (S3-compatible backends such as MinIO or AWS S3, or **GCS**)
- Metadata is stored in MySQL
- `GET /api/v1/media/:id` returns a presigned `downloadUrl` from the configured bucket
- `POST /api/v1/media` supports multipart upload:
  - `file`
  - optional `mediaableType` + `mediaableId`
- Allowed `mediaableType`: `post`, `user`, `category`, `comment`

## Data Model Overview

Core tables include:

- `users`
- `roles`, `permissions`, `user_roles`, `role_permissions`
- `auth_sessions`, `refresh_tokens`, `revoked_jtis`, `impersonation_audits`
- `user_two_factors`, `two_factor_challenges`
- `media`, `post_media`, `user_media`, `category_media`, `comment_media`, `mediable`
- `settings` (key/value app settings; `is_public` controls exposure on `GET /settings`)

## Testing

- **Unit tests**
  ```bash
  go test ./...
  ```
  Repository tests use in-memory SQLite; handler/service tests use mocks.

- **Integration tests (real MySQL)**
  ```bash
  go test -tags=integration ./internal/repository/...
  ```
  Each integration test runs inside a transaction and rolls back.

- **Benchmarks**
  ```bash
  go test -bench=. -benchmem ./internal/repository/... ./internal/service/...
  ```

- **Race detector / concurrency safety**
  ```bash
  go test -race ./internal/repository/... ./internal/service/...
  ```
  > Requires CGO-enabled Go toolchain.

## CI

GitHub Actions runs:

- build
- unit tests
- integration tests (MySQL)
- Redis rate limiter test
- race checks

## Windows Note

If you encounter:

`compile: version "goX.Y.Z" does not match go tool version ...`

your `GOROOT` likely points to a different Go installation than your active `go.exe`.  
Fix by aligning or unsetting `GOROOT`.