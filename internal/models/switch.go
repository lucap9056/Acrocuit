package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const SWITCH_ALL_COLUMNS = schema.FullSwitches_ID + "," + schema.FullSwitches_NAME + "," + schema.FullSwitches_SWITCH_PANEL_ID + "," + schema.FullSwitches_SPACE_GROUP_ID + "," + schema.FullSwitches_DISPLAY_ORDER + "," + schema.FullSwitches_ATTRIBUTES

var switchAllFields = []uint8{0, 1, 2, 3, 4, 5}

type Switch struct {
	Id            int        `json:"id"`
	Name          string     `json:"name"`
	DisplayOrder  int        `json:"display_order"`
	Attributes    Attributes `json:"attributes"`
	SwitchPanelId int        `json:"-"`
	SpaceGroupId  int        `json:"-"`

	SwitchPanel *SwitchPanel `json:"switch_panel"`
}

func (s *Switch) Build() {
	if s.SwitchPanel == nil && s.SwitchPanelId != 0 {
		s.SwitchPanel = &SwitchPanel{Id: s.SwitchPanelId, SpaceGroupId: s.SpaceGroupId}
	}
	s.SwitchPanel.Build()
}

func (s *Switch) AllValues() []any {
	return s.Values(switchAllFields)
}

func (s *Switch) ValuesWithPanel() []any {
	s.SwitchPanel = &SwitchPanel{}
	return append(s.Values(switchAllFields), s.SwitchPanel.Values(switchPanelWithPositionFields)...)
}

func ParseSwitchColumns(columnsStr string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return SWITCH_ALL_COLUMNS, switchAllFields, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.Switches_ID, schema.FullSwitches_ID:
			collector.Append(0, schema.FullSwitches_ID)
		case schema.Switches_NAME, schema.FullSwitches_NAME:
			collector.Append(1, schema.FullSwitches_NAME)
		case schema.Switches_SWITCH_PANEL_ID, schema.FullSwitches_SWITCH_PANEL_ID:
			collector.Append(2, schema.FullSwitches_SWITCH_PANEL_ID)
		case schema.Switches_SPACE_GROUP_ID, schema.FullSwitches_SPACE_GROUP_ID:
			collector.Append(3, schema.FullSwitches_SPACE_GROUP_ID)
		case schema.Switches_DISPLAY_ORDER, schema.FullSwitches_DISPLAY_ORDER:
			collector.Append(4, schema.FullSwitches_DISPLAY_ORDER)
		case schema.Switches_ATTRIBUTES, schema.FullSwitches_ATTRIBUTES:
			collector.Append(5, schema.FullSwitches_ATTRIBUTES)
		default:
			return "", nil, fmt.Errorf("%w: %s", ErrInvalidColumn, col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, nil
}

func (s *Switch) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &s.Id
		case 1:
			values[i] = &s.Name
		case 2:
			values[i] = &s.SwitchPanelId
		case 3:
			values[i] = &s.SpaceGroupId
		case 4:
			values[i] = &s.DisplayOrder
		case 5:
			values[i] = &s.Attributes
		}
	}
	return values
}
