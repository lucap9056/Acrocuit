package postgres

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgDevices struct {
	pool *pgxpool.Pool
}

func (d *pgDevices) GetDevices(ctx context.Context, userEmail string, spaceId int, include string) ([]models.Device, error) {
	keys, fields, includePosition, err := models.ParseDeviceColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices
	}
	query += ` WHERE ` + schema.FullDevices_SPACE_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1`

	rows, err := d.pool.Query(ctx, query, userEmail, spaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]models.Device, 0)
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(device.Values(fields)...); err != nil {
			return nil, err
		}
		device.Build()
		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (d *pgDevices) GetDevice(ctx context.Context, userEmail string, deviceId int, include string) (*models.Device, error) {
	keys, fields, includePosition, err := models.ParseDeviceColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices
	}
	query += ` WHERE ` + schema.FullDevices_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1`

	var device models.Device
	err = d.pool.QueryRow(ctx, query, userEmail, deviceId).Scan(device.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	device.Build()

	return &device, nil
}

func (d *pgDevices) AddDevice(ctx context.Context, userEmail string, spaceId int, name string, position *models.Position) (*models.Device, error) {
	var posX, posY, posZ *int
	if position != nil {
		posX, posY, posZ = &position.X, &position.Y, &position.Z
	}

	query := `
	WITH check_space AS (
		SELECT 1
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		FOR UPDATE OF ` + schema.SpacesTable + `
	),
	inserted_device AS (
		INSERT INTO ` + schema.DevicesTable + ` (` + schema.Devices_NAME + `, ` + schema.Devices_SPACE_ID + `)
		SELECT $3, $1
		WHERE EXISTS (SELECT 1 FROM check_space)
		RETURNING ` + schema.Devices_ID + `
	),
	inserted_position AS (
		INSERT INTO ` + schema.DevicePositionsTable + ` (` + schema.DevicePositions_DEVICE_ID + `, ` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + `)
		SELECT ` + schema.Devices_ID + `, $4, $5, $6
		FROM inserted_device
		WHERE $4::int IS NOT NULL
		RETURNING ` + schema.DevicePositions_DEVICE_ID + `
	)
	SELECT ` + schema.Devices_ID + ` FROM inserted_device;
	`

	device := &models.Device{Name: name, SpaceId: spaceId, Position: position}

	err := d.pool.QueryRow(ctx, query, spaceId, userEmail, name, posX, posY, posZ).Scan(&device.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *pgDevices) SetDevice(ctx context.Context, userEmail string, deviceId int, name *string, position *models.Position) (*models.Device, error) {
	var posX, posY, posZ *int
	if position != nil {
		posX, posY, posZ = &position.X, &position.Y, &position.Z
	}

	query := `
		WITH permitted AS (
			SELECT ` + schema.FullDevices_ID + `
			FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		permitted_update AS (
			UPDATE ` + schema.DevicesTable + `
			SET ` + schema.Devices_NAME + ` = COALESCE($3, ` + schema.Devices_NAME + `)
			WHERE ` + schema.Devices_ID + ` = $1 AND EXISTS (SELECT 1 FROM permitted)
			RETURNING ` + schema.FullDevices_ID + `, ` + schema.FullDevices_NAME + `, ` + schema.FullDevices_SPACE_ID + `
		),
		current_position AS (
			SELECT p.` + schema.DevicePositions_POSITION_X + `, p.` + schema.DevicePositions_POSITION_Y + `, p.` + schema.DevicePositions_POSITION_Z + `
			FROM ` + schema.DevicePositionsTable + ` p
			JOIN permitted_update ON permitted_update.id = p.` + schema.DevicePositions_DEVICE_ID + `
		),
		upserted_position AS (
			INSERT INTO ` + schema.DevicePositionsTable + ` (` + schema.DevicePositions_DEVICE_ID + `, ` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + `)
			SELECT id, $4, $5, $6
			FROM permitted_update
			WHERE $4::integer IS NOT NULL
			ON CONFLICT (` + schema.DevicePositions_DEVICE_ID + `) DO UPDATE
			SET ` + schema.DevicePositions_POSITION_X + ` = EXCLUDED.` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + ` = EXCLUDED.` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + ` = EXCLUDED.` + schema.DevicePositions_POSITION_Z + `
			RETURNING ` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + `
		)
		SELECT
			permitted_update.` + schema.Devices_ID + `,
			permitted_update.` + schema.Devices_NAME + `,
			permitted_update.` + schema.Devices_SPACE_ID + `,
			COALESCE(upserted_position.` + schema.DevicePositions_POSITION_X + `, current_position.` + schema.DevicePositions_POSITION_X + `),
			COALESCE(upserted_position.` + schema.DevicePositions_POSITION_Y + `, current_position.` + schema.DevicePositions_POSITION_Y + `),
			COALESCE(upserted_position.` + schema.DevicePositions_POSITION_Z + `, current_position.` + schema.DevicePositions_POSITION_Z + `)
		FROM permitted_update
		LEFT JOIN current_position ON true
		LEFT JOIN upserted_position ON true
	`

	device := &models.Device{Id: deviceId}

	err := d.pool.QueryRow(ctx, query, deviceId, userEmail, name, posX, posY, posZ).
		Scan(&device.Id, &device.Name, &device.SpaceId, &device.PositionX, &device.PositionY, &device.PositionZ)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *pgDevices) SetDevicePosition(ctx context.Context, userEmail string, deviceId int, position models.Position) (*models.Device, error) {
	return d.SetDevice(ctx, userEmail, deviceId, nil, &position)
}

func (d *pgDevices) DelDevice(ctx context.Context, userEmail string, deviceId int) error {
	query := `
		DELETE FROM ` + schema.DevicesTable + `
		WHERE ` + schema.FullDevices_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.JoinDevices_Spaces + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	t, err := d.pool.Exec(ctx, query, deviceId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (d *pgDevices) GetDeviceSwitches(ctx context.Context, userEmail string, deviceId int, include string) ([]models.Switch, error) {
	existsQuery := `
		SELECT EXISTS (
			SELECT 1 FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	var exists bool
	err := d.pool.QueryRow(ctx, existsQuery, deviceId, userEmail).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}

	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.DeviceSwitchesTable + `
		JOIN ` + schema.SwitchesTable + ` ON ` + schema.JoinDeviceSwitches_Switches + `
		WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $1`

	rows, err := d.pool.Query(ctx, query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switches := make([]models.Switch, 0)
	for rows.Next() {
		var sw models.Switch
		if err := rows.Scan(sw.Values(fields)...); err != nil {
			return nil, err
		}
		sw.Build()
		switches = append(switches, sw)
	}

	return switches, rows.Err()
}

func (d *pgDevices) AddDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId int) error {
	var result int
	query := `SELECT ` + schema.FnAddDeviceSwitch_O_RESULT + ` FROM ` + schema.FnAddDeviceSwitch + `($1, $2, $3)`
	err := d.pool.QueryRow(ctx, query, userEmail, deviceId, switchId).Scan(&result)
	if err != nil {
		return err
	}

	switch result {
	case 0:
		return nil
	case 3:
		return repository.ErrConflict
	default:
		return repository.ErrNotFound
	}
}

func (d *pgDevices) DelDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId int) error {
	query := `
		DELETE FROM ` + schema.DeviceSwitchesTable + `
		USING ` + schema.SwitchesTable + `
		JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.JoinSwitches_SwitchPanels + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SwitchPanels + `
		WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $1 AND ` + schema.FullDeviceSwitches_SWITCH_ID + ` = $2
			AND ` + schema.JoinDeviceSwitches_Switches + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $3
	`
	t, err := d.pool.Exec(ctx, query, deviceId, switchId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (d *pgDevices) GetDeviceUpstream(ctx context.Context, userEmail string, deviceId int) ([]models.Switch, error) {
	existsQuery := `
		SELECT EXISTS (
			SELECT 1 FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	var exists bool
	err := d.pool.QueryRow(ctx, existsQuery, deviceId, userEmail).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}

	query := `
		WITH RECURSIVE switch_chain AS (
			SELECT ` + schema.FullSwitches_ID + `, ` + schema.FullSwitches_NAME + `, ` + schema.FullSwitches_SWITCH_PANEL_ID + `, ` + schema.FullSwitches_SPACE_GROUP_ID + `, ` + schema.FullSwitches_DISPLAY_ORDER + `, ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + `
			FROM ` + schema.DeviceSwitchesTable + `
			JOIN ` + schema.SwitchesTable + ` ON ` + schema.JoinDeviceSwitches_Switches + `
			WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $1

			UNION ALL

			SELECT ` + schema.FullSwitches_ID + `, ` + schema.FullSwitches_NAME + `, ` + schema.FullSwitches_SWITCH_PANEL_ID + `, ` + schema.FullSwitches_SPACE_GROUP_ID + `, ` + schema.FullSwitches_DISPLAY_ORDER + `, ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + `
			FROM ` + schema.SwitchesTable + `
			JOIN switch_chain ON ` + schema.FullSwitches_ID + ` = switch_chain.` + schema.Switches_UPSTREAM_SWITCH_ID + `
		)
		SELECT DISTINCT ` + switchChainWithPanelColumns + switchChainJoinPanel + `
		ORDER BY switch_chain.` + schema.Switches_ID + `
	`
	rows, err := d.pool.Query(ctx, query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switches := make([]models.Switch, 0)
	for rows.Next() {
		var sw models.Switch
		if err := rows.Scan(sw.ValuesWithPanel()...); err != nil {
			return nil, err
		}
		sw.Build()
		switches = append(switches, sw)
	}

	return switches, rows.Err()
}
