# pipebin.dev — CLAUDE.md

Minimal developer pastebin. Users pipe logs/code/stdout via curl or a browser form.
Two Go services: **API** (port 8001) and **Frontend** (port 8002).

---

## Tech stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25+ |
| Database driver | `github.com/jackc/pgx/v5` — pgxpool, raw SQL (no ORM) |
| Migrations | `github.com/golang-migrate/migrate/v4` — embedded SQL files |
| Validation | `github.com/go-ozzo/ozzo-validation/v4` |
| Logging | `log/slog` (stdlib) — text handler in dev, JSON handler in prod |
| ID generation | `github.com/matoous/go-nanoid/v2` |
| Frontend styles | oat.css CDN — do not replace; dark mode forced via `color-scheme: dark` |
| Build | Bazel (`bazel build //...`) **and** `go build ./...` both work |

---

## Service architecture

```
services/api/          — REST API (port 8001)
  cmd/main.go          — entry point; wires DB → repo → service → handler
  handler/             — HTTP handlers (pastes_handler.go)
  repository/          — pgx raw SQL (pastes_repository.go)
  internal/
    config/            — API env config (APP_PORT, POSTGRESQL_DSN, FRONTEND_URL, …)
    database/          — pgxpool connection + golang-migrate runner
    httpx/             — JSON response helpers + client IP extraction
    middleware/        — RequestID middleware
    server/            — mux / router
    services/          — business logic (CreatePaste, GetPasteByPublicID)
  migrations/          — embedded SQL migration files + embed.go

services/frontend/     — Frontend server (port 8002)
  cmd/main.go          — entry point; embeds templates + static, starts http.Server
  handlers/            — page handlers (Home, CreatePaste, Paste, RawPaste, NotFound)
  internal/
    config/            — frontend env config (FA_PORT, API_BASE_URL, LOGGER)
    server/            — mux / router
  templates/           — Go html/template files (layout, home, paste, upload_result, 404, error)
  static/              — style.css

libs/
  config/              — shared env helpers (GetEnv, MustGetEnv, LoadDotEnv)
  hash/                — SHA-256 IP hashing (GetSHA256Hash)
  logger/              — zap setup
  models/              — Paste, CreatePasteInput domain structs
```

---

## Environment variables

### API — `configs/.env`

```
APP_PORT=8001
POSTGRESQL_DSN=postgresql://pipebin:pipebin@localhost:5432/pipebin
FRONTEND_URL=http://localhost:8002
MAX_PASTE_SIZE_IN_BYTES=10485760
MAX_NANO_ID_LENGTH=24
LOGGER=development
```

### Frontend — same `configs/.env` (or separate env)

```
FA_PORT=8002
API_BASE_URL=http://localhost:8001
LOGGER=development
```

---

## Running locally

```bash
# Start Postgres
docker compose -f deployment/compose.local.yaml up postgres -d

# Run API
cd services/api && go run ./cmd

# Run frontend (separate terminal)
cd services/frontend && go run ./cmd
```

Or start everything with Docker Compose (requires Dockerfiles):

```bash
docker compose -f deployment/compose.local.yaml up
```

---

## Key design decisions

- **No ORM** — pgx v5 with raw SQL; all queries are in `services/api/repository/pastes_repository.go`
- **Migrations embedded** — golang-migrate reads from `services/api/migrations/embed.go` (go:embed *.sql); runs on startup, skips if no change
- **Validation** — ozzo-validation (not go-playground/validator); rules live on the request struct's `Validate()` method in `handler/pastes_handler.go`
- **IP privacy** — client IP is SHA-256 hashed before storage (`libs/hash`)
- **Frontend** — purely server-side rendered Go templates, no client-side JS required
- **Dark mode** — `color-scheme: dark` on `:root` in `static/style.css` forces oat.css dark palette
- **oat.css** — provides CSS custom properties (`--foreground`, `--border`, `--space-*`, etc.); do not replace

---

## API contract

### POST / — create paste

Request body (JSON):
```json
{ "title": "string (required)", "content": "string (required)", "language": "string (required)", "expires_at": "1h|24h|… (optional)" }
```

Response 201:
```json
{ "data": { "url": "http://localhost:8002/p/<id>" }, "status": "Created" }
```

### GET /p/{id} — get paste (JSON)

Response 200 includes: `id`, `public_id`, `title`, `content`, `language`, `created_at`, `expires_at`
Response 410 when expired.

---

## What's not done yet (see README.md TODO)

- OpenTelemetry instrumentation
- Repository integration tests (requires live Postgres / testcontainers)
- Structured config validation on startup
