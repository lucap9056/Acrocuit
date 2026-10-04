[English](README.md) | 繁體中文

<div align="center">
	<h1>Acrocuit</h1>
</div>

---

以 `空間群組 → 空間 → 斷路器群組 → 斷路器` 記錄建築物的電路拓樸

- **排序**：空間與斷路器皆有 `display_order`，可模擬樓層與配電箱內的排列
- **斷路器鏈**：斷路器可指向同一空間群組內的上游斷路器
  - 查詢上游鏈與下游樹，找出關閉某個斷路器的影響範圍
- **裝置接線**：裝置可掛在多個斷路器上，例如由兩個開關控制的燈
- **背景圖**：每個空間可上傳一張背景圖

## 快速開始

### SQLite（單人）

```bash
go build -o bin/acrocuit ./cmd/acrocuit
./bin/acrocuit
```

- 資料庫位於執行檔所在目錄的 `data/acrocuit.db`，啟動時自動建立 schema
- `go run` 會編譯到暫存目錄，需保留資料時請設定 `DATABASE_DSN=sqlite://<path>`

### Docker（PostgreSQL，多人）

```bash
cp docker/.env.example docker/.env   # 填入 IDENTITY_JWT_SECRET 與 OAuth2.0 / OIDC provider
docker compose -f docker/docker-compose.yml up -d --build
```

- 包含 PostgreSQL、app、[corvauth](https://github.com/lucap9056/corvauth) 與 nginx
  - `dbinit` 與 `corvauth-schema` 在啟動前建立 table
  - 只有 nginx 對外（`http://localhost:8080`）
- 在 provider 註冊 `<PUBLIC_URL>/auth/callback` 為 redirect URL
- 經 `/auth/login` 登入後，以 `Authorization: Bearer <access_token>` 呼叫 `/api`
- 移除內建使用者之前建立的 volume 須以 `down -v` 重置

### PostgreSQL

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/acrocuit?sslmode=disable"
go run ./cmd/dbinit
go run ./cmd/acrocuit
```

`dbinit` 可重複執行，但不會修改既有 table

## 設定

| 變數 | 預設值 |
| :--- | :--- |
| `DATABASE_DSN` | 執行檔所在目錄的 `data/acrocuit.db`；可設為 `postgres://...` 或 `sqlite://<path>` |
| `IDENTITY_JWT_SECRET` | 未設定（見[身份驗證](#身份驗證)） |
| `HTTP_ADDR` | `:8080` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `LOG_STD_FORMAT` / `LOG_FILE_FORMAT` | console（可選 `json`） |
| `LOG_FILE_PATH` | 停用 |

## 身份驗證

Acrocuit 不管理使用者，模式由資料庫決定：

| `DATABASE_DSN` | `IDENTITY_JWT_SECRET` | 模式 |
| :--- | :--- | :--- |
| `sqlite://` | — | 單人，固定為 `owner@localhost`，不做驗證 |
| `postgres://` | 有設定 | 多人，驗證簽章過的 `X-Forwarded-Identity` |
| `postgres://` | 未設定 | 多人，信任明文的 `X-Forwarded-User-Email` |

- 多人模式須在 reverse proxy 後方搭配 [corvauth](https://github.com/lucap9056/corvauth) 的 `/verify`
  - Email 即為擁有者，缺少或無效時回傳 `401`
- Secret 至少 32 bytes，與 corvauth 使用同一個
- 未設定 secret 時，請確保有使用 proxy 清除 client 自帶的同名 header
- SQLite 設定 secret 只會發出警告

## 測試

對執行中的 server 發送 request（`ACROCUIT_HOST`，預設 `http://localhost:8080`），連不上時 skip

- `tests/e2e_test.go`：含斷言的行為測試，每個 test 自行建立並清除資料
- `tests/api_test.go`：每個端點一個 request，只記錄回應

```bash
go test ./tests -run TestE2E -v
ACROCUIT_SPACE_ID=1 go test ./tests -v -run 'TestGetSpace$'
```

多人模式以 `ACROCUIT_ACCESS_TOKEN`（經 proxy）或 `ACROCUIT_IDENTITY_TOKEN`（直連 app）帶入身份

## API

- 路由皆在 `/api` 底下
- 回應包成 `{ "success", "data", "error" }`，欄位採 `snake_case`

| 路由 | 子路由 |
| :--- | :--- |
| `/api/space-groups` | |
| `/api/spaces` | `/{id}/background-image`：multipart 欄位 `image`，PNG/JPEG/WebP/GIF，上限 10 MB |
| `/api/breaker-groups` | |
| `/api/breakers` | `/{id}/upstream`、`/{id}/downstream` |
| `/api/devices` | `/{id}/breakers`、`/{id}/upstream` |
