# Blog REST API (Gin + GORM + MySQL)

Production-ready local blog API with clean architecture:

- `cmd/`: entrypoint
- `internal/`: config, db, models, repos, services, handlers, middleware
- `pkg/`: shared response + logger

## Run

1. Create a MySQL database named `blog` (or set `DB_NAME`).
2. Copy `.env.example` to `.env` and update DB credentials.
3. Start:

```bash
go mod tidy
go run cmd/main.go
```

Auto migration runs on startup.

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

## Swagger

After starting the server, open:

- `http://localhost:8080/swagger/index.html`

## Auth (fintech-style JWT)

This API uses **RS256** access tokens and **refresh token rotation**.

### Token lifecycle
- **Access token**: short-lived (`ACCESS_TOKEN_TTL_MINUTES`, 5–15 minutes). Contains `iss`, `aud`, `exp`, `iat`, `nbf`, `jti`, plus app claims: `user_id`, `role`, `permissions`, `session_id`, `device_id`.
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

### Database tables (recommended)
- `users`: includes `role`, `permissions`
- `auth_sessions`: per device/session, revocable
- `refresh_tokens`: hashed, rotated, reuse-detectable
- `revoked_jtis`: blacklist for revoked access tokens
- `impersonation_audits`: immutable audit trail

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

## API

### Register

```bash
curl -X POST localhost:8080/api/register `
  -H "Content-Type: application/json" `
  -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"password\":\"password123\"}"
```

### Login

```bash
curl -X POST localhost:8080/api/login `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"alice@example.com\",\"password\":\"password123\",\"deviceId\":\"device-1\"}"
```

Copy `access_token` as `%ACCESS_TOKEN%` and `refresh_token` as `%REFRESH_TOKEN%`.

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
curl -X POST localhost:8080/api/posts `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"title\":\"Hello World\",\"content\":\"My first post\",\"categoryId\":1}"
```

### List Posts (cursor pagination, next/prev, no COUNT)

- First page:

```bash
curl "localhost:8080/api/posts?limit=10"
```

- Next page:

```bash
curl "localhost:8080/api/posts?limit=10&cursor=%NEXT_CURSOR%&dir=next"
```

- Prev page:

```bash
curl "localhost:8080/api/posts?limit=10&cursor=%PREV_CURSOR%&dir=prev"
```

### Get Post by Slug

```bash
curl "localhost:8080/api/v1/posts/slug/hello-world"
```

### Update Post (auth, owner only)

```bash
curl -X PUT localhost:8080/api/posts/1 `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"title\":\"Updated title\",\"content\":\"Updated content\"}"
```

### Delete Post (auth, owner only)

```bash
curl -X DELETE localhost:8080/api/posts/1 `
  -H "Authorization: Bearer %ACCESS_TOKEN%"
```

### Add Comment (auth)

```bash
curl -X POST localhost:8080/api/posts/1/comments `
  -H "Authorization: Bearer %ACCESS_TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"content\":\"Nice post\"}"
```

### List Comments

```bash
curl "localhost:8080/api/posts/1/comments?limit=50"
```

