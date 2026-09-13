[English](README.md) | 繁體中文

<div align="center">
	<h1>Acrocuit</h1>

[![.NET](https://img.shields.io/badge/.NET-8.0-512BD4?logo=dotnet)](Acrocuit.csproj)
[![EF Core](https://img.shields.io/badge/EF%20Core-8.0-512BD4)](Acrocuit.csproj)
</div>

---

Acrocuit 是一個使用 ASP.NET Core 8 與 Entity Framework Core 打造的空間電路走向紀錄工具。它將建築物建模為 `空間群組 → 空間 → 斷路器群組 → 斷路器` 的階層結構，並讓裝置能連接到一個或多個斷路器，藉此準確記錄開關、燈具與多切開關等接線情境——它不控制裝置，只記錄電路的走向與拓樸。

## 核心功能

- **空間階層**：`SpaceGroup → Space → BreakerGroup → Breaker`，每一層都歸屬特定擁有者。
- **排序**：可透過 `Space` 的 `DisplayOrder` 來模擬樓層排序，可透過 `Breaker` 的 `DisplayOrder` 來模擬配電箱中的斷路器排序。
- **斷路器鏈**：斷路器可以在同一個 space group 內指向上游斷路器，形成配電鏈，並透過遞迴查詢取得（上游鏈、下游樹），有助於看出特定斷路器的影響範圍——若將其關閉，會影響哪些裝置與下游斷路器。
- **裝置接線**：裝置與斷路器透過多對多的關聯表連接，讓單一裝置（例如一盞燈）可以記錄成同時由多個開關控制。

## 開始使用

### 先決條件

- .NET SDK 8.0（詳見 `global.json`）。
- 一個 SQL Server 執行個體（或直接使用提供的 Docker Compose 環境）。

### 安裝

```bash
git clone https://github.com/lucap9056/Acrocuit.git
cd Acrocuit

dotnet restore
```

### 設定

複製範例設定檔並填入自己的機敏資訊：

```bash
cp appsettings.json.example appsettings.json
```

```json
{
  "JwtSettings": {
    "SecretKey": "REPLACE_WITH_YOUR_OWN_SECRET"
  },
  "ConnectionStrings": {
    "Default": "Server=localhost;Database=Acrocuit;User Id=sa;Password=REPLACE_ME;TrustServerCertificate=True"
  }
}
```

`appsettings.json` 已加入 `.gitignore`，不得提交到版本控制。

### 資料庫

Schema 完全由 EF Core migrations 管理：

```bash
dotnet ef database update
```

Stored procedure 與 trigger（位於 `StoredProcedures/` 與 `Triggers/`）不屬於 migrations 的一部分，而是在應用程式啟動時，透過具備冪等性的 `CREATE OR ALTER` 腳本自動套用。

### 執行

```bash
dotnet run
```

API 會依 ASP.NET Core 預設的啟動設定監聽，並在 `Development` 環境下提供 Swagger UI。

## Docker 支援

`docker/` 目錄下提供了可直接使用的 SQL Server + app 服務棧。

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

此指令會啟動：
- `mssql`：SQL Server 2022，並透過 health check 確保其就緒後才啟動 app。
- `app`：Acrocuit API，由 `docker/Dockerfile` 建置，對外暴露於 `http://localhost:5153`。

服務棧健康後，請先對其套用 migrations（例如從主機端，將 `ConnectionStrings__Default` 指向對外暴露的 SQL Server port），再呼叫 API。

## API 概覽

除了 `/auth/*` 之外，所有端點都需要 Bearer access token。

| 資源 | 基礎路由 |
| :--- | :--- |
| 身份驗證 | `/auth`（`register`、`login`、`refresh`） |
| 空間群組 | `/space-groups` |
| 空間 | `/spaces` |
| 斷路器群組 | `/breaker-groups` |
| 斷路器 | `/breakers`（包含 `/breakers/{id}/downstream`、`/breakers/{id}/upstream`） |
| 裝置 | `/devices`（包含 `/devices/{id}/breakers`、`/devices/{id}/upstream`） |

