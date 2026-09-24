English | [繁體中文](README.zh-TW.md)

<div align="center">
	<h1>Acrocuit</h1>

[![.NET](https://img.shields.io/badge/.NET-8.0-512BD4?logo=dotnet)](Acrocuit.csproj)
[![EF Core](https://img.shields.io/badge/EF%20Core-8.0-512BD4)](Acrocuit.csproj)
</div>

---

Acrocuit is a backend for recording how a building's electrical wiring is routed, built with ASP.NET Core 8 and Entity Framework Core. It models a building as a hierarchy of space groups, spaces, breaker groups, and breakers, and lets devices be wired to one or more breakers so switches, lights, and multi-way wiring scenarios can all be documented accurately — it does not control devices, only records the wiring topology.

## Core Features

- **Space Hierarchy**: `SpaceGroup → Space → BreakerGroup → Breaker`, each scoped to its owner.
- **Ordering**: `Space.DisplayOrder` can be used to simulate floor ordering, and `Breaker.DisplayOrder` can be used to simulate the ordering of breakers within a distribution panel.
- **Breaker Chains**: breakers can reference an upstream breaker within the same space group, forming power-distribution chains that are queried recursively (upstream chain, downstream tree), which helps identify the impact scope of a specific breaker — which devices and downstream breakers would be affected if it were switched off.
- **Device Wiring**: devices are linked to breakers through a many-to-many junction table, so a single device (e.g. a light) can be recorded as wired to multiple switches.

## Getting Started

### Prerequisites

- .NET SDK 8.0 (see `global.json`).
- A SQL Server instance (or use the provided Docker Compose setup).

### Installation

```bash
git clone https://github.com/lucap9056/Acrocuit.git
cd Acrocuit

dotnet restore
```

### Configuration

Copy the example settings file and fill in your own secrets:

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

`appsettings.json` is gitignored and must never be committed.

### Database

Schema is managed entirely through EF Core migrations:

```bash
dotnet ef database update
```

Stored procedures and triggers (under `StoredProcedures/` and `Triggers/`) are not part of the migrations — they are applied automatically at application startup via idempotent `CREATE OR ALTER` scripts.

### Running

```bash
dotnet run
```

The API listens on the URL configured by ASP.NET Core's default launch profile, with Swagger UI available in the `Development` environment.

## Docker Support

A ready-to-use SQL Server + app stack is provided under `docker.net/`.

```bash
docker compose -f docker.net/docker-compose.yml up -d --build
```

This starts:
- `mssql`: SQL Server 2022 with a health check gating the app's startup.
- `app`: the Acrocuit API, built from `docker.net/Dockerfile`, exposed on `http://localhost:5153`.

After the stack is healthy, apply migrations against it (e.g. from the host, pointing `ConnectionStrings__Default` at the exposed SQL Server port) before calling the API.

## API Overview

All endpoints except `/auth/*` require a Bearer access token.

| Resource | Base Route |
| :--- | :--- |
| Authentication | `/auth` (`register`, `login`, `refresh`) |
| Space Groups | `/space-groups` |
| Spaces | `/spaces` |
| Breaker Groups | `/breaker-groups` |
| Breakers | `/breakers` (includes `/breakers/{id}/downstream`, `/breakers/{id}/upstream`) |
| Devices | `/devices` (includes `/devices/{id}/breakers`, `/devices/{id}/upstream`) |
