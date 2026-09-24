package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const SPACE_GROUP_ALL_COLUMNS = schema.FullSpaceGroups_ID + "," + schema.FullSpaceGroups_NAME

type SpaceGroup struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func ParseSpaceGroupColumns(columnsStr string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return SPACE_GROUP_ALL_COLUMNS, []uint8{0, 1}, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.SpaceGroups_ID, schema.FullSpaceGroups_ID:
			collector.Append(0, schema.FullSpaceGroups_ID)
		case schema.SpaceGroups_NAME, schema.FullSpaceGroups_NAME:
			collector.Append(1, schema.FullSpaceGroups_NAME)
		default:
			return "", nil, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, nil
}

func (sg *SpaceGroup) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &sg.Id
		case 1:
			values[i] = &sg.Name
		}
	}
	return values
}
