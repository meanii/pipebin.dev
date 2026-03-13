# pipebin.dev
> A developer-friendly pastebin for sharing logs, code and stdout.

## Architecture

```
pipebin.dev/
├── services/
│   ├── api/                  # REST API (port 8001)
│   │   ├── cmd/              # Entrypoint
│   │   ├── handler/          # HTTP handlers
│   │   ├── repository/       # GORM DB layer
│   │   └── internal/
│   │       ├── config/       # API config loader
│   │       ├── database/     # DB connection + migrations
│   │       ├── httpx/        # HTTP utilities (client IP)
│   │       ├── server/       # Router / mux setup
│   │       └── services/     # Business logic
│   └── frontend/             # Frontend server (port 8002)
│       ├── cmd/              # Entrypoint
│       └── internal/
│           └── config/       # Frontend config loader
├── libs/
│   ├── config/               # Shared env/config helpers
│   ├── logger/               # Shared zap logger setup
│   └── models/               # Shared domain models (Paste, CreatePasteInput)
├── deployment/
│   └── compose.local.yaml    # Local Docker Compose
└── configs/                  # .env files (gitignored)
```

## TODO

### Bugs

- [ ] **`pastes_handler.go`** — `zap.S().Infof("userIp: ", userIP)` is missing the `%s` format verb; should be `zap.S().Infof("userIp: %s", userIP)`
- [ ] **`frontend/cmd/main.go`** — calls `logger.SetupLogger(...)` which does not exist; the correct function is `logger.Setup(...)`
- [ ] **`repository/pastes_repository.go`** — `GetByPublicID` does not populate `PublicID` in the returned `models.Paste` struct

---

### Infrastructure

- [ ] **Remove AutoMigrate** — replace `db.AutoMigrate(...)` in `database/db.go` with a proper migration tool (e.g. [`golang-migrate`](https://github.com/golang-migrate/migrate)); write versioned SQL migration files under `services/api/migrations/`
- [ ] **Add API + frontend services to `compose.local.yaml`** — currently only `postgres` is defined; add `api` and `frontend` service definitions with correct port mappings, env vars, and `depends_on: postgres`
- [ ] **Populate `configs/`** — add a `configs/.env.example` template documenting all required env vars (`APP_PORT`, `POSTGRESQL_DSN`, `LOGGER`); document how to copy it to `configs/.env` in the README
- [ ] **Add a health-check endpoint** — `GET /healthz` in the API returning DB connectivity status; use it in Docker Compose `healthcheck`
- [ ] **Structured config validation on startup** — fail fast with a clear error message if required env vars are missing or malformed, rather than panicking deep in the stack

---

### API Service (`services/api`)

- [ ] **Add `GET /p/{id}` route** — wire up the existing `GetPasteByPublicID` service method to a new `GetPaste` HTTP handler; register the route in `server/router.go`
- [ ] **Implement `GetPaste` handler** — return the full paste payload as JSON (`id`, `title`, `content`, `language`, `created_at`, `expires_at`); return `404` when not found, `410 Gone` when expired
- [ ] **Return structured JSON from all handlers** — replace plain-text `w.Write([]byte(...))` responses with a consistent JSON envelope (e.g. `{ "data": ..., "error": null }`)
- [ ] **Make the paste URL dynamic** — the `pipebinUrl` in `CreatePaste` is hardcoded to `http://localhost:8002`; derive the base URL from config (`FRONTEND_BASE_URL` env var)
- [ ] **Hash IP addresses before persisting** — store a SHA-256 (or bcrypt) hash of the client IP instead of the raw IP to respect user privacy; update `INET` column type to `VARCHAR` or `BYTEA` accordingly
- [ ] **Add input validation** — validate `CreatePaste` request body:
  - `content` must not be empty
  - `title` max length 255 characters
  - `content` max length configurable (e.g. 1 MB default)
  - `language` should be from an allowlist of known syntax identifiers
  - `expires_at` must be in the future if provided
- [ ] **Add OpenTelemetry trace ID injection** — instrument the API with `go.opentelemetry.io/otel`; propagate `trace_id` and `span_id` through context; export traces to an OTLP collector
- [ ] **Add request-aware logger middleware** — write a middleware that enriches the zap logger in `context.Context` with `trace_id`, `request_id`, `method`, `path`, and `user_agent`; use this contextual logger in handlers and services instead of the global `zap.S()`
- [ ] **Add rate limiting** — implement per-IP rate limiting on paste creation (e.g. using a token-bucket or leaky-bucket algorithm, or a Redis-backed store) to prevent abuse
- [ ] **Add paste expiry cleanup worker** — implement a background goroutine (or scheduled job) that periodically deletes pastes where `expires_at < now()`; also enforce expiry on read (return `410 Gone` for expired pastes)
- [ ] **Fix `GetByPublicID` GORM query** — column name casing issue: `Where("PublicID = ?", ...)` should use the snake_case column name `Where("public_id = ?", ...)` to match the `gorm:"column:public_id"` tag

---

### Frontend Service (`services/frontend`)

- [ ] **Bootstrap the HTTP server** — `frontend/cmd/main.go` loads config but never starts a server; add `net/http` listener on `FA_PORT`
- [ ] **Implement paste creation page (`GET /`)** — HTML form with fields for title, content (textarea), language (dropdown), and optional expiry; `POST` to the API service on submit
- [ ] **Implement paste view page (`GET /p/{id}`)** — fetch paste from the API, render content with syntax highlighting
- [ ] **Add syntax highlighting** — integrate a server-side or client-side highlighter (e.g. [Chroma](https://github.com/alecthomas/chroma) server-side, or [highlight.js](https://highlightjs.org/) / [Shiki](https://shiki.matsu.io/) client-side)
- [ ] **Add `FRONTEND_API_BASE_URL` config var** — frontend needs to know the API address; add to `frontend/internal/config/config.go` and document in `.env.example`
- [ ] **Add copy-to-clipboard button** on paste view
- [ ] **Add "raw" view** — `GET /p/{id}/raw` serving paste content as `text/plain` for use in shell pipelines (the core pipebin use case)
- [ ] **Add `curl`-friendly API response** — detect `curl` user-agent on `POST /` and return the raw paste URL as plain text (no JSON wrapper) so `cmd | curl -F 'c=@-' https://pipebin.dev` works out of the box

---

### Testing

- [ ] **Service layer unit tests** — test `CreatePaste` and `GetPasteByPublicID` with a mock repository; cover happy path, invalid publicID length, expired paste, and DB errors
- [ ] **Repository integration tests** — test `Create` and `GetByPublicID` against a real Postgres instance (spin up via `testcontainers-go` or use `compose.local.yaml`)
- [ ] **Handler unit tests** — use `net/http/httptest` to test `CreatePaste` and `GetPaste` handlers; mock the service layer; cover bad request bodies, missing fields, and service errors
- [ ] **IP hashing utility tests** — once IP hashing is added, unit-test the hash function with known inputs
- [ ] **End-to-end test** — script a full round-trip: create a paste via `POST /`, fetch it via `GET /p/{id}`, assert content matches
- [ ] **Add CI pipeline** — GitHub Actions workflow running `go test ./...` and `go vet ./...` on every PR; add Bazel build check
