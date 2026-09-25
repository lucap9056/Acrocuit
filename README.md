English | [繁體中文](README.zh-TW.md)

<div align="center">
	<h1>Acrocuit</h1>
</div>

---

Acrocuit models a building as a hierarchy of space groups, spaces, breaker groups, and breakers, and lets devices be wired to one or more breakers, so the circuit routing and topology of switches and appliances can be recorded accurately.

## Core Features

- **Space Hierarchy**: `SpaceGroup → Space → BreakerGroup → Breaker`, each scoped to its owner.
- **Ordering**: `Space.DisplayOrder` can be used to simulate floor ordering, and `Breaker.DisplayOrder` can be used to simulate the ordering of breakers within a distribution panel.
- **Breaker Chains**: breakers can reference an upstream breaker within the same space group, forming power-distribution chains that are queried recursively (upstream chain, downstream tree), which helps identify the impact scope of a specific breaker — which devices and downstream breakers would be affected if it were switched off.
- **Device Wiring**: devices are linked to breakers through a many-to-many junction table, so a single device (e.g. a light) can be recorded as wired to multiple switches.

## Getting Started

### Docker

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

Starts PostgreSQL, applies the schema via `dbinit`, then serves the API on `http://localhost:8080`. Reset the database with `docker compose -f docker/docker-compose.yml down -v`.

### Local

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/acrocuit?sslmode=disable"
export JWT_SECRET_KEY="REPLACE_WITH_YOUR_OWN_SECRET"

go run ./cmd/dbinit
go run ./cmd/acrocuit
```

`dbinit` is idempotent, but it does not alter existing tables — recreate the database to apply table structure changes.

### Configuration

| Variable | Default |
| :--- | :--- |
| `DATABASE_DSN` | required |
| `JWT_SECRET_KEY` | required |
| `JWT_ISSUER` / `JWT_AUDIENCE` | `Acrocuit` / `AcrocuitClient` |
| `JWT_ACCESS_TOKEN_EXPIRY_MINUTES` / `JWT_REFRESH_TOKEN_EXPIRY_DAYS` | `15` / `7` |
| `HTTP_ADDR` | `:8080` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `LOG_STD_FORMAT` / `LOG_FILE_FORMAT` | console (`json` optional) |
| `LOG_FILE_PATH` | disabled |

## Testing

`tests/api_test.go` has one test per endpoint; each sends a single request to a running server and logs the response.

```bash
ACROCUIT_ACCESS_TOKEN=... ACROCUIT_SPACE_ID=1 go test ./tests -v -run 'TestGetSpace$'
```

## API Overview

All endpoints except `/auth/*` require a Bearer access token. Responses are wrapped as `{ "success", "data", "error" }` with `snake_case` fields.

| Resource | Base Route |
| :--- | :--- |
| Authentication | `/auth` (`register`, `login`, `refresh`) |
| Space Groups | `/space-groups` |
| Spaces | `/spaces` |
| Breaker Groups | `/breaker-groups` |
| Breakers | `/breakers` (includes `/breakers/{id}/upstream`, `/breakers/{id}/downstream`) |
| Devices | `/devices` (includes `/devices/{id}/breakers`, `/devices/{id}/upstream`) |
