package schema

import (
	_ "embed"
	"io"
	"text/template"
)

//go:embed "postgres/schema.sql"
var rawPostgresSchema string

//go:embed "sqlite/schema.sql"
var rawSQLiteSchema string

const (
	baseTable         = "table"
	baseId            = "id"
	baseName          = "name"
	baseSpaceGroupID  = "space_group_id"
	baseSpaceID       = "space_id"
	baseSwitchPanelID = "switch_panel_id"
	baseDisplayOrder  = "display_order"
	basePositionX     = "position_x"
	basePositionY     = "position_y"
	basePositionZ     = "position_z"
	baseDeviceID      = "device_id"
	baseSwitchID      = "switch_id"
	baseAttributes    = "attributes"

	AttributesMaxBytes = 2048

	SpaceGroupsTable = "space_groups"
	SpaceGroups_ID   = baseId
	SpaceGroups_NAME = baseName

	SpacesTable                        = "spaces"
	Spaces_ID                          = baseId
	Spaces_NAME                        = baseName
	Spaces_SPACE_GROUP_ID              = baseSpaceGroupID
	Spaces_DISPLAY_ORDER               = baseDisplayOrder
	Spaces_BACKGROUND_IMAGE_UPDATED_AT = "background_image_updated_at"

	SwitchPanelsTable           = "switch_panels"
	SwitchPanels_ID             = baseId
	SwitchPanels_NAME           = baseName
	SwitchPanels_SPACE_ID       = baseSpaceID
	SwitchPanels_SPACE_GROUP_ID = baseSpaceGroupID
	SwitchPanels_ATTRIBUTES     = baseAttributes

	SwitchPanelPositionsTable            = "switch_panel_positions"
	SwitchPanelPositions_SWITCH_PANEL_ID = baseSwitchPanelID
	SwitchPanelPositions_POSITION_X      = basePositionX
	SwitchPanelPositions_POSITION_Y      = basePositionY
	SwitchPanelPositions_POSITION_Z      = basePositionZ

	SwitchesTable            = "switches"
	Switches_ID              = baseId
	Switches_NAME            = baseName
	Switches_SWITCH_PANEL_ID = baseSwitchPanelID
	Switches_SPACE_GROUP_ID  = baseSpaceGroupID
	Switches_DISPLAY_ORDER   = baseDisplayOrder
	Switches_ATTRIBUTES      = baseAttributes

	SwitchLinksTable                 = "switch_links"
	SwitchLinks_UPSTREAM_SWITCH_ID   = "upstream_switch_id"
	SwitchLinks_DOWNSTREAM_SWITCH_ID = "downstream_switch_id"
	SwitchLinks_SPACE_GROUP_ID       = baseSpaceGroupID
	SwitchLinks_ATTRIBUTES           = baseAttributes

	SpaceBackgroundImagesTable         = "space_background_images"
	SpaceBackgroundImages_SPACE_ID     = baseSpaceID
	SpaceBackgroundImages_CONTENT_TYPE = "content_type"
	SpaceBackgroundImages_DATA         = "data"

	DevicesTable       = "devices"
	Devices_ID         = baseId
	Devices_NAME       = baseName
	Devices_SPACE_ID   = baseSpaceID
	Devices_ATTRIBUTES = baseAttributes

	DevicePositionsTable       = "device_positions"
	DevicePositions_DEVICE_ID  = baseDeviceID
	DevicePositions_POSITION_X = basePositionX
	DevicePositions_POSITION_Y = basePositionY
	DevicePositions_POSITION_Z = basePositionZ

	DeviceSwitchesTable       = "device_switches"
	DeviceSwitches_DEVICE_ID  = baseDeviceID
	DeviceSwitches_SWITCH_ID  = baseSwitchID
	DeviceSwitches_ATTRIBUTES = baseAttributes

	SpaceGroupOwnersTable           = "space_group_owners"
	SpaceGroupOwners_USER_EMAIL     = "user_email"
	SpaceGroupOwners_SPACE_GROUP_ID = baseSpaceGroupID
)

const (
	FullSpaceGroups_ID   = SpaceGroupsTable + "." + SpaceGroups_ID
	FullSpaceGroups_NAME = SpaceGroupsTable + "." + SpaceGroups_NAME

	FullSpaces_ID                          = SpacesTable + "." + Spaces_ID
	FullSpaces_NAME                        = SpacesTable + "." + Spaces_NAME
	FullSpaces_SPACE_GROUP_ID              = SpacesTable + "." + Spaces_SPACE_GROUP_ID
	FullSpaces_DISPLAY_ORDER               = SpacesTable + "." + Spaces_DISPLAY_ORDER
	FullSpaces_BACKGROUND_IMAGE_UPDATED_AT = SpacesTable + "." + Spaces_BACKGROUND_IMAGE_UPDATED_AT

	FullSwitchPanels_ID             = SwitchPanelsTable + "." + SwitchPanels_ID
	FullSwitchPanels_NAME           = SwitchPanelsTable + "." + SwitchPanels_NAME
	FullSwitchPanels_SPACE_ID       = SwitchPanelsTable + "." + SwitchPanels_SPACE_ID
	FullSwitchPanels_SPACE_GROUP_ID = SwitchPanelsTable + "." + SwitchPanels_SPACE_GROUP_ID
	FullSwitchPanels_ATTRIBUTES     = SwitchPanelsTable + "." + SwitchPanels_ATTRIBUTES

	FullSwitchPanelPositions_SWITCH_PANEL_ID = SwitchPanelPositionsTable + "." + SwitchPanelPositions_SWITCH_PANEL_ID
	FullSwitchPanelPositions_POSITION_X      = SwitchPanelPositionsTable + "." + SwitchPanelPositions_POSITION_X
	FullSwitchPanelPositions_POSITION_Y      = SwitchPanelPositionsTable + "." + SwitchPanelPositions_POSITION_Y
	FullSwitchPanelPositions_POSITION_Z      = SwitchPanelPositionsTable + "." + SwitchPanelPositions_POSITION_Z

	FullSwitches_ID              = SwitchesTable + "." + Switches_ID
	FullSwitches_NAME            = SwitchesTable + "." + Switches_NAME
	FullSwitches_SWITCH_PANEL_ID = SwitchesTable + "." + Switches_SWITCH_PANEL_ID
	FullSwitches_SPACE_GROUP_ID  = SwitchesTable + "." + Switches_SPACE_GROUP_ID
	FullSwitches_DISPLAY_ORDER   = SwitchesTable + "." + Switches_DISPLAY_ORDER
	FullSwitches_ATTRIBUTES      = SwitchesTable + "." + Switches_ATTRIBUTES

	FullSwitchLinks_UPSTREAM_SWITCH_ID   = SwitchLinksTable + "." + SwitchLinks_UPSTREAM_SWITCH_ID
	FullSwitchLinks_DOWNSTREAM_SWITCH_ID = SwitchLinksTable + "." + SwitchLinks_DOWNSTREAM_SWITCH_ID
	FullSwitchLinks_SPACE_GROUP_ID       = SwitchLinksTable + "." + SwitchLinks_SPACE_GROUP_ID
	FullSwitchLinks_ATTRIBUTES           = SwitchLinksTable + "." + SwitchLinks_ATTRIBUTES

	FullSpaceBackgroundImages_SPACE_ID     = SpaceBackgroundImagesTable + "." + SpaceBackgroundImages_SPACE_ID
	FullSpaceBackgroundImages_CONTENT_TYPE = SpaceBackgroundImagesTable + "." + SpaceBackgroundImages_CONTENT_TYPE
	FullSpaceBackgroundImages_DATA         = SpaceBackgroundImagesTable + "." + SpaceBackgroundImages_DATA

	FullDevices_ID         = DevicesTable + "." + Devices_ID
	FullDevices_NAME       = DevicesTable + "." + Devices_NAME
	FullDevices_SPACE_ID   = DevicesTable + "." + Devices_SPACE_ID
	FullDevices_ATTRIBUTES = DevicesTable + "." + Devices_ATTRIBUTES

	FullDevicePositions_DEVICE_ID  = DevicePositionsTable + "." + DevicePositions_DEVICE_ID
	FullDevicePositions_POSITION_X = DevicePositionsTable + "." + DevicePositions_POSITION_X
	FullDevicePositions_POSITION_Y = DevicePositionsTable + "." + DevicePositions_POSITION_Y
	FullDevicePositions_POSITION_Z = DevicePositionsTable + "." + DevicePositions_POSITION_Z

	FullDeviceSwitches_DEVICE_ID  = DeviceSwitchesTable + "." + DeviceSwitches_DEVICE_ID
	FullDeviceSwitches_SWITCH_ID  = DeviceSwitchesTable + "." + DeviceSwitches_SWITCH_ID
	FullDeviceSwitches_ATTRIBUTES = DeviceSwitchesTable + "." + DeviceSwitches_ATTRIBUTES

	FullSpaceGroupOwners_USER_EMAIL     = SpaceGroupOwnersTable + "." + SpaceGroupOwners_USER_EMAIL
	FullSpaceGroupOwners_SPACE_GROUP_ID = SpaceGroupOwnersTable + "." + SpaceGroupOwners_SPACE_GROUP_ID
)

const (
	JoinSwitchPanels_Spaces               = FullSwitchPanels_SPACE_ID + " = " + FullSpaces_ID
	JoinSpaceGroupOwners_Spaces           = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSpaces_SPACE_GROUP_ID
	JoinSwitchPanelPositions_SwitchPanels = FullSwitchPanelPositions_SWITCH_PANEL_ID + " = " + FullSwitchPanels_ID
	JoinSwitches_SwitchPanels             = FullSwitches_SWITCH_PANEL_ID + " = " + FullSwitchPanels_ID
	JoinSpaceGroupOwners_SwitchPanels     = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSwitchPanels_SPACE_GROUP_ID
	JoinSpaceGroupOwners_Switches         = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSwitches_SPACE_GROUP_ID
	JoinSpaceBackgroundImages_Spaces      = FullSpaceBackgroundImages_SPACE_ID + " = " + FullSpaces_ID
	JoinDevices_Spaces                    = FullDevices_SPACE_ID + " = " + FullSpaces_ID
	JoinDevicePositions_Devices           = FullDevicePositions_DEVICE_ID + " = " + FullDevices_ID
	JoinDeviceSwitches_Devices            = FullDeviceSwitches_DEVICE_ID + " = " + FullDevices_ID
	JoinDeviceSwitches_Switches           = FullDeviceSwitches_SWITCH_ID + " = " + FullSwitches_ID
	JoinSpaceGroupOwners_SwitchLinks      = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSwitchLinks_SPACE_GROUP_ID
	JoinSpaceGroupOwners_SpaceGroups      = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSpaceGroups_ID
)

var funcMap = template.FuncMap{
	"attributes_max_bytes": func() int {
		return AttributesMaxBytes
	},
	SpaceGroupsTable: func() map[string]string {
		return map[string]string{
			baseTable:        SpaceGroupsTable,
			SpaceGroups_ID:   SpaceGroups_ID,
			SpaceGroups_NAME: SpaceGroups_NAME,
		}
	},
	SpacesTable: func() map[string]string {
		return map[string]string{
			baseTable:                          SpacesTable,
			Spaces_ID:                          Spaces_ID,
			Spaces_NAME:                        Spaces_NAME,
			Spaces_SPACE_GROUP_ID:              Spaces_SPACE_GROUP_ID,
			Spaces_DISPLAY_ORDER:               Spaces_DISPLAY_ORDER,
			Spaces_BACKGROUND_IMAGE_UPDATED_AT: Spaces_BACKGROUND_IMAGE_UPDATED_AT,
		}
	},
	SwitchPanelsTable: func() map[string]string {
		return map[string]string{
			baseTable:                   SwitchPanelsTable,
			SwitchPanels_ID:             SwitchPanels_ID,
			SwitchPanels_NAME:           SwitchPanels_NAME,
			SwitchPanels_SPACE_ID:       SwitchPanels_SPACE_ID,
			SwitchPanels_SPACE_GROUP_ID: SwitchPanels_SPACE_GROUP_ID,
			SwitchPanels_ATTRIBUTES:     SwitchPanels_ATTRIBUTES,
		}
	},
	SwitchPanelPositionsTable: func() map[string]string {
		return map[string]string{
			baseTable:                            SwitchPanelPositionsTable,
			SwitchPanelPositions_SWITCH_PANEL_ID: SwitchPanelPositions_SWITCH_PANEL_ID,
			SwitchPanelPositions_POSITION_X:      SwitchPanelPositions_POSITION_X,
			SwitchPanelPositions_POSITION_Y:      SwitchPanelPositions_POSITION_Y,
			SwitchPanelPositions_POSITION_Z:      SwitchPanelPositions_POSITION_Z,
		}
	},
	SwitchesTable: func() map[string]string {
		return map[string]string{
			baseTable:                SwitchesTable,
			Switches_ID:              Switches_ID,
			Switches_NAME:            Switches_NAME,
			Switches_SWITCH_PANEL_ID: Switches_SWITCH_PANEL_ID,
			Switches_SPACE_GROUP_ID:  Switches_SPACE_GROUP_ID,
			Switches_DISPLAY_ORDER:   Switches_DISPLAY_ORDER,
			Switches_ATTRIBUTES:      Switches_ATTRIBUTES,
		}
	},
	SwitchLinksTable: func() map[string]string {
		return map[string]string{
			baseTable:                        SwitchLinksTable,
			SwitchLinks_UPSTREAM_SWITCH_ID:   SwitchLinks_UPSTREAM_SWITCH_ID,
			SwitchLinks_DOWNSTREAM_SWITCH_ID: SwitchLinks_DOWNSTREAM_SWITCH_ID,
			SwitchLinks_SPACE_GROUP_ID:       SwitchLinks_SPACE_GROUP_ID,
			SwitchLinks_ATTRIBUTES:           SwitchLinks_ATTRIBUTES,
		}
	},
	SpaceBackgroundImagesTable: func() map[string]string {
		return map[string]string{
			baseTable:                          SpaceBackgroundImagesTable,
			SpaceBackgroundImages_SPACE_ID:     SpaceBackgroundImages_SPACE_ID,
			SpaceBackgroundImages_CONTENT_TYPE: SpaceBackgroundImages_CONTENT_TYPE,
			SpaceBackgroundImages_DATA:         SpaceBackgroundImages_DATA,
		}
	},
	DevicesTable: func() map[string]string {
		return map[string]string{
			baseTable:          DevicesTable,
			Devices_ID:         Devices_ID,
			Devices_NAME:       Devices_NAME,
			Devices_SPACE_ID:   Devices_SPACE_ID,
			Devices_ATTRIBUTES: Devices_ATTRIBUTES,
		}
	},
	DevicePositionsTable: func() map[string]string {
		return map[string]string{
			baseTable:                  DevicePositionsTable,
			DevicePositions_DEVICE_ID:  DevicePositions_DEVICE_ID,
			DevicePositions_POSITION_X: DevicePositions_POSITION_X,
			DevicePositions_POSITION_Y: DevicePositions_POSITION_Y,
			DevicePositions_POSITION_Z: DevicePositions_POSITION_Z,
		}
	},
	DeviceSwitchesTable: func() map[string]string {
		return map[string]string{
			baseTable:                 DeviceSwitchesTable,
			DeviceSwitches_DEVICE_ID:  DeviceSwitches_DEVICE_ID,
			DeviceSwitches_SWITCH_ID:  DeviceSwitches_SWITCH_ID,
			DeviceSwitches_ATTRIBUTES: DeviceSwitches_ATTRIBUTES,
		}
	},
	SpaceGroupOwnersTable: func() map[string]string {
		return map[string]string{
			baseTable:                       SpaceGroupOwnersTable,
			SpaceGroupOwners_USER_EMAIL:     SpaceGroupOwners_USER_EMAIL,
			SpaceGroupOwners_SPACE_GROUP_ID: SpaceGroupOwners_SPACE_GROUP_ID,
		}
	},
}

func render(name, raw string, wr io.Writer) error {
	tmpl, err := template.New(name).Option("missingkey=error").Funcs(funcMap).Parse(raw)
	if err != nil {
		return err
	}

	return tmpl.Execute(wr, nil)
}

func RenderPostgres(wr io.Writer) error {
	return render("postgres", rawPostgresSchema, wr)
}

func RenderSQLite(wr io.Writer) error {
	return render("sqlite", rawSQLiteSchema, wr)
}
