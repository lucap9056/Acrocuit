package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
	"time"
)

const SPACE_ALL_COLUMNS = schema.FullSpaces_ID + "," + schema.FullSpaces_NAME + "," + schema.FullSpaces_SPACE_GROUP_ID + "," + schema.FullSpaces_DISPLAY_ORDER + "," + schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT

type Space struct {
	Id                       int       `json:"id"`
	Name                     string    `json:"name"`
	DisplayOrder             int       `json:"display_order"`
	BackgroundImageUpdatedAt time.Time `json:"background_image_updated_at"`
	SpaceGroupId             int       `json:"-"`

	SpaceGroup *SpaceGroup `json:"space_group"`
}

func (s *Space) Build() {
	if s.SpaceGroup == nil && s.SpaceGroupId != 0 {
		s.SpaceGroup = &SpaceGroup{Id: s.SpaceGroupId}
	}
}

func ParseSpaceColumns(columnsStr string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return SPACE_ALL_COLUMNS, []uint8{0, 1, 2, 3, 4}, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.Spaces_ID, schema.FullSpaces_ID:
			collector.Append(0, schema.FullSpaces_ID)
		case schema.Spaces_NAME, schema.FullSpaces_NAME:
			collector.Append(1, schema.FullSpaces_NAME)
		case schema.Spaces_SPACE_GROUP_ID, schema.FullSpaces_SPACE_GROUP_ID:
			collector.Append(2, schema.FullSpaces_SPACE_GROUP_ID)
		case schema.Spaces_DISPLAY_ORDER, schema.FullSpaces_DISPLAY_ORDER:
			collector.Append(3, schema.FullSpaces_DISPLAY_ORDER)
		case schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT, schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT:
			collector.Append(4, schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT)
		default:
			return "", nil, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, nil
}

func (s *Space) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &s.Id
		case 1:
			values[i] = &s.Name
		case 2:
			values[i] = &s.SpaceGroupId
		case 3:
			values[i] = &s.DisplayOrder
		case 4:
			values[i] = &s.BackgroundImageUpdatedAt
		}
	}
	return values
}
