package schema

import (
	_ "embed"
	"io"
)

//go:embed "postgres/functions.sql"
var rawFunctions string

const (
	baseFunc                 = "func"
	baseOutResult            = "o_result"
	baseOutID                = "o_id"
	baseOutName              = "o_name"
	baseOutSpaceGroupID      = "o_space_group_id"
	baseOutBreakerGroupID    = "o_breaker_group_id"
	baseOutDisplayOrder      = "o_display_order"
	baseOutUpstreamBreakerID = "o_upstream_breaker_id"
	baseErrSelfReference     = "err_self_reference"
	baseErrCycleDetected     = "err_cycle_detected"

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

	FnSetBreaker                       = "fn_set_breaker"
	FnSetBreaker_O_ID                  = baseOutID
	FnSetBreaker_O_NAME                = baseOutName
	FnSetBreaker_O_BREAKER_GROUP_ID    = baseOutBreakerGroupID
	FnSetBreaker_O_SPACE_GROUP_ID      = baseOutSpaceGroupID
	FnSetBreaker_O_DISPLAY_ORDER       = baseOutDisplayOrder
	FnSetBreaker_O_UPSTREAM_BREAKER_ID = baseOutUpstreamBreakerID

	FnMoveBreaker                       = "fn_move_breaker"
	FnMoveBreaker_O_ID                  = baseOutID
	FnMoveBreaker_O_NAME                = baseOutName
	FnMoveBreaker_O_BREAKER_GROUP_ID    = baseOutBreakerGroupID
	FnMoveBreaker_O_SPACE_GROUP_ID      = baseOutSpaceGroupID
	FnMoveBreaker_O_DISPLAY_ORDER       = baseOutDisplayOrder
	FnMoveBreaker_O_UPSTREAM_BREAKER_ID = baseOutUpstreamBreakerID

	FnSetBreakerUpstream                       = "fn_set_breaker_upstream"
	FnSetBreakerUpstream_O_ID                  = baseOutID
	FnSetBreakerUpstream_O_NAME                = baseOutName
	FnSetBreakerUpstream_O_BREAKER_GROUP_ID    = baseOutBreakerGroupID
	FnSetBreakerUpstream_O_SPACE_GROUP_ID      = baseOutSpaceGroupID
	FnSetBreakerUpstream_O_DISPLAY_ORDER       = baseOutDisplayOrder
	FnSetBreakerUpstream_O_UPSTREAM_BREAKER_ID = baseOutUpstreamBreakerID
	FnSetBreakerUpstream_ERR_SELF_REFERENCE    = "AC001"
	FnSetBreakerUpstream_ERR_CYCLE_DETECTED    = "AC002"

	FnAddDeviceBreaker          = "fn_add_device_breaker"
	FnAddDeviceBreaker_O_RESULT = baseOutResult
)

const (
	FnSetSpace_OUTPUTS = FnSetSpace_O_ID + "," + FnSetSpace_O_NAME + "," + FnSetSpace_O_SPACE_GROUP_ID + "," +
		FnSetSpace_O_DISPLAY_ORDER + "," + FnSetSpace_O_BACKGROUND_IMAGE_UPDATED_AT

	FnMoveSpace_OUTPUTS = FnMoveSpace_O_ID + "," + FnMoveSpace_O_NAME + "," + FnMoveSpace_O_SPACE_GROUP_ID + "," +
		FnMoveSpace_O_DISPLAY_ORDER + "," + FnMoveSpace_O_BACKGROUND_IMAGE_UPDATED_AT

	FnSetBreaker_OUTPUTS = FnSetBreaker_O_ID + "," + FnSetBreaker_O_NAME + "," + FnSetBreaker_O_BREAKER_GROUP_ID + "," +
		FnSetBreaker_O_SPACE_GROUP_ID + "," + FnSetBreaker_O_DISPLAY_ORDER + "," + FnSetBreaker_O_UPSTREAM_BREAKER_ID

	FnMoveBreaker_OUTPUTS = FnMoveBreaker_O_ID + "," + FnMoveBreaker_O_NAME + "," + FnMoveBreaker_O_BREAKER_GROUP_ID + "," +
		FnMoveBreaker_O_SPACE_GROUP_ID + "," + FnMoveBreaker_O_DISPLAY_ORDER + "," + FnMoveBreaker_O_UPSTREAM_BREAKER_ID

	FnSetBreakerUpstream_OUTPUTS = FnSetBreakerUpstream_O_ID + "," + FnSetBreakerUpstream_O_NAME + "," + FnSetBreakerUpstream_O_BREAKER_GROUP_ID + "," +
		FnSetBreakerUpstream_O_SPACE_GROUP_ID + "," + FnSetBreakerUpstream_O_DISPLAY_ORDER + "," + FnSetBreakerUpstream_O_UPSTREAM_BREAKER_ID
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
	funcMap[FnSetBreaker] = func() map[string]string {
		return map[string]string{
			baseFunc:                           FnSetBreaker,
			FnSetBreaker_O_ID:                  FnSetBreaker_O_ID,
			FnSetBreaker_O_NAME:                FnSetBreaker_O_NAME,
			FnSetBreaker_O_BREAKER_GROUP_ID:    FnSetBreaker_O_BREAKER_GROUP_ID,
			FnSetBreaker_O_SPACE_GROUP_ID:      FnSetBreaker_O_SPACE_GROUP_ID,
			FnSetBreaker_O_DISPLAY_ORDER:       FnSetBreaker_O_DISPLAY_ORDER,
			FnSetBreaker_O_UPSTREAM_BREAKER_ID: FnSetBreaker_O_UPSTREAM_BREAKER_ID,
		}
	}
	funcMap[FnMoveBreaker] = func() map[string]string {
		return map[string]string{
			baseFunc:                            FnMoveBreaker,
			FnMoveBreaker_O_ID:                  FnMoveBreaker_O_ID,
			FnMoveBreaker_O_NAME:                FnMoveBreaker_O_NAME,
			FnMoveBreaker_O_BREAKER_GROUP_ID:    FnMoveBreaker_O_BREAKER_GROUP_ID,
			FnMoveBreaker_O_SPACE_GROUP_ID:      FnMoveBreaker_O_SPACE_GROUP_ID,
			FnMoveBreaker_O_DISPLAY_ORDER:       FnMoveBreaker_O_DISPLAY_ORDER,
			FnMoveBreaker_O_UPSTREAM_BREAKER_ID: FnMoveBreaker_O_UPSTREAM_BREAKER_ID,
		}
	}
	funcMap[FnSetBreakerUpstream] = func() map[string]string {
		return map[string]string{
			baseFunc:                                   FnSetBreakerUpstream,
			FnSetBreakerUpstream_O_ID:                  FnSetBreakerUpstream_O_ID,
			FnSetBreakerUpstream_O_NAME:                FnSetBreakerUpstream_O_NAME,
			FnSetBreakerUpstream_O_BREAKER_GROUP_ID:    FnSetBreakerUpstream_O_BREAKER_GROUP_ID,
			FnSetBreakerUpstream_O_SPACE_GROUP_ID:      FnSetBreakerUpstream_O_SPACE_GROUP_ID,
			FnSetBreakerUpstream_O_DISPLAY_ORDER:       FnSetBreakerUpstream_O_DISPLAY_ORDER,
			FnSetBreakerUpstream_O_UPSTREAM_BREAKER_ID: FnSetBreakerUpstream_O_UPSTREAM_BREAKER_ID,
			baseErrSelfReference:                       FnSetBreakerUpstream_ERR_SELF_REFERENCE,
			baseErrCycleDetected:                       FnSetBreakerUpstream_ERR_CYCLE_DETECTED,
		}
	}
	funcMap[FnAddDeviceBreaker] = func() map[string]string {
		return map[string]string{
			baseFunc:                    FnAddDeviceBreaker,
			FnAddDeviceBreaker_O_RESULT: FnAddDeviceBreaker_O_RESULT,
		}
	}
}

func RenderFunctions(wr io.Writer) error {
	return render("functions", rawFunctions, wr)
}
