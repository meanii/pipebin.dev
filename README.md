# pipebin.dev

<div align="center">
  <img src="assets/logo.svg" alt="pipebin.dev" width="520" />
  <p>Minimal pastebin for developers. Pipe logs, code, and stdout directly from your terminal.</p>
</div>

```bash
cat crash.log | curl -sT - https://pipebin.dev/
→ https://pipebin.dev/p/xk9d2mA1B2fg
```

---

## Features

- **No login required** — paste and share immediately
- **Pipe-friendly** — raw body, JSON, or HTML form
- **Language auto-detection** — Chroma analyses content when no language is specified
- **Syntax highlighting with line numbers** — server-side via Chroma v2 (`github-dark` theme)
- **Burn after reading** — paste is deleted immediately after the first view (`?once=1`)
- **Expiring pastes** — set a TTL with `?e=1h`, `?e=24h`, `?e=7d`
- **Raw plain text** — `GET /raw/{id}` for piping into other tools
- **curl-friendly** — plain text URL response when called from curl
- **Shell alias** — one-liner to add to your shell config
- **10 MB paste limit**
- **No JavaScript required**
- **IP privacy** — client IP is SHA-256 hashed before storage

---

## Usage

### Simple pipe

```bash
cat file.go | curl -sT - https://pipebin.dev/
→ https://pipebin.dev/p/xk9d2m
```

> **Why `-sT -`?**
> `curl -T` sends a raw PUT request with no encoding — whitespace, tabs, and indentation are preserved exactly.
> `curl -d` URL-encodes the body which destroys indentation.

### With metadata

Query params let you attach metadata in raw pipe mode:

| Param | Alias | Description |
| --- | --- | --- |
| `?t=` | — | Paste title |
| `?lang=` | — | Language (e.g. `go`, `python`, `bash`, `json`) |
| `?e=` | `?expires=` | Expiry duration (e.g. `1h`, `24h`, `168h`) |
| `?once=1` | `?burn=1` | Burn after reading — deleted on first view |

```bash
# with title, language, and expiry
dmesg | curl -sT - "https://pipebin.dev/?t=dmesg&lang=bash&e=6h"
→ https://pipebin.dev/p/ab3f7x

# burn after reading — deleted the moment someone opens it
echo "my-secret-token" | curl -sT - "https://pipebin.dev/?once=1"
→ https://pipebin.dev/p/zz9k1q
   ⚠  burn after reading — deleted on first view
```

### Fetch raw output

```bash
curl -s https://pipebin.dev/raw/xk9d2m
```

### JSON API

```bash
curl -s https://pipebin.dev/ \
  -H 'Content-Type: application/json' \
  -d '{"title":"hello","content":"world","language":"go","expires_at":"24h"}'

→ {"data":{"url":"https://pipebin.dev/p/cd4e8w","burn":false},"status":"Created"}
```

---

## Shell alias

Add this to your `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish` for a one-command pipe:

### bash / zsh

```bash
pb() { curl -sT - "https://pipebin.dev/$@"; }
```

### fish

```fish
function pb
    curl -sT - "https://pipebin.dev/$argv"
end
```

### Usage after adding the alias

```bash
cat crash.log | pb
→ https://pipebin.dev/p/xk9d2m

cat secret.txt | pb '?once=1'
→ https://pipebin.dev/p/zz9k1q
   ⚠  burn after reading — deleted on first view

cat deploy.sh | pb '?t=deploy&lang=bash&e=24h'
→ https://pipebin.dev/p/ab3f7x
```

---

## API Reference

### `POST /` or `PUT /` — Create paste

#### Raw pipe (recommended for curl)

```bash
curl -sT - "https://pipebin.dev/?t=title&lang=go&e=24h&once=1"
```

#### JSON body

```json
{
  "title": "string (required)",
  "content": "string (required)",
  "language": "string (required, e.g. go / python / text)",
  "expires_at": "string (optional, Go duration e.g. 1h, 24h, 168h)",
  "burn": false
}
```

#### Response `201 Created`

```json
{
  "data": { "url": "https://pipebin.dev/p/<id>", "burn": false },
  "status": "Created"
}
```

#### curl response — plain text

```text
https://pipebin.dev/p/<id>
```

### `GET /p/{id}` — Get paste (JSON)

#### Response `200 OK`

```json
{
  "data": {
    "id": "...",
    "public_id": "...",
    "title": "...",
    "content": "...",
    "language": "go",
    "created_at": "2025-01-01T00:00:00Z",
    "expires_at": null,
    "burn_after_reading": false
  }
}
```

#### Error responses

- `404 Not Found` — paste does not exist
- `410 Gone` — paste has expired

### `GET /raw/{id}` — Raw plain text

Returns the paste content as `text/plain`. Use for piping into other tools:

```bash
curl -s https://pipebin.dev/raw/<id> | grep ERROR
```

---

## Language auto-detection

When `?lang=` is not provided in raw pipe mode, pipebin analyses the content using
[Chroma's lexer analyser](https://github.com/alecthomas/chroma). If a language is confidently
detected, it is applied automatically. Falls back to `text` when confidence is low or the
content is plain prose.

Auto-detection is only applied in raw pipe mode — JSON and form submissions always use the
language you specify.

---

## Burn after reading

Append `?once=1` (or `?burn=1`) to the URL when creating a paste. The paste is permanently
deleted from the database the moment anyone opens the view URL — there is no second chance.

```bash
echo "POSTGRES_PASSWORD=hunter2" | curl -sT - "https://pipebin.dev/?once=1"
→ https://pipebin.dev/p/zz9k1q
   ⚠  burn after reading — deleted on first view
```

The view page displays a prominent warning confirming the paste has been burned.

In JSON mode, set `"burn": true` in the request body.

---

## Architecture

```text
pipebin.dev/
├── services/
│   ├── api/                  # REST API (port 8001)
│   │   ├── cmd/              # Entrypoint — wires DB → repo → service → handler
│   │   ├── handler/          # HTTP handlers (pastes_handler.go)
│   │   ├── repository/       # pgx raw SQL (pastes_repository.go + interface)
│   │   ├── migrations/       # Embedded SQL migration files (golang-migrate)
│   │   └── internal/
│   │       ├── config/       # API config (APP_PORT, POSTGRESQL_DSN, FRONTEND_URL, …)
│   │       ├── database/     # pgxpool connection + migrate runner
│   │       ├── httpx/        # JSON response helpers + client IP extraction
│   │       ├── middleware/   # RequestID + logger middleware
│   │       ├── server/       # Router / mux setup
│   │       └── services/     # Business logic (CreatePaste, GetPasteByPublicID)
│   └── frontend/             # Frontend server (port 8002)
│       ├── cmd/              # Entrypoint — embeds templates + static, starts server
│       ├── handlers/         # Page handlers (Home, CreatePaste, Paste, RawPaste)
│       └── internal/
│           ├── config/       # Frontend config (FA_PORT, API_BASE_URL, LOGGER)
│           └── server/       # Router / mux setup
├── libs/
│   ├── config/               # Shared env helpers (GetEnv, MustGetEnv, LoadDotEnv)
│   ├── hash/                 # SHA-256 IP hashing
│   ├── logger/               # slog setup
│   └── models/               # Shared domain structs (Paste, CreatePasteInput)
├── deployment/
│   └── compose.local.yaml    # Local Docker Compose (postgres + api + frontend)
└── configs/                  # .env files (gitignored); see .env.example
```

---

## Running locally

**Prerequisites:** Go 1.24+, Docker

```bash
# Start Postgres
docker compose -f deployment/compose.local.yaml up postgres -d

# Run API (port 8001)
cd services/api && go run ./cmd

# Run frontend in a separate terminal (port 8002)
cd services/frontend && go run ./cmd
```

Or start everything with Docker Compose:

```bash
docker compose -f deployment/compose.local.yaml up
```

### Environment variables

Copy `configs/.env.example` to `configs/.env` and adjust as needed.

#### API (`configs/.env`)

```bash
APP_PORT=8001
POSTGRESQL_DSN=postgresql://pipebin:pipebin@localhost:5432/pipebin
FRONTEND_URL=http://localhost:8002
MAX_PASTE_SIZE_IN_BYTES=10485760
MAX_NANO_ID_LENGTH=24
LOGGER=development
```

#### Frontend (same file or separate env)

```bash
FA_PORT=8002
API_BASE_URL=http://localhost:8001
LOGGER=development
```

---

## Tech stack

| Layer | Technology |
| --- | --- |
| Language | Go 1.24+ |
| Database driver | `github.com/jackc/pgx/v5` — pgxpool, raw SQL (no ORM) |
| Migrations | `github.com/golang-migrate/migrate/v4` — embedded SQL |
| Validation | `github.com/go-ozzo/ozzo-validation/v4` |
| Syntax highlighting | `github.com/alecthomas/chroma/v2` |
| Logging | `log/slog` (stdlib) |
| ID generation | `github.com/matoous/go-nanoid/v2` |
| Frontend styles | [oat.css](https://oatcss.dev) CDN + custom CSS |
| Build | Bazel + plain `go build ./...` |

---

## TODO / Not yet implemented

- [ ] OpenTelemetry instrumentation
- [ ] Repository integration tests (testcontainers)
- [ ] Structured config validation on startup
- [ ] Copy-to-clipboard button (requires JS; contradicts no-JS goal)
- [ ] Rate limiting (currently only on API paste creation)
