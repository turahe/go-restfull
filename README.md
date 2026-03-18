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
  -d "{\"email\":\"alice@example.com\",\"password\":\"password123\"}"
```

Copy the `token` from response to use below as `%TOKEN%`.

### Create Post (auth)

```bash
curl -X POST localhost:8080/api/posts `
  -H "Authorization: Bearer %TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"title\":\"Hello World\",\"content\":\"My first post\"}"
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
  -H "Authorization: Bearer %TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"title\":\"Updated title\",\"content\":\"Updated content\"}"
```

### Delete Post (auth, owner only)

```bash
curl -X DELETE localhost:8080/api/posts/1 `
  -H "Authorization: Bearer %TOKEN%"
```

### Add Comment (auth)

```bash
curl -X POST localhost:8080/api/posts/1/comments `
  -H "Authorization: Bearer %TOKEN%" `
  -H "Content-Type: application/json" `
  -d "{\"content\":\"Nice post\"}"
```

### List Comments

```bash
curl "localhost:8080/api/posts/1/comments?limit=50"
```

