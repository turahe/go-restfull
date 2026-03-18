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

## Notes (Windows)

If you ever see errors like `compile: version "goX.Y.Z" does not match go tool version ...`, your `GOROOT` is likely pointing at a different Go installation than your `go.exe`. Fix by ensuring `GOROOT` matches `go version`, or by unsetting `GOROOT`.

## API (examples)

### Register

```bash
curl -X POST localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"password\":\"password123\"}"
```

### Login

```bash
curl -X POST localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"alice@example.com\",\"password\":\"password123\",\"deviceId\":\"device-1\"}"
```

If `twoFactorRequired=false`, copy `accessToken` as `%ACCESS_TOKEN%` and `refreshToken` as `%REFRESH_TOKEN%`.
If `twoFactorRequired=true`, call `/api/v1/auth/2fa/verify`:

```bash
curl -X POST localhost:8080/api/v1/auth/2fa/verify `
  -H "Content-Type: application/json" `
  -d "{\"challengeId\":\"...\",\"code\":\"123456\",\"deviceId\":\"device-1\"}"
```

### Refresh (rotation)

```bash
curl -X POST localhost:8080/api/v1/auth/refresh `
  -H "Content-Type: application/json" `
  -d "{\"refreshToken\":\"%REFRESH_TOKEN%\",\"deviceId\":\"device-1\"}"
```

### Impersonate (admin/support only)

```bash
curl -X POST localhost:8080/api/v1/auth/impersonate `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"userId\":2,\"reason\":\"Support investigation\",\"deviceId\":\"device-1\"}"
```

### Create Post (auth)

```bash
curl -X POST localhost:8080/api/v1/posts `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"title\":\"Hello World\",\"content\":\"My first post\",\"categoryId\":1}"
```

### List Posts (cursor pagination, next/prev, no COUNT)

- First page:

```bash
curl "localhost:8080/api/v1/posts?limit=10"
```

- Next page:

```bash
curl "localhost:8080/api/v1/posts?limit=10&cursor=%NEXT_CURSOR%&dir=next"
```

- Prev page:

```bash
curl "localhost:8080/api/v1/posts?limit=10&cursor=%PREV_CURSOR%&dir=prev"
```

### Get Post by Slug

```bash
curl "localhost:8080/api/v1/posts/slug/hello-world"
```

### Update Post (auth, owner only)

```bash
curl -X PUT localhost:8080/api/v1/posts/1 `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"title\":\"Updated title\",\"content\":\"Updated content\"}"
```

### Delete Post (auth, owner only)

```bash
curl -X DELETE localhost:8080/api/v1/posts/1 `
  -H "Authorization: Bearer %ACCESS_TOKEN%"
```

### Add Comment (auth)

```bash
curl -X POST localhost:8080/api/v1/posts/1/comments `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"content\":\"Nice post\"}"
```

### List Comments

```bash
curl "localhost:8080/api/v1/posts/1/comments?limit=50"
```

