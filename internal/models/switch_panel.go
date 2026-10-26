package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const SWITCH_PANEL_ALL_COLUMNS = schema.FullSwitchPanels_ID + "," + schema.FullSwitchPanels_NAME + "," + schema.FullSwitchPanels_SPACE_ID + "," + schema.FullSwitchPanels_SPACE_GROUP_ID + "," + schema.FullSwitchPanels_ATTRIBUTES

const SWITCH_PANEL_WITH_POSITION_COLUMNS = SWITCH_PANEL_ALL_COLUMNS + "," + schema.FullSwitchPanelPositions_POSITION_X + "," + schema.FullSwitchPanelPositions_POSITION_Y + "," + schema.FullSwitchPanelPositions_POSITION_Z

var switchPanelWithPositionFields = []uint8{0, 1, 2, 3, 4, 5, 6, 7}

type SwitchPanel struct {
	Id           int        `json:"id"`
	Name         string     `json:"name"`
	Attributes   Attributes `json:"attributes"`
	SpaceId      int        `json:"-"`
	SpaceGroupId int        `json:"-"`
	PositionX    *int       `json:"-"`
	PositionY    *int       `json:"-"`
	PositionZ    *int       `json:"-"`

	Space    *Space    `json:"space"`
	Position *Position `json:"position"`
}

func (panel *SwitchPanel) Build() {
	if panel == nil {
		return
	}
	if panel.Space == nil && panel.SpaceId != 0 {
		panel.Space = &Space{Id: panel.SpaceId, SpaceGroupId: panel.SpaceGroupId}
	}
	if panel.Position == nil && panel.PositionX != nil {
		panel.Position = &Position{X: *panel.PositionX, Y: *panel.PositionY, Z: *panel.PositionZ}
	}
}

func ParseSwitchPanelColumns(columnsStr string, allowPosition bool) (string, []uint8, bool, error) {
	if strings.TrimSpace(columnsStr) == "" {
		if allowPosition {
			return SWITCH_PANEL_WITH_POSITION_COLUMNS, switchPanelWithPositionFields, true, nil
		}
		return SWITCH_PANEL_ALL_COLUMNS, []uint8{0, 1, 2, 3, 4}, false, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	includePosition := false

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.SwitchPanels_ID, schema.FullSwitchPanels_ID:
			collector.Append(0, schema.FullSwitchPanels_ID)
		case schema.SwitchPanels_NAME, schema.FullSwitchPanels_NAME:
			collector.Append(1, schema.FullSwitchPanels_NAME)
		case schema.SwitchPanels_SPACE_ID, schema.FullSwitchPanels_SPACE_ID:
			collector.Append(2, schema.FullSwitchPanels_SPACE_ID)
		case schema.SwitchPanels_SPACE_GROUP_ID, schema.FullSwitchPanels_SPACE_GROUP_ID:
			collector.Append(3, schema.FullSwitchPanels_SPACE_GROUP_ID)
		case schema.SwitchPanels_ATTRIBUTES, schema.FullSwitchPanels_ATTRIBUTES:
			collector.Append(4, schema.FullSwitchPanels_ATTRIBUTES)
		case BASE_POSITION:
			if !allowPosition {
				return "", nil, false, fmt.Errorf("invalid column: %s", col)
			}
			collector.Append(5, schema.FullSwitchPanelPositions_POSITION_X)
			collector.Append(6, schema.FullSwitchPanelPositions_POSITION_Y)
			collector.Append(7, schema.FullSwitchPanelPositions_POSITION_Z)
			includePosition = true
		default:
			return "", nil, false, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, includePosition, nil
}

func (panel *SwitchPanel) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &panel.Id
		case 1:
			values[i] = &panel.Name
		case 2:
			values[i] = &panel.SpaceId
		case 3:
			values[i] = &panel.SpaceGroupId
		case 4:
			values[i] = &panel.Attributes
		case 5:
			values[i] = &panel.PositionX
		case 6:
			values[i] = &panel.PositionY
		case 7:
			values[i] = &panel.PositionZ
		}
	}
	return values
}

func (panel *SwitchPanel) ValuesWithPosition() []any {
	return panel.Values(switchPanelWithPositionFields)
}
