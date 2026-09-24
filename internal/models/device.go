package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const DEVICE_ALL_COLUMNS = schema.FullDevices_ID + "," + schema.FullDevices_NAME + "," + schema.FullDevices_SPACE_ID

type Device struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	SpaceId      int    `json:"-"`
	SpaceGroupId int    `json:"-"`
	BreakerId    int    `json:"-"`
	PositionX    *int   `json:"-"`
	PositionY    *int   `json:"-"`
	PositionZ    *int   `json:"-"`

	Space    *Space    `json:"space"`
	Breaker  *Breaker  `json:"breaker"`
	Position *Position `json:"position"`
}

func (d *Device) Build() {
	if d.Space == nil && d.SpaceId != 0 {
		d.Space = &Space{Id: d.SpaceId, SpaceGroupId: d.SpaceGroupId}
	}
	if d.Breaker == nil && d.BreakerId != 0 {
		d.Breaker = &Breaker{Id: d.BreakerId}
	}
	if d.Position == nil && d.PositionX != nil {
		d.Position = &Position{X: *d.PositionX, Y: *d.PositionY, Z: *d.PositionZ}
	}
}

func ParseDeviceColumns(columnsStr string, allowPosition bool) (string, []uint8, bool, error) {
	if strings.TrimSpace(columnsStr) == "" {
		if allowPosition {
			return DEVICE_ALL_COLUMNS + "," + schema.FullDevicePositions_POSITION_X + "," + schema.FullDevicePositions_POSITION_Y + "," + schema.FullDevicePositions_POSITION_Z,
				[]uint8{0, 1, 2, 3, 4, 5}, true, nil
		}
		return DEVICE_ALL_COLUMNS, []uint8{0, 1, 2}, false, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

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
		case BASE_POSITION:
			if !allowPosition {
				return "", nil, false, fmt.Errorf("invalid column: %s", col)
			}
			collector.Append(3, schema.FullDevicePositions_POSITION_X)
			collector.Append(4, schema.FullDevicePositions_POSITION_Y)
			collector.Append(5, schema.FullDevicePositions_POSITION_Z)
			includePosition = true
		default:
			return "", nil, false, fmt.Errorf("invalid column: %s", col)
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
			values[i] = &d.PositionX
		case 4:
			values[i] = &d.PositionY
		case 5:
			values[i] = &d.PositionZ
		}
	}
	return values
}
