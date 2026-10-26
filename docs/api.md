English | [繁體中文](api.zh-TW.md)

# API

## Conventions

- All routes are under `/api`; request and response bodies are JSON with `snake_case` fields
- Every route except `/api/meta` needs an identity; see [Authentication](../README.md#authentication)
  - Through the proxy, send `Authorization: Bearer <access_token>`
- IDs in the path are integers

### Responses

Success:

```json
{
  "success": true,
  "data": {}
}
```

Failure:

```json
{
  "success": false,
  "data": null,
  "error": "not found"
}
```

| Status | When |
| :--- | :--- |
| `400` | Invalid path ID, invalid body, an unknown `include` column, a switch fed by itself, or a link that would form a cycle |
| `401` | Missing or invalid identity |
| `404` | The resource does not exist or belongs to someone else |
| `409` | The link already exists |
| `413` | Background image over 10 MB |
| `415` | Background image is not PNG, JPEG, WebP or GIF |

### Partial fields

`GET` routes that return space groups, spaces, switch panels, switches or devices accept `include`, a comma-separated list of columns. Columns left out come back as zero values.

```
GET /api/spaces/20?include=id,name
```

```json
{
  "success": true,
  "data": {
    "id": 20,
    "name": "5F",
    "display_order": 0,
    "background_image_updated_at": 0,
    "space_group": null
  }
}
```

| Resource | Columns |
| :--- | :--- |
| Space group | `id`, `name` |
| Space | `id`, `name`, `space_group_id`, `display_order`, `background_image_updated_at` |
| Switch panel | `id`, `name`, `space_id`, `space_group_id`, `attributes`, `position` |
| Switch | `id`, `name`, `switch_panel_id`, `space_group_id`, `display_order`, `attributes` |
| Device | `id`, `name`, `space_id`, `attributes`, `position` |

### Attributes

Switch panels, switches, devices and links carry `attributes`, a JSON object of up to 2048 bytes. The backend stores it as is; its keys are up to the client.

- On create, an omitted `attributes` becomes `{}`
- On update, an omitted `attributes` keeps the current value; a given one replaces it entirely

### Nested references

Nested `space_group`, `space` and `switch_panel` objects are references holding only `id`. Power chains are the exception: each switch's `switch_panel` is complete.

## Meta

### `GET /api/meta`

No identity needed.

```json
{
  "success": true,
  "data": {
    "auth_required": true
  }
}
```

`auth_required` tells whether other routes need an identity.

## Space Groups

### `GET /api/space-groups`

```json
{
  "success": true,
  "data": [
    {
      "id": 3,
      "name": "Acrocuit Tower"
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

| Field | Rule |
| :--- | :--- |
| `name` | Required, 1–255 characters |

```json
{
  "success": true,
  "data": {
    "id": 3,
    "name": "Acrocuit Tower"
  }
}
```

### `GET /api/space-groups/{spaceGroupId}`

```json
{
  "success": true,
  "data": {
    "id": 3,
    "name": "Acrocuit Tower"
  }
}
```

### `PUT /api/space-groups/{spaceGroupId}`

Renames the space group.

```json
{
  "name": "Acrocuit Tower"
}
```

```json
{
  "success": true,
  "data": {
    "id": 3,
    "name": "Acrocuit Tower"
  }
}
```

### `DELETE /api/space-groups/{spaceGroupId}`

Deletes the space group with all its spaces, switch panels, switches, devices and links.

```json
{
  "success": true,
  "data": null
}
```

## Spaces

### `GET /api/space-groups/{spaceGroupId}/spaces`

Ordered by `display_order`.

```json
{
  "success": true,
  "data": [
    {
      "id": 13,
      "name": "B3",
      "display_order": 1,
      "background_image_updated_at": 1793011394640,
      "space_group": {
        "id": 3
      }
    }
  ]
}
```

`background_image_updated_at` is a Unix time in milliseconds, or `0` without a background image.

### `POST /api/space-groups/{spaceGroupId}/spaces`

Adds a space at the end of the order.

```json
{
  "name": "5F"
}
```

| Field | Rule |
| :--- | :--- |
| `name` | Required, 1–255 characters |

```json
{
  "success": true,
  "data": {
    "id": 20,
    "name": "5F",
    "display_order": 8,
    "background_image_updated_at": 0,
    "space_group": {
      "id": 3
    }
  }
}
```

### `GET /api/spaces/{spaceId}`

```json
{
  "success": true,
  "data": {
    "id": 20,
    "name": "5F",
    "display_order": 8,
    "background_image_updated_at": 1793011395389,
    "space_group": {
      "id": 3
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

| Field | Rule |
| :--- | :--- |
| `name` | Optional, 1–255 characters |

Returns the space, as in `GET /api/spaces/{spaceId}`.

### `DELETE /api/spaces/{spaceId}`

Deletes the space with its background image, switch panels, switches, devices and links.

```json
{
  "success": true,
  "data": null
}
```

### `PATCH /api/spaces/{spaceId}/order`

Moves the space to `display_order`; the spaces in between shift by one.

```json
{
  "display_order": 3
}
```

Returns every space in the space group, ordered by `display_order`.

```json
{
  "success": true,
  "data": [
    {
      "id": 13,
      "name": "B3",
      "display_order": 1,
      "background_image_updated_at": 1793011394640,
      "space_group": {
        "id": 3
      }
    }
  ]
}
```

## Background Images

### `GET /api/spaces/{spaceId}/background-image`

Returns the image itself with its `Content-Type`.

- `ETag` follows `background_image_updated_at`; a matching `If-None-Match` gets `304`
- `404` when the space has no background image

### `PUT /api/spaces/{spaceId}/background-image`

A `multipart/form-data` request with the image in the `image` field.

```
curl -X PUT -F image=@5f.png http://localhost:8080/api/spaces/20/background-image
```

- PNG, JPEG, WebP or GIF, decided by the file content
- Up to 10 MB

Returns the space with the new `background_image_updated_at`.

```json
{
  "success": true,
  "data": {
    "id": 20,
    "name": "5F",
    "display_order": 8,
    "background_image_updated_at": 1793011395389,
    "space_group": {
      "id": 3
    }
  }
}
```

### `DELETE /api/spaces/{spaceId}/background-image`

Returns the space with `background_image_updated_at` set to `0`.

## Switch Panels

### `GET /api/spaces/{spaceId}/switch-panels`

```json
{
  "success": true,
  "data": [
    {
      "id": 30,
      "name": "DB-5F",
      "attributes": {
        "icon": "default:distribution_board",
        "switch_capacity": 16
      },
      "space": {
        "id": 20
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

`position` is `null` until the switch panel is placed on the floor plan.

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

| Field | Rule |
| :--- | :--- |
| `name` | Required, 1–255 characters |
| `position` | Optional, integer `x`, `y`, `z` |
| `attributes` | Optional |

Returns the switch panel, as in `GET /api/switch-panels/{switchPanelId}`.

### `GET /api/switch-panels/{switchPanelId}`

```json
{
  "success": true,
  "data": {
    "id": 30,
    "name": "DB-5F",
    "attributes": {
      "icon": "default:distribution_board",
      "switch_capacity": 16
    },
    "space": {
      "id": 20
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

| Field | Rule |
| :--- | :--- |
| `name` | Optional, 1–255 characters |
| `attributes` | Optional, replaces the current value |

Returns the switch panel.

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

Returns the switch panel.

### `DELETE /api/switch-panels/{switchPanelId}`

Deletes the switch panel with its switches and their links.

```json
{
  "success": true,
  "data": null
}
```

## Switches

### `GET /api/switch-panels/{switchPanelId}/switches`

Ordered by `display_order`.

```json
{
  "success": true,
  "data": [
    {
      "id": 182,
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
        "id": 30
      }
    }
  ]
}
```

### `POST /api/switch-panels/{switchPanelId}/switches`

Adds a switch at the end of the order.

```json
{
  "name": "UPS-5 Input",
  "attributes": {
    "max_current": 100
  }
}
```

| Field | Rule |
| :--- | :--- |
| `name` | Required, 1–255 characters |
| `attributes` | Optional |

Returns the switch, as in `GET /api/switches/{switchId}`.

### `GET /api/switches/{switchId}`

```json
{
  "success": true,
  "data": {
    "id": 182,
    "name": "UPS-5 Input",
    "display_order": 7,
    "attributes": {
      "max_current": 100
    },
    "switch_panel": {
      "id": 30
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

| Field | Rule |
| :--- | :--- |
| `name` | Optional, 1–255 characters |
| `attributes` | Optional, replaces the current value |

Returns the switch.

### `DELETE /api/switches/{switchId}`

Deletes the switch with its links to other switches and devices.

```json
{
  "success": true,
  "data": null
}
```

### `PATCH /api/switches/{switchId}/order`

Moves the switch to `display_order` within its switch panel; the switches in between shift by one.

```json
{
  "display_order": 2
}
```

Returns every switch in the switch panel, ordered by `display_order`, in the same shape as `GET /api/switch-panels/{switchPanelId}/switches`.

## Switch Links

A switch link connects a switch to one of its upstream switches. Both switches must be in the same space group.

### `GET /api/switches/{switchId}/upstreams`

The switches that feed this switch, each with the link's attributes.

```json
{
  "success": true,
  "data": [
    {
      "id": 176,
      "name": "DB-5F Main",
      "display_order": 1,
      "attributes": {
        "max_current": 100
      },
      "switch_panel": {
        "id": 30
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
  "upstream_switch_id": 176,
  "attributes": {
    "wire_size": 150,
    "wire_length": 1
  }
}
```

| Field | Rule |
| :--- | :--- |
| `upstream_switch_id` | Required |
| `attributes` | Optional |

- `400` when the switch is its own upstream, or the link would form a cycle
- `404` when the upstream switch is not in the same space group
- `409` when the link already exists

```json
{
  "success": true,
  "data": {
    "upstream_switch_id": 176,
    "downstream_switch_id": 182,
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

| Field | Rule |
| :--- | :--- |
| `attributes` | Required, replaces the current value |

Returns the link, as in `POST /api/switches/{switchId}/upstreams`.

### `DELETE /api/switches/{switchId}/upstreams/{upstreamSwitchId}`

```json
{
  "success": true,
  "data": null
}
```

## Devices

### `GET /api/spaces/{spaceId}/devices`

```json
{
  "success": true,
  "data": [
    {
      "id": 267,
      "name": "Server Rack 1",
      "attributes": {
        "icon": "default:outlet_grounded",
        "rated_power": 4000,
        "rated_voltage": 220,
        "power_factor": 0.95
      },
      "space": {
        "id": 20
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

`position` is `null` until the device is placed on the floor plan.

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

| Field | Rule |
| :--- | :--- |
| `name` | Required, 1–255 characters |
| `position` | Optional, integer `x`, `y`, `z` |
| `attributes` | Optional |

Returns the device, as in `GET /api/devices/{deviceId}`.

### `GET /api/devices/{deviceId}`

```json
{
  "success": true,
  "data": {
    "id": 267,
    "name": "Server Rack 1",
    "attributes": {
      "rated_power": 4000,
      "rated_voltage": 220
    },
    "space": {
      "id": 20
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

| Field | Rule |
| :--- | :--- |
| `name` | Optional, 1–255 characters |
| `position` | Optional |
| `attributes` | Optional, replaces the current value |

Returns the device.

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

Returns the device.

### `DELETE /api/devices/{deviceId}`

Deletes the device with its links to switches.

```json
{
  "success": true,
  "data": null
}
```

## Device Links

A device link connects a device to a switch that feeds it. The switch must be in the device's space group.

### `GET /api/devices/{deviceId}/switches`

The switches that feed this device, each with the link's attributes.

```json
{
  "success": true,
  "data": [
    {
      "id": 185,
      "name": "Rack 1",
      "display_order": 2,
      "attributes": {
        "max_current": 32
      },
      "switch_panel": {
        "id": 31
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
  "switch_id": 185,
  "attributes": {
    "wire_size": 5.5,
    "wire_length": 17.3
  }
}
```

| Field | Rule |
| :--- | :--- |
| `switch_id` | Required |
| `attributes` | Optional |

- `404` when the switch is not in the device's space group
- `409` when the link already exists

```json
{
  "success": true,
  "data": {
    "device_id": 267,
    "switch_id": 185,
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

| Field | Rule |
| :--- | :--- |
| `attributes` | Required, replaces the current value |

Returns the link, as in `POST /api/devices/{deviceId}/switches`.

### `DELETE /api/devices/{deviceId}/switches/{switchId}`

```json
{
  "success": true,
  "data": null
}
```

## Power Chains

Both routes return a power chain:

| Field | Content |
| :--- | :--- |
| `switches` | Switches in the chain, each with its complete `switch_panel` |
| `switch_links` | Links between those switches |
| `devices` | Devices in the chain |
| `device_links` | Links between those devices and switches |

### `GET /api/switches/{switchId}/downstream`

The switch, every switch it feeds directly or indirectly, and the devices hanging off them.

```json
{
  "success": true,
  "data": {
    "switches": [
      {
        "id": 182,
        "name": "UPS-5 Input",
        "display_order": 7,
        "attributes": {
          "max_current": 100
        },
        "switch_panel": {
          "id": 30,
          "name": "DB-5F",
          "attributes": {
            "switch_capacity": 16
          },
          "space": {
            "id": 20
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
        "upstream_switch_id": 182,
        "downstream_switch_id": 184,
        "attributes": {
          "wire_size": 22,
          "wire_length": 14.7
        }
      }
    ],
    "devices": [
      {
        "id": 267,
        "name": "Server Rack 1",
        "attributes": {
          "rated_power": 4000
        },
        "space": {
          "id": 20
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
        "device_id": 267,
        "switch_id": 185,
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

The device, the switches it hangs off, and every switch feeding those directly or indirectly. `devices` holds only this device.

```json
{
  "success": true,
  "data": {
    "switches": [
      {
        "id": 109,
        "name": "MSB Main Incomer",
        "display_order": 1,
        "attributes": {
          "max_current": 2500
        },
        "switch_panel": {
          "id": 21,
          "name": "MSB",
          "attributes": {
            "switch_capacity": 16
          },
          "space": {
            "id": 15
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
        "upstream_switch_id": 109,
        "downstream_switch_id": 117,
        "attributes": {
          "wire_size": 150,
          "wire_length": 1
        }
      }
    ],
    "devices": [
      {
        "id": 267,
        "name": "Server Rack 1",
        "attributes": {
          "rated_power": 4000
        },
        "space": {
          "id": 20
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
        "device_id": 267,
        "switch_id": 185,
        "attributes": {
          "wire_size": 5.5,
          "wire_length": 17.3
        }
      }
    ]
  }
}
```
