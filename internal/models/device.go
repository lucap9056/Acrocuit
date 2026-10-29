package models

import (
	"acrocuit/internal/models/columns"
	"acrocuit/schema"
	"fmt"
	"strings"
)

var deviceWithPositionFields = []uint8{0, 1, 2, 3, 4, 5, 6}

type Device struct {
	Id           DeviceId     `json:"id"`
	Name         string       `json:"name"`
	Attributes   Attributes   `json:"attributes" tstype:"Attributes | null"`
	SpaceId      SpaceId      `json:"-"`
	SpaceGroupId SpaceGroupId `json:"-"`
	PositionX    *int         `json:"-"`
	PositionY    *int         `json:"-"`
	PositionZ    *int         `json:"-"`

	Space    *Ref[SpaceId] `json:"space"`
	Position *Position     `json:"position"`
}

func (d *Device) Build() {
	if d.Space == nil && d.SpaceId != "" {
		d.Space = &Ref[SpaceId]{Id: d.SpaceId}
	}
	if d.Position == nil && d.PositionX != nil {
		d.Position = &Position{X: *d.PositionX, Y: *d.PositionY, Z: *d.PositionZ}
	}
}

func ParseDeviceColumns(columnsStr string, allowPosition bool) (string, []uint8, bool, error) {
	if strings.TrimSpace(columnsStr) == "" {
		if allowPosition {
			return DEVICE_WITH_POSITION_COLUMNS, deviceWithPositionFields, true, nil
		}
		return DEVICE_ALL_COLUMNS, []uint8{0, 1, 2, 3}, false, nil
	}

	collector := columns.NewCollector(columnsStr)
	defer collector.Release()

	scanner := columns.NewScanner(columnsStr)

	includePosition := false

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.Devices_ID, schema.FullDevices_ID:
			collector.Append(0, schema.FullDevices_ID)
		case schema.Devices_NAME, schema.FullDevices_NAME:
			collector.Append(1, schema.FullDevices_NAME)
		case schema.Devices_SPACE_ID, schema.FullDevices_SPACE_ID:
			collector.Append(2, schema.FullDevices_SPACE_ID)
		case schema.Devices_ATTRIBUTES, schema.FullDevices_ATTRIBUTES:
			collector.Append(3, schema.FullDevices_ATTRIBUTES)
		case BASE_POSITION:
			if !allowPosition {
				return "", nil, false, fmt.Errorf("%w: %s", ErrInvalidColumn, col)
			}
			collector.Append(4, schema.FullDevicePositions_POSITION_X)
			collector.Append(5, schema.FullDevicePositions_POSITION_Y)
			collector.Append(6, schema.FullDevicePositions_POSITION_Z)
			includePosition = true
		default:
			return "", nil, false, fmt.Errorf("%w: %s", ErrInvalidColumn, col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, includePosition, nil
}

func (d *Device) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &d.Id
		case 1:
			values[i] = &d.Name
		case 2:
			values[i] = &d.SpaceId
		case 3:
			values[i] = &d.Attributes
		case 4:
			values[i] = &d.PositionX
		case 5:
			values[i] = &d.PositionY
		case 6:
			values[i] = &d.PositionZ
		}
	}
	return values
}

func (d *Device) ValuesWithPosition() []any {
	return d.Values(deviceWithPositionFields)
}
