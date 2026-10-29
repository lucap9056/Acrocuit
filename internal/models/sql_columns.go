package models

import "acrocuit/schema"

const DEVICE_ALL_COLUMNS = schema.FullDevices_ID + "," + schema.FullDevices_NAME + "," + schema.FullDevices_SPACE_ID + "," + schema.FullDevices_ATTRIBUTES

const DEVICE_WITH_POSITION_COLUMNS = DEVICE_ALL_COLUMNS + "," + schema.FullDevicePositions_POSITION_X + "," + schema.FullDevicePositions_POSITION_Y + "," + schema.FullDevicePositions_POSITION_Z

const SWITCH_LINK_COLUMNS = schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + "," + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + "," + schema.FullSwitchLinks_ATTRIBUTES

const DEVICE_LINK_COLUMNS = schema.FullDeviceSwitches_DEVICE_ID + "," + schema.FullDeviceSwitches_SWITCH_ID + "," + schema.FullDeviceSwitches_ATTRIBUTES

const SWITCH_ALL_COLUMNS = schema.FullSwitches_ID + "," + schema.FullSwitches_NAME + "," + schema.FullSwitches_SWITCH_PANEL_ID + "," + schema.FullSwitches_SPACE_GROUP_ID + "," + schema.FullSwitches_DISPLAY_ORDER + "," + schema.FullSwitches_ATTRIBUTES

const SPACE_ALL_COLUMNS = schema.FullSpaces_ID + "," + schema.FullSpaces_NAME + "," + schema.FullSpaces_SPACE_GROUP_ID + "," + schema.FullSpaces_DISPLAY_ORDER + "," + schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT

const SPACE_GROUP_ALL_COLUMNS = schema.FullSpaceGroups_ID + "," + schema.FullSpaceGroups_NAME

const SWITCH_PANEL_ALL_COLUMNS = schema.FullSwitchPanels_ID + "," + schema.FullSwitchPanels_NAME + "," + schema.FullSwitchPanels_SPACE_ID + "," + schema.FullSwitchPanels_SPACE_GROUP_ID + "," + schema.FullSwitchPanels_ATTRIBUTES

const SWITCH_PANEL_WITH_POSITION_COLUMNS = SWITCH_PANEL_ALL_COLUMNS + "," + schema.FullSwitchPanelPositions_POSITION_X + "," + schema.FullSwitchPanelPositions_POSITION_Y + "," + schema.FullSwitchPanelPositions_POSITION_Z

const BASE_POSITION = "position"
