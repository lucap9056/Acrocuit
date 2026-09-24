package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const BREAKER_ALL_COLUMNS = schema.FullBreakers_ID + "," + schema.FullBreakers_NAME + "," + schema.FullBreakers_BREAKER_GROUP_ID + "," + schema.FullBreakers_SPACE_GROUP_ID + "," + schema.FullBreakers_DISPLAY_ORDER + "," + schema.FullBreakers_UPSTREAM_BREAKER_ID

type Breaker struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	DisplayOrder      int    `json:"display_order"`
	BreakerGroupId    int    `json:"-"`
	SpaceGroupId      int    `json:"-"`
	UpstreamBreakerId *int   `json:"-"`

	BreakerGroup   *BreakerGroup `json:"break_group"`
	UpsteamBreaker *Breaker      `json:"upstream_breaker"`
}

func (b *Breaker) Build() {
	if b.BreakerGroup == nil && b.BreakerGroupId != 0 {
		b.BreakerGroup = &BreakerGroup{Id: b.BreakerGroupId, SpaceGroupId: b.SpaceGroupId}
	}
	if b.UpsteamBreaker == nil && b.UpstreamBreakerId != nil {
		b.UpsteamBreaker = &Breaker{Id: *b.UpstreamBreakerId, SpaceGroupId: b.SpaceGroupId}
	}
}

func ParseBreakerColumns(columnsStr string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return BREAKER_ALL_COLUMNS, []uint8{0, 1, 2, 3, 4, 5}, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.Breakers_ID, schema.FullBreakers_ID:
			collector.Append(0, schema.FullBreakers_ID)
		case schema.Breakers_NAME, schema.FullBreakers_NAME:
			collector.Append(1, schema.FullBreakers_NAME)
		case schema.Breakers_BREAKER_GROUP_ID, schema.FullBreakers_BREAKER_GROUP_ID:
			collector.Append(2, schema.FullBreakers_BREAKER_GROUP_ID)
		case schema.Breakers_SPACE_GROUP_ID, schema.FullBreakers_SPACE_GROUP_ID:
			collector.Append(3, schema.FullBreakers_SPACE_GROUP_ID)
		case schema.Breakers_DISPLAY_ORDER, schema.FullBreakers_DISPLAY_ORDER:
			collector.Append(4, schema.FullBreakers_DISPLAY_ORDER)
		case schema.Breakers_UPSTREAM_BREAKER_ID, schema.FullBreakers_UPSTREAM_BREAKER_ID:
			collector.Append(5, schema.FullBreakers_UPSTREAM_BREAKER_ID)
		default:
			return "", nil, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, nil
}

func (b *Breaker) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &b.Id
		case 1:
			values[i] = &b.Name
		case 2:
			values[i] = &b.BreakerGroupId
		case 3:
			values[i] = &b.SpaceGroupId
		case 4:
			values[i] = &b.DisplayOrder
		case 5:
			values[i] = &b.UpstreamBreakerId
		}
	}
	return values
}
