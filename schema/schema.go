package schema

import (
	_ "embed"
	"io"
	"text/template"
)

//go:embed "schema.sql"
var rawSchema string

const (
	baseTable          = "table"
	baseId             = "id"
	baseName           = "name"
	baseSpaceGroupID   = "space_group_id"
	baseSpaceID        = "space_id"
	baseBreakerGroupID = "breaker_group_id"
	baseDisplayOrder   = "display_order"
	basePositionX      = "position_x"
	basePositionY      = "position_y"
	basePositionZ      = "position_z"
	baseDeviceID       = "device_id"
	baseUserID         = "user_id"

	SpaceGroupsTable = "space_groups"
	SpaceGroups_ID   = baseId
	SpaceGroups_NAME = baseName

	SpacesTable                        = "spaces"
	Spaces_ID                          = baseId
	Spaces_NAME                        = baseName
	Spaces_SPACE_GROUP_ID              = baseSpaceGroupID
	Spaces_DISPLAY_ORDER               = baseDisplayOrder
	Spaces_BACKGROUND_IMAGE_UPDATED_AT = "background_image_updated_at"

	BreakerGroupsTable           = "breaker_groups"
	BreakerGroups_ID             = baseId
	BreakerGroups_NAME           = baseName
	BreakerGroups_SPACE_ID       = baseSpaceID
	BreakerGroups_SPACE_GROUP_ID = baseSpaceGroupID

	BreakerGroupPositionsTable             = "breaker_group_positions"
	BreakerGroupPositions_BREAKER_GROUP_ID = baseBreakerGroupID
	BreakerGroupPositions_POSITION_X       = basePositionX
	BreakerGroupPositions_POSITION_Y       = basePositionY
	BreakerGroupPositions_POSITION_Z       = basePositionZ

	BreakersTable                = "breakers"
	Breakers_ID                  = baseId
	Breakers_NAME                = baseName
	Breakers_BREAKER_GROUP_ID    = baseBreakerGroupID
	Breakers_SPACE_GROUP_ID      = baseSpaceGroupID
	Breakers_DISPLAY_ORDER       = baseDisplayOrder
	Breakers_UPSTREAM_BREAKER_ID = "upstream_breaker_id"

	DevicesTable     = "devices"
	Devices_ID       = baseId
	Devices_NAME     = baseName
	Devices_SPACE_ID = baseSpaceID

	DevicePositionsTable       = "device_positions"
	DevicePositions_DEVICE_ID  = baseDeviceID
	DevicePositions_POSITION_X = basePositionX
	DevicePositions_POSITION_Y = basePositionY
	DevicePositions_POSITION_Z = basePositionZ

	DeviceBreakersTable       = "device_breakers"
	DeviceBreakers_DEVICE_ID  = baseDeviceID
	DeviceBreakers_BREAKER_ID = "breaker_id"

	UsersTable     = "users"
	Users_ID       = baseId
	Users_USERNAME = "username"

	UserSecretsTable          = "user_secrets"
	UserSecrets_USER_ID       = baseUserID
	UserSecrets_PASSWORD_SALT = "password_salt"
	UserSecrets_PASSWORD_HASH = "password_hash"

	SpaceGroupOwnersTable           = "space_group_owners"
	SpaceGroupOwners_USER_ID        = baseUserID
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

	FullBreakerGroups_ID             = BreakerGroupsTable + "." + BreakerGroups_ID
	FullBreakerGroups_NAME           = BreakerGroupsTable + "." + BreakerGroups_NAME
	FullBreakerGroups_SPACE_ID       = BreakerGroupsTable + "." + BreakerGroups_SPACE_ID
	FullBreakerGroups_SPACE_GROUP_ID = BreakerGroupsTable + "." + BreakerGroups_SPACE_GROUP_ID

	FullBreakerGroupPositions_BREAKER_GROUP_ID = BreakerGroupPositionsTable + "." + BreakerGroupPositions_BREAKER_GROUP_ID
	FullBreakerGroupPositions_POSITION_X       = BreakerGroupPositionsTable + "." + BreakerGroupPositions_POSITION_X
	FullBreakerGroupPositions_POSITION_Y       = BreakerGroupPositionsTable + "." + BreakerGroupPositions_POSITION_Y
	FullBreakerGroupPositions_POSITION_Z       = BreakerGroupPositionsTable + "." + BreakerGroupPositions_POSITION_Z

	FullBreakers_ID                  = BreakersTable + "." + Breakers_ID
	FullBreakers_NAME                = BreakersTable + "." + Breakers_NAME
	FullBreakers_BREAKER_GROUP_ID    = BreakersTable + "." + Breakers_BREAKER_GROUP_ID
	FullBreakers_SPACE_GROUP_ID      = BreakersTable + "." + Breakers_SPACE_GROUP_ID
	FullBreakers_DISPLAY_ORDER       = BreakersTable + "." + Breakers_DISPLAY_ORDER
	FullBreakers_UPSTREAM_BREAKER_ID = BreakersTable + "." + Breakers_UPSTREAM_BREAKER_ID

	FullDevices_ID       = DevicesTable + "." + Devices_ID
	FullDevices_NAME     = DevicesTable + "." + Devices_NAME
	FullDevices_SPACE_ID = DevicesTable + "." + Devices_SPACE_ID

	FullDevicePositions_DEVICE_ID  = DevicePositionsTable + "." + DevicePositions_DEVICE_ID
	FullDevicePositions_POSITION_X = DevicePositionsTable + "." + DevicePositions_POSITION_X
	FullDevicePositions_POSITION_Y = DevicePositionsTable + "." + DevicePositions_POSITION_Y
	FullDevicePositions_POSITION_Z = DevicePositionsTable + "." + DevicePositions_POSITION_Z

	FullDeviceBreakers_DEVICE_ID  = DeviceBreakersTable + "." + DeviceBreakers_DEVICE_ID
	FullDeviceBreakers_BREAKER_ID = DeviceBreakersTable + "." + DeviceBreakers_BREAKER_ID

	FullUsers_ID       = UsersTable + "." + Users_ID
	FullUsers_USERNAME = UsersTable + "." + Users_USERNAME

	FullUserSecrets_USER_ID       = UserSecretsTable + "." + UserSecrets_USER_ID
	FullUserSecrets_PASSWORD_SALT = UserSecretsTable + "." + UserSecrets_PASSWORD_SALT
	FullUserSecrets_PASSWORD_HASH = UserSecretsTable + "." + UserSecrets_PASSWORD_HASH

	FullSpaceGroupOwners_USER_ID        = SpaceGroupOwnersTable + "." + SpaceGroupOwners_USER_ID
	FullSpaceGroupOwners_SPACE_GROUP_ID = SpaceGroupOwnersTable + "." + SpaceGroupOwners_SPACE_GROUP_ID
)

const (
	JoinBreakerGroups_Spaces                = FullBreakerGroups_SPACE_ID + " = " + FullSpaces_ID
	JoinSpaceGroupOwners_Spaces             = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSpaces_SPACE_GROUP_ID
	JoinBreakerGroupPositions_BreakerGroups = FullBreakerGroupPositions_BREAKER_GROUP_ID + " = " + FullBreakerGroups_ID
	JoinBreakers_BreakerGroups              = FullBreakers_BREAKER_GROUP_ID + " = " + FullBreakerGroups_ID
	JoinSpaceGroupOwners_BreakerGroups      = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullBreakerGroups_SPACE_GROUP_ID
	JoinSpaceGroupOwners_Breakers           = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullBreakers_SPACE_GROUP_ID
	JoinDevices_Spaces                      = FullDevices_SPACE_ID + " = " + FullSpaces_ID
	JoinDevicePositions_Devices             = FullDevicePositions_DEVICE_ID + " = " + FullDevices_ID
	JoinDeviceBreakers_Devices              = FullDeviceBreakers_DEVICE_ID + " = " + FullDevices_ID
	JoinDeviceBreakers_Breakers             = FullDeviceBreakers_BREAKER_ID + " = " + FullBreakers_ID
	JoinSpaceGroupOwners_SpaceGroups        = FullSpaceGroupOwners_SPACE_GROUP_ID + " = " + FullSpaceGroups_ID
	JoinUserSecrets_Users                   = FullUserSecrets_USER_ID + "=" + FullUsers_ID
)

var funcMap = template.FuncMap{
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
	BreakerGroupsTable: func() map[string]string {
		return map[string]string{
			baseTable:                    BreakerGroupsTable,
			BreakerGroups_ID:             BreakerGroups_ID,
			BreakerGroups_NAME:           BreakerGroups_NAME,
			BreakerGroups_SPACE_ID:       BreakerGroups_SPACE_ID,
			BreakerGroups_SPACE_GROUP_ID: BreakerGroups_SPACE_GROUP_ID,
		}
	},
	BreakerGroupPositionsTable: func() map[string]string {
		return map[string]string{
			baseTable:                              BreakerGroupPositionsTable,
			BreakerGroupPositions_BREAKER_GROUP_ID: BreakerGroupPositions_BREAKER_GROUP_ID,
			BreakerGroupPositions_POSITION_X:       BreakerGroupPositions_POSITION_X,
			BreakerGroupPositions_POSITION_Y:       BreakerGroupPositions_POSITION_Y,
			BreakerGroupPositions_POSITION_Z:       BreakerGroupPositions_POSITION_Z,
		}
	},
	BreakersTable: func() map[string]string {
		return map[string]string{
			baseTable:                    BreakersTable,
			Breakers_ID:                  Breakers_ID,
			Breakers_NAME:                Breakers_NAME,
			Breakers_BREAKER_GROUP_ID:    Breakers_BREAKER_GROUP_ID,
			Breakers_SPACE_GROUP_ID:      Breakers_SPACE_GROUP_ID,
			Breakers_DISPLAY_ORDER:       Breakers_DISPLAY_ORDER,
			Breakers_UPSTREAM_BREAKER_ID: Breakers_UPSTREAM_BREAKER_ID,
		}
	},
	DevicesTable: func() map[string]string {
		return map[string]string{
			baseTable:        DevicesTable,
			Devices_ID:       Devices_ID,
			Devices_NAME:     Devices_NAME,
			Devices_SPACE_ID: Devices_SPACE_ID,
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
	DeviceBreakersTable: func() map[string]string {
		return map[string]string{
			baseTable:                 DeviceBreakersTable,
			DeviceBreakers_DEVICE_ID:  DeviceBreakers_DEVICE_ID,
			DeviceBreakers_BREAKER_ID: DeviceBreakers_BREAKER_ID,
		}
	},
	UsersTable: func() map[string]string {
		return map[string]string{
			baseTable:      UsersTable,
			Users_ID:       Users_ID,
			Users_USERNAME: Users_USERNAME,
		}
	},
	UserSecretsTable: func() map[string]string {
		return map[string]string{
			baseTable:                 UserSecretsTable,
			UserSecrets_USER_ID:       UserSecrets_USER_ID,
			UserSecrets_PASSWORD_SALT: UserSecrets_PASSWORD_SALT,
			UserSecrets_PASSWORD_HASH: UserSecrets_PASSWORD_HASH,
		}
	},
	SpaceGroupOwnersTable: func() map[string]string {
		return map[string]string{
			baseTable:                       SpaceGroupOwnersTable,
			SpaceGroupOwners_USER_ID:        SpaceGroupOwners_USER_ID,
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

func Render(wr io.Writer) error {
	return render("schema", rawSchema, wr)
}
