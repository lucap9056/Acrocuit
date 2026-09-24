package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const BREAKER_GROUP_ALL_COLUMNS = schema.FullBreakerGroups_ID + "," + schema.FullBreakerGroups_NAME + "," + schema.FullBreakerGroups_SPACE_ID + "," + schema.FullBreakerGroups_SPACE_GROUP_ID

type BreakerGroup struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	SpaceId      int    `json:"-"`
	SpaceGroupId int    `json:"-"`
	PositionX    *int   `json:"-"`
	PositionY    *int   `json:"-"`
	PositionZ    *int   `json:"-"`

	Space    *Space    `json:"space"`
	Position *Position `json:"position"`
}

func (bg *BreakerGroup) Build() {
	if bg.Space == nil && bg.SpaceId != 0 {
		bg.Space = &Space{Id: bg.SpaceId, SpaceGroupId: bg.SpaceGroupId}
	}
	if bg.Position == nil && bg.PositionX != nil {
		bg.Position = &Position{X: *bg.PositionX, Y: *bg.PositionY, Z: *bg.PositionZ}
	}
}

func ParseBreakerGroupColumns(columnsStr string, allowPosition bool) (string, []uint8, bool, error) {
	if strings.TrimSpace(columnsStr) == "" {
		if allowPosition {
			return BREAKER_GROUP_ALL_COLUMNS + "," + schema.FullBreakerGroupPositions_POSITION_X + "," + schema.FullBreakerGroupPositions_POSITION_Y + "," + schema.FullBreakerGroupPositions_POSITION_Z,
				[]uint8{0, 1, 2, 3, 4, 5, 6}, true, nil
		}
		return BREAKER_GROUP_ALL_COLUMNS, []uint8{0, 1, 2, 3}, false, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	includePosition := false

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.BreakerGroups_ID, schema.FullBreakerGroups_ID:
			collector.Append(0, schema.FullBreakerGroups_ID)
		case schema.BreakerGroups_NAME, schema.FullBreakerGroups_NAME:
			collector.Append(1, schema.FullBreakerGroups_NAME)
		case schema.BreakerGroups_SPACE_ID, schema.FullBreakerGroups_SPACE_ID:
			collector.Append(2, schema.FullBreakerGroups_SPACE_ID)
		case schema.BreakerGroups_SPACE_GROUP_ID, schema.FullBreakerGroups_SPACE_GROUP_ID:
			collector.Append(3, schema.FullBreakerGroups_SPACE_GROUP_ID)
		case BASE_POSITION:
			if !allowPosition {
				return "", nil, false, fmt.Errorf("invalid column: %s", col)
			}
			collector.Append(4, schema.FullBreakerGroupPositions_POSITION_X)
			collector.Append(5, schema.FullBreakerGroupPositions_POSITION_Y)
			collector.Append(6, schema.FullBreakerGroupPositions_POSITION_Z)
			includePosition = true
		default:
			return "", nil, false, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, includePosition, nil
}

func (bg *BreakerGroup) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &bg.Id
		case 1:
			values[i] = &bg.Name
		case 2:
			values[i] = &bg.SpaceId
		case 3:
			values[i] = &bg.SpaceGroupId
		case 4:
			values[i] = &bg.PositionX
		case 5:
			values[i] = &bg.PositionY
		case 6:
			values[i] = &bg.PositionZ
		}
	}
	return values
}
