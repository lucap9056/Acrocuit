English | [繁體中文](README.zh-TW.md)

<div align="center">
	<h1>Acrocuit</h1>
</div>

---

Records a building's circuit topology as `space group → space → breaker group → breaker`.

- **Ordering**: spaces and breakers have a `display_order` for floors and panel layout
- **Breaker chains**: a breaker can point to an upstream breaker in the same space group
  - Query the upstream chain and downstream tree to see what switching one off affects
- **Device wiring**: a device can hang off several breakers, e.g. a light with two switches
- **Background images**: each space can have one background image

## Getting Started

### SQLite (single-user)

```bash
go build -o bin/acrocuit ./cmd/acrocuit
./bin/acrocuit
```

- The database is `data/acrocuit.db` next to the executable; its schema is created on startup
- `go run` builds into a temporary directory, so set `DATABASE_DSN=sqlite://<path>` to keep data

### Docker (PostgreSQL, multi-user)

```bash
cp docker/.env.example docker/.env   # set IDENTITY_JWT_SECRET and the OAuth2.0 / OIDC provider
docker compose -f docker/docker-compose.yml up -d --build
```

- Runs PostgreSQL, the app, [corvauth](https://github.com/lucap9056/corvauth), and nginx
  - `dbinit` and `corvauth-schema` create their tables before startup
  - Only nginx is published (`http://localhost:8080`)
- Register `<PUBLIC_URL>/auth/callback` as the provider's redirect URL
- Sign in via `/auth/login`, then call `/api` with `Authorization: Bearer <access_token>`
- Volumes created before the built-in users were removed must be reset with `down -v`

### PostgreSQL

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/acrocuit?sslmode=disable"
go run ./cmd/dbinit
go run ./cmd/acrocuit
```

`dbinit` is idempotent, but does not alter existing tables.

## Configuration

| Variable | Default |
| :--- | :--- |
| `DATABASE_DSN` | `data/acrocuit.db` next to the executable; or `postgres://...` / `sqlite://<path>` |
| `IDENTITY_JWT_SECRET` | unset (see [Authentication](#authentication)) |
| `HTTP_ADDR` | `:8080` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `LOG_STD_FORMAT` / `LOG_FILE_FORMAT` | console (`json` optional) |
| `LOG_FILE_PATH` | disabled |

## Authentication

Acrocuit does not manage users. The database decides the mode:

| `DATABASE_DSN` | `IDENTITY_JWT_SECRET` | Mode |
| :--- | :--- | :--- |
| `sqlite://` | — | Single-user, fixed owner `owner@localhost`, no authentication |
| `postgres://` | set | Multi-user, verifies the signed `X-Forwarded-Identity` |
| `postgres://` | unset | Multi-user, trusts the plain `X-Forwarded-User-Email` |

- Multi-user mode expects [corvauth](https://github.com/lucap9056/corvauth) `/verify` behind a reverse proxy
  - The email is the owner; a missing or invalid identity gets `401`
- The secret is at least 32 bytes and shared with corvauth
- Without the secret, make sure a proxy strips client-supplied identity headers
- Setting the secret with SQLite only logs a warning

## Testing

Tests send requests to a running server (`ACROCUIT_HOST`, default `http://localhost:8080`) and skip when it is unreachable.

- `tests/e2e_test.go`: behavior tests with assertions; each test creates and removes its own data
- `tests/api_test.go`: one request per endpoint, logging the response

```bash
go test ./tests -run TestE2E -v
ACROCUIT_SPACE_ID=1 go test ./tests -v -run 'TestGetSpace$'
```

In multi-user mode, pass `ACROCUIT_ACCESS_TOKEN` (through the proxy) or `ACROCUIT_IDENTITY_TOKEN` (directly to the app).

## API

- All routes are under `/api`
- Responses are wrapped as `{ "success", "data", "error" }` with `snake_case` fields

| Route | Sub-routes |
| :--- | :--- |
| `/api/space-groups` | |
| `/api/spaces` | `/{id}/background-image`: multipart field `image`, PNG/JPEG/WebP/GIF, up to 10 MB |
| `/api/breaker-groups` | |
| `/api/breakers` | `/{id}/upstream`, `/{id}/downstream` |
| `/api/devices` | `/{id}/breakers`, `/{id}/upstream` |
