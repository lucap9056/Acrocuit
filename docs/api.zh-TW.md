[English](api.md) | 繁體中文

# API

## 慣例

- 路由皆在 `/api` 底下，request 與 response body 皆為 JSON，欄位採 `snake_case`
- 除了 `/api/meta` 與公開空間群組的 `GET` 路由之外，路由都須帶身份，見[身份驗證](../README.zh-TW.md#身份驗證)
  - 經 proxy 時帶 `Authorization: Bearer <access_token>`
  - 列出與建立空間群組、成員管理以及所有寫入操作一律須帶身份
- ID 皆為 UUIDv7 字串

### 回應格式

成功：

```json
{
  "success": true,
  "data": {}
}
```

失敗：

```json
{
  "success": false,
  "data": null,
  "error": "not found"
}
```

| 狀態碼 | 情況 |
| :--- | :--- |
| `400` | 路徑 ID 不是 UUID、body 不合法、`include` 含未知欄位、開關以自己為上游、連結會形成環，或成員的角色或 email 無效 |
| `401` | 缺少或無效的身份，或匿名呼叫須帶身份的路由 |
| `404` | 資源不存在，或呼叫者的角色不允許該操作 |
| `409` | 連結或成員已存在，或擁有者離開時沒有編輯者可接手 |
| `413` | 背景圖超過 10 MB |
| `415` | 背景圖不是 PNG、JPEG、WebP 或 GIF |
| `501` | 在 SQLite 後端使用成員管理 |

### 部分欄位

回傳空間群組、空間、開關面板、開關或裝置的 `GET` 路由接受 `include`，以逗號分隔要取得的欄位，未列出的欄位以零值回傳

```
GET /api/spaces/019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3?include=id,name
```

```json
{
  "success": true,
  "data": {
    "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3",
    "name": "5F",
    "display_order": 0,
    "background_image_updated_at": 0,
    "space_group": null
  }
}
```

| 資源 | 欄位 |
| :--- | :--- |
| 空間群組 | `id`、`name`、`role` |
| 空間 | `id`、`name`、`space_group_id`、`display_order`、`background_image_updated_at` |
| 開關面板 | `id`、`name`、`space_id`、`space_group_id`、`attributes`、`position` |
| 開關 | `id`、`name`、`switch_panel_id`、`space_group_id`、`display_order`、`attributes` |
| 裝置 | `id`、`name`、`space_id`、`attributes`、`position` |

### Attributes

開關面板、開關、裝置與連結都帶有 `attributes`，為上限 2048 bytes 的 JSON object。後端原樣儲存，key 由 client 自行決定

- 建立時省略 `attributes` 會存成 `{}`
- 更新時省略 `attributes` 會保留原值，有帶則整個取代

### 角色

每個空間群組會給每位成員一個角色：

| 角色 | 可以 |
| :--- | :--- |
| `viewer` | 讀取空間群組與其下所有內容 |
| `editor` | 另外可重新命名空間群組，以及建立、修改與刪除其下的空間、開關面板、開關、裝置、連結與背景圖 |
| `owner` | 另外可刪除空間群組、管理成員、轉移擁有者與設定公開 |

- 每個空間群組恰有一位擁有者，轉移之前為建立者
- 公開的空間群組任何人都能讀取，包括匿名請求；非成員以 `viewer` 身份讀取
- 空間群組帶有 `role`，即呼叫者的角色
- 單人模式下呼叫者固定為 `owner`

### 巢狀參照

巢狀的 `space_group`、`space` 與 `switch_panel` 為只帶 `id` 的參照。供電鏈例外，其中每個開關的 `switch_panel` 是完整的

## Meta

### `GET /api/meta`

不需身份

```json
{
  "success": true,
  "data": {
    "auth_required": true
  }
}
```

`auth_required` 表示伺服器是否為多人模式，此模式下路由須帶身份，並可使用成員管理

## 空間群組

### `GET /api/space-groups`

列出呼叫者所屬的空間群組，須帶身份

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f",
      "name": "Acrocuit Tower",
      "role": "owner"
    }
  ]
}
```

### `POST /api/space-groups`

```json
{
  "name": "Acrocuit Tower"
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 必填，1–255 字元 |

```json
{
  "success": true,
  "data": {
    "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f",
    "name": "Acrocuit Tower",
    "role": "owner"
  }
}
```

### `GET /api/space-groups/{spaceGroupId}`

成員可讀取；空間群組公開時任何人皆可讀取

```json
{
  "success": true,
  "data": {
    "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f",
    "name": "Acrocuit Tower",
    "role": "owner"
  }
}
```

### `PUT /api/space-groups/{spaceGroupId}`

重新命名空間群組，須為 `editor` 或 `owner`

```json
{
  "name": "Acrocuit Tower"
}
```

```json
{
  "success": true,
  "data": {
    "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f",
    "name": "Acrocuit Tower",
    "role": "owner"
  }
}
```

### `DELETE /api/space-groups/{spaceGroupId}`

刪除空間群組，連同其下所有空間、開關面板、開關、裝置與連結，須為 `owner`

```json
{
  "success": true,
  "data": null
}
```

## 空間群組成員

這些路由須帶身份，在 SQLite 後端回傳 `501`。除了離開之外，只有擁有者能呼叫

### `GET /api/space-groups/{spaceGroupId}/members`

成員依擁有者、編輯者、檢視者排序

```json
{
  "success": true,
  "data": {
    "public": false,
    "members": [
      {
        "email": "owner@example.com",
        "role": "owner"
      },
      {
        "email": "editor@example.com",
        "role": "editor"
      }
    ]
  }
}
```

### `POST /api/space-groups/{spaceGroupId}/members`

```json
{
  "email": "viewer@example.com",
  "role": "viewer"
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `email` | 必填，email 格式 |
| `role` | 必填，`viewer` 或 `editor` |

```json
{
  "success": true,
  "data": null
}
```

### `PUT /api/space-groups/{spaceGroupId}/members/{email}`

變更成員的角色。擁有者的角色無法在此變更，請改用轉移擁有者

```json
{
  "role": "editor"
}
```

```json
{
  "success": true,
  "data": null
}
```

### `DELETE /api/space-groups/{spaceGroupId}/members/{email}`

移除成員。帶 `me` 或呼叫者自己的 email 時改為離開空間群組，任何成員都能離開

- 擁有者離開時會隨機由一位編輯者成為擁有者；沒有編輯者時回傳 `409`

```json
{
  "success": true,
  "data": null
}
```

### `PUT /api/space-groups/{spaceGroupId}/owner`

將擁有者轉移給另一位成員，原擁有者變為編輯者

```json
{
  "email": "editor@example.com"
}
```

```json
{
  "success": true,
  "data": null
}
```

### `PUT /api/space-groups/{spaceGroupId}/public`

設定是否讓任何人都能讀取空間群組

```json
{
  "public": true
}
```

```json
{
  "success": true,
  "data": null
}
```

## 空間

### `GET /api/space-groups/{spaceGroupId}/spaces`

依 `display_order` 排序

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b3d-df23-7f11-8432-ac56cf00e328",
      "name": "B3",
      "display_order": 1,
      "background_image_updated_at": 1793011394640,
      "space_group": {
        "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f"
      }
    }
  ]
}
```

`background_image_updated_at` 為毫秒單位的 Unix time，沒有背景圖時為 `0`

### `POST /api/space-groups/{spaceGroupId}/spaces`

新增的空間排在最後

```json
{
  "name": "5F"
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 必填，1–255 字元 |

```json
{
  "success": true,
  "data": {
    "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3",
    "name": "5F",
    "display_order": 8,
    "background_image_updated_at": 0,
    "space_group": {
      "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f"
    }
  }
}
```

### `GET /api/spaces/{spaceId}`

```json
{
  "success": true,
  "data": {
    "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3",
    "name": "5F",
    "display_order": 8,
    "background_image_updated_at": 1793011395389,
    "space_group": {
      "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f"
    }
  }
}
```

### `PUT /api/spaces/{spaceId}`

```json
{
  "name": "5F"
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 選填，1–255 字元 |

回傳空間，格式同 `GET /api/spaces/{spaceId}`

### `DELETE /api/spaces/{spaceId}`

刪除空間，連同其背景圖、開關面板、開關、裝置與連結

```json
{
  "success": true,
  "data": null
}
```

### `PATCH /api/spaces/{spaceId}/order`

將空間移到 `display_order`，中間的空間依序遞移

```json
{
  "display_order": 3
}
```

回傳空間群組內所有空間，依 `display_order` 排序

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b3d-df23-7f11-8432-ac56cf00e328",
      "name": "B3",
      "display_order": 1,
      "background_image_updated_at": 1793011394640,
      "space_group": {
        "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f"
      }
    }
  ]
}
```

## 背景圖

### `GET /api/spaces/{spaceId}/background-image`

直接回傳圖片本身與對應的 `Content-Type`

- `ETag` 隨 `background_image_updated_at` 變動，`If-None-Match` 相符時回傳 `304`
- 空間沒有背景圖時回傳 `404`

### `PUT /api/spaces/{spaceId}/background-image`

以 `multipart/form-data` 上傳，圖片放在 `image` 欄位

```
curl -X PUT -F image=@5f.png http://localhost:8080/api/spaces/019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3/background-image
```

- PNG、JPEG、WebP 或 GIF，依檔案內容判斷
- 上限 10 MB

回傳更新 `background_image_updated_at` 後的空間

```json
{
  "success": true,
  "data": {
    "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3",
    "name": "5F",
    "display_order": 8,
    "background_image_updated_at": 1793011395389,
    "space_group": {
      "id": "019a2b3c-a9cd-77aa-aa3f-c0799f042a8f"
    }
  }
}
```

### `DELETE /api/spaces/{spaceId}/background-image`

回傳 `background_image_updated_at` 為 `0` 的空間

## 開關面板

### `GET /api/spaces/{spaceId}/switch-panels`

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b3f-ed02-77fa-80cb-3ac593bac529",
      "name": "DB-5F",
      "attributes": {
        "icon": "default:distribution_board",
        "switch_capacity": 16
      },
      "space": {
        "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
      },
      "position": {
        "x": 82,
        "y": 169,
        "z": 0
      }
    }
  ]
}
```

尚未放到平面圖上時 `position` 為 `null`

### `POST /api/spaces/{spaceId}/switch-panels`

```json
{
  "name": "DB-5F",
  "position": {
    "x": 82,
    "y": 169,
    "z": 0
  },
  "attributes": {
    "icon": "default:distribution_board",
    "switch_capacity": 16
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 必填，1–255 字元 |
| `position` | 選填，整數 `x`、`y`、`z` |
| `attributes` | 選填 |

回傳開關面板，格式同 `GET /api/switch-panels/{switchPanelId}`

### `GET /api/switch-panels/{switchPanelId}`

```json
{
  "success": true,
  "data": {
    "id": "019a2b3f-ed02-77fa-80cb-3ac593bac529",
    "name": "DB-5F",
    "attributes": {
      "icon": "default:distribution_board",
      "switch_capacity": 16
    },
    "space": {
      "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
    },
    "position": {
      "x": 82,
      "y": 169,
      "z": 0
    }
  }
}
```

### `PUT /api/switch-panels/{switchPanelId}`

```json
{
  "name": "DB-5F",
  "attributes": {
    "icon": "default:distribution_board",
    "switch_capacity": 16
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 選填，1–255 字元 |
| `attributes` | 選填，整個取代原值 |

回傳開關面板

### `PUT /api/switch-panels/{switchPanelId}/position`

```json
{
  "position": {
    "x": 82,
    "y": 169,
    "z": 0
  }
}
```

回傳開關面板

### `DELETE /api/switch-panels/{switchPanelId}`

刪除開關面板，連同其下開關與這些開關的連結

```json
{
  "success": true,
  "data": null
}
```

## 開關

### `GET /api/switch-panels/{switchPanelId}/switches`

依 `display_order` 排序

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b52-4aea-7072-b438-e2b9a519e640",
      "name": "UPS-5 Input",
      "display_order": 7,
      "attributes": {
        "circuit_label": "UPS",
        "switch_type": "nfb",
        "max_current": 100,
        "rated_voltage": 380,
        "phase": "three",
        "poles": 3
      },
      "switch_panel": {
        "id": "019a2b3f-ed02-77fa-80cb-3ac593bac529"
      }
    }
  ]
}
```

### `POST /api/switch-panels/{switchPanelId}/switches`

新增的開關排在最後

```json
{
  "name": "UPS-5 Input",
  "attributes": {
    "max_current": 100
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 必填，1–255 字元 |
| `attributes` | 選填 |

回傳開關，格式同 `GET /api/switches/{switchId}`

### `GET /api/switches/{switchId}`

```json
{
  "success": true,
  "data": {
    "id": "019a2b52-4aea-7072-b438-e2b9a519e640",
    "name": "UPS-5 Input",
    "display_order": 7,
    "attributes": {
      "max_current": 100
    },
    "switch_panel": {
      "id": "019a2b3f-ed02-77fa-80cb-3ac593bac529"
    }
  }
}
```

### `PUT /api/switches/{switchId}`

```json
{
  "name": "UPS-5 Input",
  "attributes": {
    "max_current": 100
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 選填，1–255 字元 |
| `attributes` | 選填，整個取代原值 |

回傳開關

### `DELETE /api/switches/{switchId}`

刪除開關，連同其與其他開關、裝置的連結

```json
{
  "success": true,
  "data": null
}
```

### `PATCH /api/switches/{switchId}/order`

將開關移到所屬開關面板內的 `display_order`，中間的開關依序遞移

```json
{
  "display_order": 2
}
```

回傳開關面板內所有開關，依 `display_order` 排序，格式同 `GET /api/switch-panels/{switchPanelId}/switches`

## 開關連結

開關連結將開關連到它的一個上游開關，兩者須在同一個空間群組

### `GET /api/switches/{switchId}/upstreams`

供電給此開關的開關，各自帶有連結的 attributes

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b51-9150-7560-86e8-723c97ed63c0",
      "name": "DB-5F Main",
      "display_order": 1,
      "attributes": {
        "max_current": 100
      },
      "switch_panel": {
        "id": "019a2b3f-ed02-77fa-80cb-3ac593bac529"
      },
      "link": {
        "attributes": {
          "wire_size": 150,
          "wire_length": 1
        }
      }
    }
  ]
}
```

### `POST /api/switches/{switchId}/upstreams`

```json
{
  "upstream_switch_id": "019a2b51-9150-7560-86e8-723c97ed63c0",
  "attributes": {
    "wire_size": 150,
    "wire_length": 1
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `upstream_switch_id` | 必填 |
| `attributes` | 選填 |

- 以自己為上游，或連結會形成環時回傳 `400`
- 上游開關不在同一個空間群組時回傳 `404`
- 連結已存在時回傳 `409`

```json
{
  "success": true,
  "data": {
    "upstream_switch_id": "019a2b51-9150-7560-86e8-723c97ed63c0",
    "downstream_switch_id": "019a2b52-4aea-7072-b438-e2b9a519e640",
    "attributes": {
      "wire_size": 150,
      "wire_length": 1
    }
  }
}
```

### `PUT /api/switches/{switchId}/upstreams/{upstreamSwitchId}`

```json
{
  "attributes": {
    "wire_size": 150,
    "wire_length": 1
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `attributes` | 必填，整個取代原值 |

回傳連結，格式同 `POST /api/switches/{switchId}/upstreams`

### `DELETE /api/switches/{switchId}/upstreams/{upstreamSwitchId}`

```json
{
  "success": true,
  "data": null
}
```

## 裝置

### `GET /api/spaces/{spaceId}/devices`

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
      "name": "Server Rack 1",
      "attributes": {
        "icon": "default:outlet_grounded",
        "rated_power": 4000,
        "rated_voltage": 220,
        "power_factor": 0.95
      },
      "space": {
        "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
      },
      "position": {
        "x": -583,
        "y": 185,
        "z": 0
      }
    }
  ]
}
```

尚未放到平面圖上時 `position` 為 `null`

### `POST /api/spaces/{spaceId}/devices`

```json
{
  "name": "Server Rack 1",
  "position": {
    "x": -583,
    "y": 185,
    "z": 0
  },
  "attributes": {
    "rated_power": 4000,
    "rated_voltage": 220
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 必填，1–255 字元 |
| `position` | 選填，整數 `x`、`y`、`z` |
| `attributes` | 選填 |

回傳裝置，格式同 `GET /api/devices/{deviceId}`

### `GET /api/devices/{deviceId}`

```json
{
  "success": true,
  "data": {
    "id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
    "name": "Server Rack 1",
    "attributes": {
      "rated_power": 4000,
      "rated_voltage": 220
    },
    "space": {
      "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
    },
    "position": {
      "x": -583,
      "y": 185,
      "z": 0
    }
  }
}
```

### `PUT /api/devices/{deviceId}`

```json
{
  "name": "Server Rack 1",
  "position": {
    "x": -583,
    "y": 185,
    "z": 0
  },
  "attributes": {
    "rated_power": 4000,
    "rated_voltage": 220
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `name` | 選填，1–255 字元 |
| `position` | 選填 |
| `attributes` | 選填，整個取代原值 |

回傳裝置

### `PUT /api/devices/{deviceId}/position`

```json
{
  "position": {
    "x": -583,
    "y": 185,
    "z": 0
  }
}
```

回傳裝置

### `DELETE /api/devices/{deviceId}`

刪除裝置，連同其與開關的連結

```json
{
  "success": true,
  "data": null
}
```

## 裝置連結

裝置連結將裝置連到供電給它的開關，開關須在裝置所屬的空間群組

### `GET /api/devices/{deviceId}/switches`

供電給此裝置的開關，各自帶有連結的 attributes

```json
{
  "success": true,
  "data": [
    {
      "id": "019a2b52-a7b7-714e-a9ac-c68af050eed3",
      "name": "Rack 1",
      "display_order": 2,
      "attributes": {
        "max_current": 32
      },
      "switch_panel": {
        "id": "019a2b40-0bf1-7b12-89b0-a3f22d43ab2c"
      },
      "link": {
        "attributes": {
          "wire_size": 5.5,
          "wire_length": 17.3
        }
      }
    }
  ]
}
```

### `POST /api/devices/{deviceId}/switches`

```json
{
  "switch_id": "019a2b52-a7b7-714e-a9ac-c68af050eed3",
  "attributes": {
    "wire_size": 5.5,
    "wire_length": 17.3
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `switch_id` | 必填 |
| `attributes` | 選填 |

- 開關不在裝置所屬的空間群組時回傳 `404`
- 連結已存在時回傳 `409`

```json
{
  "success": true,
  "data": {
    "device_id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
    "switch_id": "019a2b52-a7b7-714e-a9ac-c68af050eed3",
    "attributes": {
      "wire_size": 5.5,
      "wire_length": 17.3
    }
  }
}
```

### `PUT /api/devices/{deviceId}/switches/{switchId}`

```json
{
  "attributes": {
    "wire_size": 5.5,
    "wire_length": 17.3
  }
}
```

| 欄位 | 規則 |
| :--- | :--- |
| `attributes` | 必填，整個取代原值 |

回傳連結，格式同 `POST /api/devices/{deviceId}/switches`

### `DELETE /api/devices/{deviceId}/switches/{switchId}`

```json
{
  "success": true,
  "data": null
}
```

## 供電鏈

兩個路由都回傳供電鏈：

| 欄位 | 內容 |
| :--- | :--- |
| `switches` | 鏈上的開關，各自帶有完整的 `switch_panel` |
| `switch_links` | 這些開關之間的連結 |
| `devices` | 鏈上的裝置 |
| `device_links` | 這些裝置與開關之間的連結 |

### `GET /api/switches/{switchId}/downstream`

此開關、它直接或間接供電的所有開關，以及掛在這些開關上的裝置

```json
{
  "success": true,
  "data": {
    "switches": [
      {
        "id": "019a2b52-4aea-7072-b438-e2b9a519e640",
        "name": "UPS-5 Input",
        "display_order": 7,
        "attributes": {
          "max_current": 100
        },
        "switch_panel": {
          "id": "019a2b3f-ed02-77fa-80cb-3ac593bac529",
          "name": "DB-5F",
          "attributes": {
            "switch_capacity": 16
          },
          "space": {
            "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
          },
          "position": {
            "x": 82,
            "y": 169,
            "z": 0
          }
        }
      }
    ],
    "switch_links": [
      {
        "upstream_switch_id": "019a2b52-4aea-7072-b438-e2b9a519e640",
        "downstream_switch_id": "019a2b52-88c8-7d07-9858-3ae64028a233",
        "attributes": {
          "wire_size": 22,
          "wire_length": 14.7
        }
      }
    ],
    "devices": [
      {
        "id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
        "name": "Server Rack 1",
        "attributes": {
          "rated_power": 4000
        },
        "space": {
          "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
        },
        "position": {
          "x": -583,
          "y": 185,
          "z": 0
        }
      }
    ],
    "device_links": [
      {
        "device_id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
        "switch_id": "019a2b52-a7b7-714e-a9ac-c68af050eed3",
        "attributes": {
          "wire_size": 5.5,
          "wire_length": 17.3
        }
      }
    ]
  }
}
```

### `GET /api/devices/{deviceId}/upstream`

此裝置、它掛載的開關，以及直接或間接供電給這些開關的所有開關。`devices` 只包含此裝置

```json
{
  "success": true,
  "data": {
    "switches": [
      {
        "id": "019a2b49-78c3-7335-b5d3-b98ac80bcf74",
        "name": "MSB Main Incomer",
        "display_order": 1,
        "attributes": {
          "max_current": 2500
        },
        "switch_panel": {
          "id": "019a2b3e-d69b-7630-b97c-400000f18c97",
          "name": "MSB",
          "attributes": {
            "switch_capacity": 16
          },
          "space": {
            "id": "019a2b3e-1d01-73b6-9f41-fbecf7302d57"
          },
          "position": {
            "x": -190,
            "y": -131,
            "z": 0
          }
        }
      }
    ],
    "switch_links": [
      {
        "upstream_switch_id": "019a2b49-78c3-7335-b5d3-b98ac80bcf74",
        "downstream_switch_id": "019a2b4a-703b-7d58-93ed-d5525e90d703",
        "attributes": {
          "wire_size": 150,
          "wire_length": 1
        }
      }
    ],
    "devices": [
      {
        "id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
        "name": "Server Rack 1",
        "attributes": {
          "rated_power": 4000
        },
        "space": {
          "id": "019a2b3e-b7ac-761e-be1b-1c7a58fd6ec3"
        },
        "position": {
          "x": -583,
          "y": 185,
          "z": 0
        }
      }
    ],
    "device_links": [
      {
        "device_id": "019a2b5c-9045-7fe9-beb7-afc0437e6b8d",
        "switch_id": "019a2b52-a7b7-714e-a9ac-c68af050eed3",
        "attributes": {
          "wire_size": 5.5,
          "wire_length": 17.3
        }
      }
    ]
  }
}
```
