package models

import (
	"acrocuit/internal/models/columns"
	"acrocuit/schema"
	"fmt"
	"strings"
)

type SpaceGroup struct {
	Id   SpaceGroupId `json:"id"`
	Name string       `json:"name"`
	Role string       `json:"role" tstype:"SpaceGroupRole"`
}

func ParseSpaceGroupColumns(columnsStr, roleColumn string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return SPACE_GROUP_ALL_COLUMNS + "," + roleColumn, []uint8{0, 1, 2}, nil
	}

	collector := columns.NewCollector(columnsStr)
	defer collector.Release()

	scanner := columns.NewScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.SpaceGroups_ID, schema.FullSpaceGroups_ID:
			collector.Append(0, schema.FullSpaceGroups_ID)
		case schema.SpaceGroups_NAME, schema.FullSpaceGroups_NAME:
			collector.Append(1, schema.FullSpaceGroups_NAME)
		case schema.SpaceGroupACLs_ROLE:
			collector.Append(2, roleColumn)
		default:
			return "", nil, fmt.Errorf("%w: %s", ErrInvalidColumn, col)
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
		case 2:
			values[i] = &sg.Role
		}
	}
	return values
}
