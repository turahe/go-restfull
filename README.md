# Blog REST API (Gin + GORM + MySQL)

Production-ready local blog API with clean architecture:

- `cmd/`: entrypoint
- `internal/`: config, db, models, repos, services, handlers, middleware
- `pkg/`: shared response + logger

## Run

1. Create a MySQL database named `blog` (or set `DB_NAME`).
2. Copy `.env.example` to `.env` and update DB credentials (including Redis and JWT paths).
3. Start:

```bash
go mod tidy
go run ./cmd
```

Auto migration runs on startup. Swagger is available at `http://localhost:8080/swagger/index.html`.

## Docker
This compose file also starts **MinIO** for media storage.

- Start MySQL only:

```bash
docker compose up -d mysql
```

- Start MySQL + API:

```bash
docker compose up -d --build
```

- Stop:

```bash
docker compose down
```

## Makefile

```bash
make docker-up
make docker-down
make swagger
make test
```

## Auth (fintech-style JWT + 2FA)

This API uses **RS256** access tokens and **refresh token rotation**.

### Token lifecycle
- **Access token**: short-lived (`ACCESS_TOKEN_TTL_MINUTES`, 5–15 minutes). Contains `iss`, `aud`, `exp`, `iat`, `nbf`, `jti`, plus app claims: `userId`, `role`, `permissions`, `sessionId`, `deviceId`.
- **Refresh token**: long-lived (`REFRESH_TOKEN_TTL_DAYS`). Stored **hashed** in MySQL.
- **Rotation**: every `/auth/refresh` call:
  - marks the old refresh token as `used_at`
  - issues a new refresh token
  - issues a new access token
- **Reuse detection**: if an already-used refresh token is presented again, the system treats it as theft/replay and **revokes the session** and **refresh token family**.
- **Revocation**:
  - sessions can be revoked (`auth_sessions.revoked_at`)
  - individual access tokens can be revoked via `revoked_jtis` until `exp`

### Key rotation strategy (recommended)
- Keep a **`kid`** header on tokens (already implemented).
- Store keys in a secret manager (private key) and distribute public keys to services.
- During rotation:
  - deploy new key pair (new `kid`) and start signing with it
  - keep old public key available for verification until all old access tokens expire

### Database tables
- `users`
- `roles`, `permissions`, `user_roles`, `role_permissions` (RBAC via Casbin)
- `auth_sessions`: per device/session, revocable
- `refresh_tokens`: hashed, rotated, reuse-detectable
- `revoked_jtis`: blacklist for revoked access tokens
- `impersonation_audits`: immutable audit trail
- `user_two_factors`, `two_factor_challenges`: TOTP 2FA config and login challenges
- `media`: file metadata (stored in MinIO when configured)
- `post_media`, `user_media`, `category_media`, `comment_media`: media join tables for relations
- `mediable`: single-table design for future consolidation (auto-migrated)

### 2FA (TOTP)

- Optional **TOTP** (Google Authenticator, etc.) per user.
- Flow:
  1. User logs in with email/password + `deviceId`.
  2. If 2FA **disabled** → returns `accessToken` + `refreshToken` as usual.
  3. If 2FA **enabled** → returns:
     ```json
     {
       "twoFactorRequired": true,
       "challengeId": "...",
       "expiresAt": "...",
       "sessionId": "...",
       "user": { "id", "name", "email" }
     }
     ```
  4. Client then calls `/api/v1/auth/2fa/verify` with `challengeId`, `code`, `deviceId` to obtain tokens.

2FA management endpoints (auth required):
- `POST /api/v1/auth/2fa/setup` → returns `secret` and `otpauthUrl` for TOTP app.
- `POST /api/v1/auth/2fa/enable` → body: `{ "code": "123456" }`.

## Impersonation
Admin/support users can impersonate a user with a **short-lived (default 5 min)** access token.

Rules:
- Allowed roles: `admin`, `support`
- Token contains:
  - `impersonation=true`
  - `impersonator_id`
  - `impersonated_user_id`
  - `impersonation_reason`
- Every impersonation creates an audit record with IP/UA/timestamp.

## Media (MinIO + relations)
Media is stored as objects in **MinIO** (S3-compatible) and metadata is stored in MySQL.

### Storage configuration
Set these env vars (already present in `.env.example`):
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_BUCKET`
- `MINIO_USE_SSL`

When MinIO is enabled, `GET /api/v1/media/:id` returns a `downloadUrl` (presigned).

### Attach media to entities
`POST /api/v1/media` accepts multipart upload with:
- `file` (form-data file)
- optional `mediaableType` and `mediaableId`

Allowed `mediaableType` values: `post`, `user`, `category`, `comment`.
If omitted, media defaults to attaching to the uploading `user`.

## Notes (Windows)

If you ever see errors like `compile: version "goX.Y.Z" does not match go tool version ...`, your `GOROOT` is likely pointing at a different Go installation than your `go.exe`. Fix by ensuring `GOROOT` matches `go version`, or by unsetting `GOROOT`.