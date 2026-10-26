package models

import "acrocuit/schema"

const SWITCH_LINK_COLUMNS = schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + "," + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + "," + schema.FullSwitchLinks_ATTRIBUTES

const DEVICE_LINK_COLUMNS = schema.FullDeviceSwitches_DEVICE_ID + "," + schema.FullDeviceSwitches_SWITCH_ID + "," + schema.FullDeviceSwitches_ATTRIBUTES

type SwitchLink struct {
	UpstreamSwitchId   int        `json:"upstream_switch_id"`
	DownstreamSwitchId int        `json:"downstream_switch_id"`
	Attributes         Attributes `json:"attributes"`
}

func (l *SwitchLink) Values() []any {
	return []any{&l.UpstreamSwitchId, &l.DownstreamSwitchId, &l.Attributes}
}

type DeviceLink struct {
	DeviceId   int        `json:"device_id"`
	SwitchId   int        `json:"switch_id"`
	Attributes Attributes `json:"attributes"`
}

func (l *DeviceLink) Values() []any {
	return []any{&l.DeviceId, &l.SwitchId, &l.Attributes}
}

type LinkInfo struct {
	Attributes Attributes `json:"attributes"`
}

type LinkedSwitch struct {
	Switch
	Link LinkInfo `json:"link"`
}

func (ls *LinkedSwitch) Values(fields []uint8) []any {
	return append(ls.Switch.Values(fields), &ls.Link.Attributes)
}

type PowerChain struct {
	Switches    []Switch     `json:"switches"`
	SwitchLinks []SwitchLink `json:"switch_links"`
	Devices     []Device     `json:"devices"`
	DeviceLinks []DeviceLink `json:"device_links"`
}
