package models

type SwitchLink struct {
	UpstreamSwitchId   SwitchId   `json:"upstream_switch_id"`
	DownstreamSwitchId SwitchId   `json:"downstream_switch_id"`
	Attributes         Attributes `json:"attributes" tstype:"Attributes | null"`
}

func (l *SwitchLink) Values() []any {
	return []any{&l.UpstreamSwitchId, &l.DownstreamSwitchId, &l.Attributes}
}

type DeviceLink struct {
	DeviceId   DeviceId   `json:"device_id"`
	SwitchId   SwitchId   `json:"switch_id"`
	Attributes Attributes `json:"attributes" tstype:"Attributes | null"`
}

func (l *DeviceLink) Values() []any {
	return []any{&l.DeviceId, &l.SwitchId, &l.Attributes}
}

type LinkInfo struct {
	Attributes Attributes `json:"attributes" tstype:"Attributes | null"`
}

type LinkedSwitch struct {
	Switch `tstype:",extends"`
	Link   LinkInfo `json:"link"`
}

func (ls *LinkedSwitch) Values(fields []uint8) []any {
	return append(ls.Switch.Values(fields), &ls.Link.Attributes)
}

type PowerChainGraph struct {
	Switches    []ChainSwitch `json:"switches"`
	SwitchLinks []SwitchLink  `json:"switch_links"`
	Devices     []Device      `json:"devices"`
	DeviceLinks []DeviceLink  `json:"device_links"`
}
