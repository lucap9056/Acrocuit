[English](README.md) | 繁體中文

<div align="center">
	<h1>Acrocuit</h1>
</div>

---

Acrocuit 將建築物建模為 `空間群組 → 空間 → 斷路器群組 → 斷路器` 的階層結構，並讓裝置能連接到一個或多個斷路器，藉此準確記錄開關、電器的電路走向與拓樸。

## 核心功能

- **空間階層**：`SpaceGroup → Space → BreakerGroup → Breaker`，每一層都歸屬特定擁有者。
- **排序**：可透過 `Space` 的 `DisplayOrder` 來模擬樓層排序，可透過 `Breaker` 的 `DisplayOrder` 來模擬配電箱中的斷路器排序。
- **斷路器鏈**：斷路器可以在同一個 space group 內指向上游斷路器，形成配電鏈，並透過遞迴查詢取得（上游鏈、下游樹），有助於看出特定斷路器的影響範圍——若將其關閉，會影響哪些裝置與下游斷路器。
- **裝置接線**：裝置與斷路器透過多對多的關聯表連接，讓單一裝置（例如一盞燈）可以記錄成同時由多個開關控制。

## 開始使用

### Docker

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

啟動 PostgreSQL，透過 `dbinit` 套用 schema，並在 `http://localhost:8080` 提供 API。以 `docker compose -f docker/docker-compose.yml down -v` 重置資料庫。

### 本機

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/acrocuit?sslmode=disable"
export JWT_SECRET_KEY="REPLACE_WITH_YOUR_OWN_SECRET"

go run ./cmd/dbinit
go run ./cmd/acrocuit
```

`dbinit` 可重複執行，但不會修改既有 table——table 結構變更需重建資料庫才會生效。

### 設定

| 變數 | 預設值 |
| :--- | :--- |
| `DATABASE_DSN` | 必填 |
| `JWT_SECRET_KEY` | 必填 |
| `JWT_ISSUER` / `JWT_AUDIENCE` | `Acrocuit` / `AcrocuitClient` |
| `JWT_ACCESS_TOKEN_EXPIRY_MINUTES` / `JWT_REFRESH_TOKEN_EXPIRY_DAYS` | `15` / `7` |
| `HTTP_ADDR` | `:8080` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `LOG_STD_FORMAT` / `LOG_FILE_FORMAT` | console（可選 `json`） |
| `LOG_FILE_PATH` | 停用 |

## 測試

`tests/api_test.go` 為每個端點各提供一個 test，對執行中的 server 送出單一 request 並記錄回應。

```bash
ACROCUIT_ACCESS_TOKEN=... ACROCUIT_SPACE_ID=1 go test ./tests -v -run 'TestGetSpace$'
```

## API 概覽

除了 `/auth/*` 之外，所有端點都需要 Bearer access token。回應統一包成 `{ "success", "data", "error" }`，欄位採 `snake_case`。

| 資源 | 基礎路由 |
| :--- | :--- |
| 身份驗證 | `/auth`（`register`、`login`、`refresh`） |
| 空間群組 | `/space-groups` |
| 空間 | `/spaces` |
| 斷路器群組 | `/breaker-groups` |
| 斷路器 | `/breakers`（包含 `/breakers/{id}/upstream`、`/breakers/{id}/downstream`） |
| 裝置 | `/devices`（包含 `/devices/{id}/breakers`、`/devices/{id}/upstream`） |
