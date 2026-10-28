English | [繁體中文](README.zh-TW.md)

<div align="center">
	<h1>Acrocuit</h1>
	<p>Map a building's circuits floor by floor, then trace any switch or device through its power chain</p>
</div>

![Building overview with stacked floor plans](docs/screenshots/overview.png)

---

## Features

### Floors and floor plans

- Add a building and its floors, shown stacked in order
- Upload a floor plan for each floor and place switch panels and devices on it
- Each item shows its icon on the plan, so outlets, lights and panels read like a drawing

![Floor plan with switch panels and devices](docs/screenshots/floor-plan.png)

### Switch panels, switches and devices

- Lay out a switch panel's switches in rows or columns, and drag to reorder
- A switch can be fed by several switches, and a device can hang off several switches, e.g. a two-way light
- Record ratings such as power, voltage, current limit, phase, power factor and wire size

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/switch-panel-editor.png" alt="Switch panel editor"></td>
    <td width="50%"><img src="docs/screenshots/device-editor.png" alt="Device editor"></td>
  </tr>
</table>

### Power chain

- Trace what feeds a device, or what a switch feeds, to see what switching one off affects
- See the current on each switch against its limit, wire load and cumulative voltage drop, with overloads highlighted
- Turn devices and switches off in the tree to try a change before making it
- Switch to the 3D view to follow the chain across floors

<img src="docs/screenshots/downstream-tree.png" alt="Downstream tree with load per switch">

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/upstream-tree.png" alt="Upstream chain of a device"></td>
    <td width="50%"><img src="docs/screenshots/chain-scene.png" alt="Power chain across floors in 3D"></td>
  </tr>
</table>

<img src="docs/screenshots/wire-attributes.png" alt="Wire attributes and voltage drop" width="75%">

### Sharing

- Invite people to a space group as editors or viewers; viewers see everything but cannot change it
- Make a space group public to share its link with anyone, no sign-in needed
- Transfer ownership, or leave a space group you no longer need

### Icon packs

- The built-in pack covers common electrical symbols: distribution, switches, outlets, lighting and equipment
- Add more icon packs from a URL

<img src="docs/screenshots/icon-picker.png" alt="Icon picker" width="40%">

## Getting Started

### Single user

```bash
(cd web && bun install && bun run build)
go build -o bin/acrocuit ./cmd/acrocuit
./bin/acrocuit
```

Open `http://localhost:8080`. Data is kept in `data/acrocuit.db` next to the executable.

### Multiple users (Docker)

```bash
cp docker/.env.example docker/.env   # set IDENTITY_JWT_SECRET and the OAuth2.0 / OIDC provider
docker compose -f docker/docker-compose.yml up -d --build
```

Open `http://localhost:8080` and sign in with your provider. See [Deployment](#deployment) for the details.

---

## Development

### Data Model

`space group → space → switch panel → switch`, with devices attached to switches.

- **Ordering**: spaces and switches keep their order, matching the floors and the panel layout
- **Switch chains**: a switch can have several upstream switches in the same space group, cycles are rejected
- **Device wiring**: a device can hang off several switches
- **Background images**: each space can have one background image
- **IDs**: every resource is identified by a UUIDv7
- **Access**: each space group has members with the role `owner`, `editor` or `viewer`, and can be made public; see [Roles](docs/api.md#roles)
- **Attributes**: devices, switches and their links carry a free-form JSON object `attributes` (up to 2048 bytes)
  - The backend only keeps the structure valid; load and capacity rules are evaluated in the frontend

### Deployment

#### SQLite

- The schema is created on startup
- `go run` builds into a temporary directory, so set `DATABASE_DSN=sqlite://<path>` to keep data

#### Docker

- Runs PostgreSQL, the app, [corvauth](https://github.com/lucap9056/corvauth), and nginx
  - `dbinit` and `corvauth-schema` create their tables before startup
  - Only nginx is published (`http://localhost:8080`)
- Register `<PUBLIC_URL>/auth/callback` as the provider's redirect URL
- Sign in via `/auth/login`, then call `/api` with `Authorization: Bearer <access_token>`

#### PostgreSQL

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/acrocuit?sslmode=disable"
go run ./cmd/dbinit
go run ./cmd/acrocuit
```

`dbinit` is idempotent, but does not alter existing tables.

### Configuration

| Variable | Default |
| :--- | :--- |
| `DATABASE_DSN` | `data/acrocuit.db` next to the executable; or `postgres://...` / `sqlite://<path>` |
| `IDENTITY_JWT_SECRET` | unset (see [Authentication](#authentication)) |
| `HTTP_ADDR` | `:8080` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `LOG_STD_FORMAT` / `LOG_FILE_FORMAT` | console (`json` optional) |
| `LOG_FILE_PATH` | disabled |

### Authentication

Acrocuit does not manage users. The database decides the mode:

| `DATABASE_DSN` | `IDENTITY_JWT_SECRET` | Mode |
| :--- | :--- | :--- |
| `sqlite://` | — | Single-user, fixed owner `owner@localhost`, no authentication |
| `postgres://` | set | Multi-user, verifies the signed `X-Forwarded-Identity` |
| `postgres://` | unset | Multi-user, trusts the plain `X-Forwarded-User-Email` |

- Multi-user mode expects [corvauth](https://github.com/lucap9056/corvauth) `/verify-optional` behind a reverse proxy
  - A request without an identity passes through as anonymous and can only read public space groups; anything else gets `401`
  - An invalid identity gets `401`
  - What each email can do in a space group follows its role; see [Roles](docs/api.md#roles)
- The secret is at least 32 bytes and shared with corvauth
- Without the secret, make sure a proxy strips client-supplied identity headers
- Setting the secret with SQLite only logs a warning

### Icon Packs

An icon pack is a `manifest.json` that describes the pack and its categories, plus an icon bundle it points to. The icon picker accepts the URL of the manifest or of its folder. See the built-in [`web/public/icon-packs/default`](web/public/icon-packs/default).

### Frontend

The frontend lives in `web/` (Svelte 5, TypeScript, Vite) and is embedded into the binary on build.

```bash
docker compose -f docker-dev/docker-compose.yml up -d --build
cd web && bun run dev
```

- Open `http://localhost:5173`; Vite forwards `/api`, `/auth`, and `/dex` to nginx, so this is the only port you need
- `docker-dev/` is a standalone stack with a local Dex OIDC provider; no `.env` needed
- Built with `WEB=none` (the `noweb` build tag), so the app serves no frontend; pages come from Vite
- Accounts `testuser_A` (`testuser_a@example.com`) and `testuser_B` (`testuser_b@example.com`), both with password `password`; sign in with the email
- Sign-in URLs point to `http://localhost:5173`, so Vite must be running to sign in

### Testing

Tests send requests to a running server (`ACROCUIT_HOST`, default `http://localhost:8080`) and skip when it is unreachable.

- `tests/e2e_test.go`: behavior tests with assertions; each test creates and removes its own data
- `tests/api_test.go`: one request per endpoint, logging the response

```bash
go test ./tests -run TestE2E -v
ACROCUIT_SPACE_ID=019a2b3c-6bef-7778-9ef0-728ee35bfd93 go test ./tests -v -run 'TestGetSpace$'
```

In multi-user mode, pass `ACROCUIT_ACCESS_TOKEN` (through the proxy) or `ACROCUIT_IDENTITY_TOKEN` (directly to the app).

### API

See [docs/api.md](docs/api.md) for every route with its request body and response.
