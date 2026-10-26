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

	FnSetSwitch_OUTPUTS = FnSetSwitch_O_ID + "," + FnSetSwitch_O_NAME + "," + FnSetSwitch_O_SWITCH_PANEL_ID + "," +
		FnSetSwitch_O_SPACE_GROUP_ID + "," + FnSetSwitch_O_DISPLAY_ORDER + "," + FnSetSwitch_O_ATTRIBUTES

	FnMoveSwitch_OUTPUTS = FnMoveSwitch_O_ID + "," + FnMoveSwitch_O_NAME + "," + FnMoveSwitch_O_SWITCH_PANEL_ID + "," +
		FnMoveSwitch_O_SPACE_GROUP_ID + "," + FnMoveSwitch_O_DISPLAY_ORDER + "," + FnMoveSwitch_O_ATTRIBUTES

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
