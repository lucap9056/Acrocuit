[English](README.md) | 繁體中文

<div align="center">
	<h1>Acrocuit</h1>
	<p>逐層記錄建築物的電路，並追蹤任一開關或裝置的供電鏈</p>
</div>

![建築總覽，樓層平面圖堆疊顯示](docs/screenshots/overview.png)

---

## 功能

### 樓層與平面圖

- 建立建築物與各樓層，依樓層順序堆疊顯示
- 每層樓可上傳平面圖，並在圖上放置開關面板與裝置
- 平面圖上的每個項目都會顯示 icon，插座、燈具與配電盤一眼就能分辨

![平面圖上的開關面板與裝置](docs/screenshots/floor-plan.png)

### 開關面板、開關與裝置

- 開關面板內的開關可依列或欄排列，拖曳即可調整順序
- 開關可由多個開關供電，裝置也可掛在多個開關上，例如雙切燈
- 記錄額定功率、電壓、電流上限、相數、功率因數、線徑等規格

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/switch-panel-editor.png" alt="開關面板編輯"></td>
    <td width="50%"><img src="docs/screenshots/device-editor.png" alt="裝置編輯"></td>
  </tr>
</table>

### 供電鏈

- 追蹤裝置由哪些開關供電，或開關供電給哪些對象，找出關閉某個開關的影響範圍
- 檢視各開關的電流與上限、線路負載與累計電壓降，超載會標示出來
- 可在樹上關閉裝置與開關，先模擬變更的結果
- 切換到 3D 檢視，跨樓層追蹤整條供電鏈

<img src="docs/screenshots/downstream-tree.png" alt="下游樹與各開關負載">

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/upstream-tree.png" alt="裝置的上游供電鏈"></td>
    <td width="50%"><img src="docs/screenshots/chain-scene.png" alt="跨樓層供電鏈 3D 檢視"></td>
  </tr>
</table>

<img src="docs/screenshots/wire-attributes.png" alt="線路 Attributes 與電壓降" width="75%">

### 共享

- 邀請他人以編輯者或檢視者身份加入空間群組，檢視者能看到所有內容但無法修改
- 將空間群組設為公開，任何人拿到連結都能檢視，不需登入
- 可轉移擁有者，或離開不再需要的空間群組

### 圖示包

- 內建圖示包涵蓋常見電氣符號：配電、開關、插座、照明與設備
- 可透過網址加入其他圖示包

<img src="docs/screenshots/icon-picker.png" alt="Icon 選擇器" width="40%">

## 快速開始

### 單人使用

```bash
(cd web && bun install && bun run build)
go build -o bin/acrocuit ./cmd/acrocuit
./bin/acrocuit
```

開啟 `http://localhost:8080`，資料存在執行檔所在目錄的 `data/acrocuit.db`

### 多人使用（Docker）

```bash
cp docker/.env.example docker/.env   # 填入 IDENTITY_JWT_SECRET 與 OAuth2.0 / OIDC provider
docker compose -f docker/docker-compose.yml up -d --build
```

開啟 `http://localhost:8080` 並以 provider 帳號登入，細節見[部署](#部署)

---

## 關於開發

### 資料模型

`空間群組 → 空間 → 開關面板 → 開關`，裝置則連接到開關

- **排序**：空間與開關皆可排序，對應樓層與配電箱內的排列
- **開關鏈**：開關可有多個同一空間群組內的上游開關，會拒絕形成環的連結
- **裝置接線**：裝置可掛在多個開關上
- **背景圖**：每個空間可上傳一張背景圖
- **ID**：所有資源皆以 UUIDv7 識別
- **存取權限**：每個空間群組的成員具有 `owner`、`editor` 或 `viewer` 角色，且可設為公開，見[角色](docs/api.zh-TW.md#角色)
- **Attributes**：裝置、開關與兩者的連結都帶有自由格式的 JSON object `attributes`（上限 2048 bytes）
  - 後端只負責結構完整性，負載與容量規則都在前端計算

### 部署

#### SQLite

- 啟動時自動建立 schema
- `go run` 會編譯到暫存目錄，需保留資料時請設定 `DATABASE_DSN=sqlite://<path>`

#### Docker

- 包含 PostgreSQL、app、[corvauth](https://github.com/lucap9056/corvauth) 與 nginx
  - `dbinit` 與 `corvauth-schema` 在啟動前建立 table
  - 只有 nginx 對外（`http://localhost:8080`）
- 在 provider 註冊 `<PUBLIC_URL>/auth/callback` 為 redirect URL
- 經 `/auth/login` 登入後，以 `Authorization: Bearer <access_token>` 呼叫 `/api`

#### PostgreSQL

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/acrocuit?sslmode=disable"
go run ./cmd/dbinit
go run ./cmd/acrocuit
```

`dbinit` 可重複執行，但不會修改既有 table

### 設定

| 變數 | 預設值 |
| :--- | :--- |
| `DATABASE_DSN` | 執行檔所在目錄的 `data/acrocuit.db`；可設為 `postgres://...` 或 `sqlite://<path>` |
| `IDENTITY_JWT_SECRET` | 未設定（見[身份驗證](#身份驗證)） |
| `HTTP_ADDR` | `:8080` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `LOG_STD_FORMAT` / `LOG_FILE_FORMAT` | console（可選 `json`） |
| `LOG_FILE_PATH` | 停用 |

### 身份驗證

Acrocuit 不管理使用者，模式由資料庫決定：

| `DATABASE_DSN` | `IDENTITY_JWT_SECRET` | 模式 |
| :--- | :--- | :--- |
| `sqlite://` | — | 單人，固定為 `owner@localhost`，不做驗證 |
| `postgres://` | 有設定 | 多人，驗證簽章過的 `X-Forwarded-Identity` |
| `postgres://` | 未設定 | 多人，信任明文的 `X-Forwarded-User-Email` |

- 多人模式須在 reverse proxy 後方搭配 [corvauth](https://github.com/lucap9056/corvauth) 的 `/verify-optional`
  - 未帶身份的請求會以匿名身份通過，只能讀取公開的空間群組，其他操作回傳 `401`
  - 身份無效時回傳 `401`
  - 每個 email 在空間群組中能做的事依其角色而定，見[角色](docs/api.zh-TW.md#角色)
- Secret 至少 32 bytes，與 corvauth 使用同一個
- 未設定 secret 時，請確保有使用 proxy 清除 client 自帶的同名 header
- SQLite 設定 secret 只會發出警告

### 圖示包格式

圖示包由描述圖示包與分類的 `manifest.json`，以及它指向的 icon bundle 組成。Icon 選擇器接受 manifest 或其所在資料夾的網址，格式參考內建的 [`web/public/icon-packs/default`](web/public/icon-packs/default)

### 前端

前端位於 `web/`（Svelte 5、TypeScript、Vite），build 時會嵌入執行檔

```bash
docker compose -f docker-dev/docker-compose.yml up -d --build
cd web && bun run dev
```

- 開啟 `http://localhost:5173`，Vite 將 `/api`、`/auth`、`/dex` 轉給 nginx，只需這一個 port
- `docker-dev/` 為獨立的 stack，內含本地 Dex OIDC provider，不需另設 `.env`
- 以 `WEB=none`（`noweb` build tag）編譯，app 不提供前端頁面，頁面由 Vite 提供
- 帳號 `testuser_A`（`testuser_a@example.com`）、`testuser_B`（`testuser_b@example.com`），密碼皆為 `password`，登入時輸入 email
- 登入流程的網址皆指向 `http://localhost:5173`，須啟動 Vite 才能登入

### 測試

對執行中的 server 發送 request（`ACROCUIT_HOST`，預設 `http://localhost:8080`），連不上時 skip

- `tests/e2e_test.go`：含斷言的行為測試，每個 test 自行建立並清除資料
- `tests/api_test.go`：每個端點一個 request，只記錄回應

```bash
go test ./tests -run TestE2E -v
ACROCUIT_SPACE_ID=019a2b3c-6bef-7778-9ef0-728ee35bfd93 go test ./tests -v -run 'TestGetSpace$'
```

多人模式以 `ACROCUIT_ACCESS_TOKEN`（經 proxy）或 `ACROCUIT_IDENTITY_TOKEN`（直連 app）帶入身份

### API

各路由的 request body 與 response 詳見 [docs/api.zh-TW.md](docs/api.zh-TW.md)
