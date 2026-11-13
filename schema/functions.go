package schema

import (
	_ "embed"
	"io"
)

//go:embed "postgres/functions.sql"
var rawFunctions string

const (
	baseFunc             = "func"
	baseOutResult        = "o_result"
	baseOutID            = "o_id"
	baseOutName          = "o_name"
	baseOutSpaceGroupID  = "o_space_group_id"
	baseOutSwitchPanelID = "o_switch_panel_id"
	baseOutDisplayOrder  = "o_display_order"
	baseOutAttributes    = "o_attributes"
	baseErrSelfReference = "err_self_reference"
	baseErrCycleDetected = "err_cycle_detected"
	baseErrNotFound      = "err_not_found"

	FnSetSpace                               = "fn_set_space"
	FnSetSpace_O_ID                          = baseOutID
	FnSetSpace_O_NAME                        = baseOutName
	FnSetSpace_O_SPACE_GROUP_ID              = baseOutSpaceGroupID
	FnSetSpace_O_DISPLAY_ORDER               = baseOutDisplayOrder
	FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT = "o_background_image_updated_at"

	FnMoveSpace                               = "fn_move_space"
	FnMoveSpace_O_ID                          = baseOutID
	FnMoveSpace_O_NAME                        = baseOutName
	FnMoveSpace_O_SPACE_GROUP_ID              = baseOutSpaceGroupID
	FnMoveSpace_O_DISPLAY_ORDER               = baseOutDisplayOrder
	FnMoveSpace_O_BACKGROUND_IMAGE_UPDATED_AT = FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT

	FnDelSpace                               = "fn_del_space"
	FnDelSpace_O_ID                          = baseOutID
	FnDelSpace_O_NAME                        = baseOutName
	FnDelSpace_O_SPACE_GROUP_ID              = baseOutSpaceGroupID
	FnDelSpace_O_DISPLAY_ORDER               = baseOutDisplayOrder
	FnDelSpace_O_BACKGROUND_IMAGE_UPDATED_AT = FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT
	FnDelSpace_ERR_NOT_FOUND                 = "AC003"

	FnSetSwitch                   = "fn_set_switch"
	FnSetSwitch_O_ID              = baseOutID
	FnSetSwitch_O_NAME            = baseOutName
	FnSetSwitch_O_SWITCH_PANEL_ID = baseOutSwitchPanelID
	FnSetSwitch_O_SPACE_GROUP_ID  = baseOutSpaceGroupID
	FnSetSwitch_O_DISPLAY_ORDER   = baseOutDisplayOrder
	FnSetSwitch_O_ATTRIBUTES      = baseOutAttributes

	FnMoveSwitch                   = "fn_move_switch"
	FnMoveSwitch_O_ID              = baseOutID
	FnMoveSwitch_O_NAME            = baseOutName
	FnMoveSwitch_O_SWITCH_PANEL_ID = baseOutSwitchPanelID
	FnMoveSwitch_O_SPACE_GROUP_ID  = baseOutSpaceGroupID
	FnMoveSwitch_O_DISPLAY_ORDER   = baseOutDisplayOrder
	FnMoveSwitch_O_ATTRIBUTES      = baseOutAttributes

	FnDelSwitch                   = "fn_del_switch"
	FnDelSwitch_O_ID              = baseOutID
	FnDelSwitch_O_NAME            = baseOutName
	FnDelSwitch_O_SWITCH_PANEL_ID = baseOutSwitchPanelID
	FnDelSwitch_O_SPACE_GROUP_ID  = baseOutSpaceGroupID
	FnDelSwitch_O_DISPLAY_ORDER   = baseOutDisplayOrder
	FnDelSwitch_O_ATTRIBUTES      = baseOutAttributes
	FnDelSwitch_ERR_NOT_FOUND     = "AC003"

	FnAddSwitchUpstream                        = "fn_add_switch_upstream"
	FnAddSwitchUpstream_O_UPSTREAM_SWITCH_ID   = "o_upstream_switch_id"
	FnAddSwitchUpstream_O_DOWNSTREAM_SWITCH_ID = "o_downstream_switch_id"
	FnAddSwitchUpstream_O_ATTRIBUTES           = baseOutAttributes
	FnAddSwitchUpstream_ERR_SELF_REFERENCE     = "AC001"
	FnAddSwitchUpstream_ERR_CYCLE_DETECTED     = "AC002"

	FnAddDeviceSwitch          = "fn_add_device_switch"
	FnAddDeviceSwitch_O_RESULT = baseOutResult
)

const (
	FnSetSpace_OUTPUTS = FnSetSpace_O_ID + "," + FnSetSpace_O_NAME + "," + FnSetSpace_O_SPACE_GROUP_ID + "," +
		FnSetSpace_O_DISPLAY_ORDER + "," + FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT

	FnMoveSpace_OUTPUTS = FnMoveSpace_O_ID + "," + FnMoveSpace_O_NAME + "," + FnMoveSpace_O_SPACE_GROUP_ID + "," +
		FnMoveSpace_O_DISPLAY_ORDER + "," + FnMoveSpace_O_BACKGROUND_IMAGE_UPDATED_AT

	FnDelSpace_OUTPUTS = FnDelSpace_O_ID + "," + FnDelSpace_O_NAME + "," + FnDelSpace_O_SPACE_GROUP_ID + "," +
		FnDelSpace_O_DISPLAY_ORDER + "," + FnDelSpace_O_BACKGROUND_IMAGE_UPDATED_AT

	FnSetSwitch_OUTPUTS = FnSetSwitch_O_ID + "," + FnSetSwitch_O_NAME + "," + FnSetSwitch_O_SWITCH_PANEL_ID + "," +
		FnSetSwitch_O_SPACE_GROUP_ID + "," + FnSetSwitch_O_DISPLAY_ORDER + "," + FnSetSwitch_O_ATTRIBUTES

	FnMoveSwitch_OUTPUTS = FnMoveSwitch_O_ID + "," + FnMoveSwitch_O_NAME + "," + FnMoveSwitch_O_SWITCH_PANEL_ID + "," +
		FnMoveSwitch_O_SPACE_GROUP_ID + "," + FnMoveSwitch_O_DISPLAY_ORDER + "," + FnMoveSwitch_O_ATTRIBUTES

	FnDelSwitch_OUTPUTS = FnDelSwitch_O_ID + "," + FnDelSwitch_O_NAME + "," + FnDelSwitch_O_SWITCH_PANEL_ID + "," +
		FnDelSwitch_O_SPACE_GROUP_ID + "," + FnDelSwitch_O_DISPLAY_ORDER + "," + FnDelSwitch_O_ATTRIBUTES

	FnAddSwitchUpstream_OUTPUTS = FnAddSwitchUpstream_O_UPSTREAM_SWITCH_ID + "," + FnAddSwitchUpstream_O_DOWNSTREAM_SWITCH_ID + "," + FnAddSwitchUpstream_O_ATTRIBUTES
)

func init() {
	funcMap[FnSetSpace] = func() map[string]string {
		return map[string]string{
			baseFunc:                                 FnSetSpace,
			FnSetSpace_O_ID:                          FnSetSpace_O_ID,
			FnSetSpace_O_NAME:                        FnSetSpace_O_NAME,
			FnSetSpace_O_SPACE_GROUP_ID:              FnSetSpace_O_SPACE_GROUP_ID,
			FnSetSpace_O_DISPLAY_ORDER:               FnSetSpace_O_DISPLAY_ORDER,
			FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT: FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT,
		}
	}
	funcMap[FnMoveSpace] = func() map[string]string {
		return map[string]string{
			baseFunc:                                  FnMoveSpace,
			FnMoveSpace_O_ID:                          FnMoveSpace_O_ID,
			FnMoveSpace_O_NAME:                        FnMoveSpace_O_NAME,
			FnMoveSpace_O_SPACE_GROUP_ID:              FnMoveSpace_O_SPACE_GROUP_ID,
			FnMoveSpace_O_DISPLAY_ORDER:               FnMoveSpace_O_DISPLAY_ORDER,
			FnMoveSpace_O_BACKGROUND_IMAGE_UPDATED_AT: FnMoveSpace_O_BACKGROUND_IMAGE_UPDATED_AT,
		}
	}
	funcMap[FnDelSpace] = func() map[string]string {
		return map[string]string{
			baseFunc:                                 FnDelSpace,
			FnDelSpace_O_ID:                          FnDelSpace_O_ID,
			FnDelSpace_O_NAME:                        FnDelSpace_O_NAME,
			FnDelSpace_O_SPACE_GROUP_ID:              FnDelSpace_O_SPACE_GROUP_ID,
			FnDelSpace_O_DISPLAY_ORDER:               FnDelSpace_O_DISPLAY_ORDER,
			FnDelSpace_O_BACKGROUND_IMAGE_UPDATED_AT: FnDelSpace_O_BACKGROUND_IMAGE_UPDATED_AT,
			baseErrNotFound:                          FnDelSpace_ERR_NOT_FOUND,
		}
	}
	funcMap[FnSetSwitch] = func() map[string]string {
		return map[string]string{
			baseFunc:                      FnSetSwitch,
			FnSetSwitch_O_ID:              FnSetSwitch_O_ID,
			FnSetSwitch_O_NAME:            FnSetSwitch_O_NAME,
			FnSetSwitch_O_SWITCH_PANEL_ID: FnSetSwitch_O_SWITCH_PANEL_ID,
			FnSetSwitch_O_SPACE_GROUP_ID:  FnSetSwitch_O_SPACE_GROUP_ID,
			FnSetSwitch_O_DISPLAY_ORDER:   FnSetSwitch_O_DISPLAY_ORDER,
			FnSetSwitch_O_ATTRIBUTES:      FnSetSwitch_O_ATTRIBUTES,
		}
	}
	funcMap[FnMoveSwitch] = func() map[string]string {
		return map[string]string{
			baseFunc:                       FnMoveSwitch,
			FnMoveSwitch_O_ID:              FnMoveSwitch_O_ID,
			FnMoveSwitch_O_NAME:            FnMoveSwitch_O_NAME,
			FnMoveSwitch_O_SWITCH_PANEL_ID: FnMoveSwitch_O_SWITCH_PANEL_ID,
			FnMoveSwitch_O_SPACE_GROUP_ID:  FnMoveSwitch_O_SPACE_GROUP_ID,
			FnMoveSwitch_O_DISPLAY_ORDER:   FnMoveSwitch_O_DISPLAY_ORDER,
			FnMoveSwitch_O_ATTRIBUTES:      FnMoveSwitch_O_ATTRIBUTES,
		}
	}
	funcMap[FnDelSwitch] = func() map[string]string {
		return map[string]string{
			baseFunc:                      FnDelSwitch,
			FnDelSwitch_O_ID:              FnDelSwitch_O_ID,
			FnDelSwitch_O_NAME:            FnDelSwitch_O_NAME,
			FnDelSwitch_O_SWITCH_PANEL_ID: FnDelSwitch_O_SWITCH_PANEL_ID,
			FnDelSwitch_O_SPACE_GROUP_ID:  FnDelSwitch_O_SPACE_GROUP_ID,
			FnDelSwitch_O_DISPLAY_ORDER:   FnDelSwitch_O_DISPLAY_ORDER,
			FnDelSwitch_O_ATTRIBUTES:      FnDelSwitch_O_ATTRIBUTES,
			baseErrNotFound:               FnDelSwitch_ERR_NOT_FOUND,
		}
	}
	funcMap[FnAddSwitchUpstream] = func() map[string]string {
		return map[string]string{
			baseFunc:                                   FnAddSwitchUpstream,
			FnAddSwitchUpstream_O_UPSTREAM_SWITCH_ID:   FnAddSwitchUpstream_O_UPSTREAM_SWITCH_ID,
			FnAddSwitchUpstream_O_DOWNSTREAM_SWITCH_ID: FnAddSwitchUpstream_O_DOWNSTREAM_SWITCH_ID,
			FnAddSwitchUpstream_O_ATTRIBUTES:           FnAddSwitchUpstream_O_ATTRIBUTES,
			baseErrSelfReference:                       FnAddSwitchUpstream_ERR_SELF_REFERENCE,
			baseErrCycleDetected:                       FnAddSwitchUpstream_ERR_CYCLE_DETECTED,
		}
	}
	funcMap[FnAddDeviceSwitch] = func() map[string]string {
		return map[string]string{
			baseFunc:                   FnAddDeviceSwitch,
			FnAddDeviceSwitch_O_RESULT: FnAddDeviceSwitch_O_RESULT,
		}
	}
}

func RenderFunctions(wr io.Writer) error {
	return render("functions", rawFunctions, wr)
}
