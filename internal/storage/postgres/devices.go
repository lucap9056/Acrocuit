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

func (d *pgDevices) GetDevices(ctx context.Context, userEmail string, spaceId models.SpaceId, include string) ([]models.Device, error) {
	keys, fields, includePosition, err := models.ParseDeviceColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices
	}
	query += ` WHERE ` + schema.FullDevices_SPACE_ID + ` = $1
		  AND EXISTS (
			  SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			  WHERE ` + schema.JoinSpaceGroupACLs_Spaces + `
			    AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		  )`

	rows, err := d.pool.Query(ctx, query, spaceId, userEmail, models.PublicAccessEmail)
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

func (d *pgDevices) GetDevice(ctx context.Context, userEmail string, deviceId models.DeviceId, include string) (*models.Device, error) {
	keys, fields, includePosition, err := models.ParseDeviceColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices
	}
	query += ` WHERE ` + schema.FullDevices_ID + ` = $1
		  AND EXISTS (
			  SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			  WHERE ` + schema.JoinSpaceGroupACLs_Spaces + `
			    AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		  )`

	var device models.Device
	err = d.pool.QueryRow(ctx, query, deviceId, userEmail, models.PublicAccessEmail).Scan(device.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	device.Build()

	return &device, nil
}

func (d *pgDevices) AddDevice(ctx context.Context, userEmail string, spaceId models.SpaceId, name string, position *models.Position, attributes models.Attributes) (*models.Device, error) {
	var posX, posY, posZ *int
	if position != nil {
		posX, posY, posZ = &position.X, &position.Y, &position.Z
	}

	query := `
	WITH check_space AS (
		SELECT 1
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $1 AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
		FOR UPDATE OF ` + schema.SpacesTable + `
	),
	inserted_device AS (
		INSERT INTO ` + schema.DevicesTable + ` (` + schema.Devices_NAME + `, ` + schema.Devices_SPACE_ID + `, ` + schema.Devices_ATTRIBUTES + `)
		SELECT $3, $1, $7::json
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

	device := &models.Device{Name: name, SpaceId: spaceId, Position: position, Attributes: attributes}

	err := d.pool.QueryRow(ctx, query, spaceId, userEmail, name, posX, posY, posZ, attributes).Scan(&device.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *pgDevices) SetDevice(ctx context.Context, userEmail string, deviceId models.DeviceId, name *string, position *models.Position, attributes models.Attributes) (*models.Device, error) {
	var posX, posY, posZ *int
	if position != nil {
		posX, posY, posZ = &position.X, &position.Y, &position.Z
	}

	query := `
		WITH permitted AS (
			SELECT ` + schema.FullDevices_ID + `
			FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1 AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		permitted_update AS (
			UPDATE ` + schema.DevicesTable + `
			SET ` + schema.Devices_NAME + ` = COALESCE($3, ` + schema.Devices_NAME + `),
				` + schema.Devices_ATTRIBUTES + ` = COALESCE($7::json, ` + schema.Devices_ATTRIBUTES + `)
			WHERE ` + schema.Devices_ID + ` = $1 AND EXISTS (SELECT 1 FROM permitted)
			RETURNING ` + models.DEVICE_ALL_COLUMNS + `
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
			permitted_update.` + schema.Devices_ATTRIBUTES + `,
			COALESCE(upserted_position.` + schema.DevicePositions_POSITION_X + `, current_position.` + schema.DevicePositions_POSITION_X + `),
			COALESCE(upserted_position.` + schema.DevicePositions_POSITION_Y + `, current_position.` + schema.DevicePositions_POSITION_Y + `),
			COALESCE(upserted_position.` + schema.DevicePositions_POSITION_Z + `, current_position.` + schema.DevicePositions_POSITION_Z + `)
		FROM permitted_update
		LEFT JOIN current_position ON true
		LEFT JOIN upserted_position ON true
	`

	device := &models.Device{Id: deviceId}

	err := d.pool.QueryRow(ctx, query, deviceId, userEmail, name, posX, posY, posZ, attributes).Scan(device.ValuesWithPosition()...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *pgDevices) SetDevicePosition(ctx context.Context, userEmail string, deviceId models.DeviceId, position models.Position) (*models.Device, error) {
	return d.SetDevice(ctx, userEmail, deviceId, nil, &position, nil)
}

func (d *pgDevices) DelDevice(ctx context.Context, userEmail string, deviceId models.DeviceId) error {
	query := `
		DELETE FROM ` + schema.DevicesTable + `
		WHERE ` + schema.FullDevices_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_Spaces + `
			WHERE ` + schema.JoinDevices_Spaces + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
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

func (d *pgDevices) GetDeviceSwitches(ctx context.Context, userEmail string, deviceId models.DeviceId, include string) ([]models.LinkedSwitch, error) {
	existsQuery := `
		SELECT EXISTS (
			SELECT 1 FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1 AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		)
	`
	var exists bool
	err := d.pool.QueryRow(ctx, existsQuery, deviceId, userEmail, models.PublicAccessEmail).Scan(&exists)
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

	query := `SELECT ` + keys + `,` + schema.FullDeviceSwitches_ATTRIBUTES + `
		FROM ` + schema.DeviceSwitchesTable + `
		JOIN ` + schema.SwitchesTable + ` ON ` + schema.JoinDeviceSwitches_Switches + `
		WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $1`

	rows, err := d.pool.Query(ctx, query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	linkedSwitches := make([]models.LinkedSwitch, 0)
	for rows.Next() {
		var linked models.LinkedSwitch
		if err := rows.Scan(linked.Values(fields)...); err != nil {
			return nil, err
		}
		linked.Build()
		linkedSwitches = append(linkedSwitches, linked)
	}

	return linkedSwitches, rows.Err()
}

func (d *pgDevices) AddDeviceSwitch(ctx context.Context, userEmail string, deviceId models.DeviceId, switchId models.SwitchId, attributes models.Attributes) (*models.DeviceLink, error) {
	var result int
	query := `SELECT ` + schema.FnAddDeviceSwitch_O_RESULT + ` FROM ` + schema.FnAddDeviceSwitch + `($1, $2, $3, $4)`
	err := d.pool.QueryRow(ctx, query, userEmail, deviceId, switchId, attributes).Scan(&result)
	if err != nil {
		return nil, err
	}

	switch result {
	case 0:
		return &models.DeviceLink{DeviceId: deviceId, SwitchId: switchId, Attributes: attributes}, nil
	case 3:
		return nil, repository.ErrConflict
	default:
		return nil, repository.ErrNotFound
	}
}

func (d *pgDevices) SetDeviceSwitch(ctx context.Context, userEmail string, deviceId models.DeviceId, switchId models.SwitchId, attributes models.Attributes) (*models.DeviceLink, error) {
	query := `
		UPDATE ` + schema.DeviceSwitchesTable + `
		SET ` + schema.DeviceSwitches_ATTRIBUTES + ` = $1
		FROM ` + schema.SwitchesTable + `
		JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_Switches + `
		WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $2 AND ` + schema.FullDeviceSwitches_SWITCH_ID + ` = $3
			AND ` + schema.JoinDeviceSwitches_Switches + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $4 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
		RETURNING ` + models.DEVICE_LINK_COLUMNS

	var link models.DeviceLink
	err := d.pool.QueryRow(ctx, query, attributes, deviceId, switchId, userEmail).Scan(link.Values()...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (d *pgDevices) DelDeviceSwitch(ctx context.Context, userEmail string, deviceId models.DeviceId, switchId models.SwitchId) error {
	query := `
		DELETE FROM ` + schema.DeviceSwitchesTable + `
		USING ` + schema.SwitchesTable + `
		JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.JoinSwitches_SwitchPanels + `
		JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_SwitchPanels + `
		WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $1 AND ` + schema.FullDeviceSwitches_SWITCH_ID + ` = $2
			AND ` + schema.JoinDeviceSwitches_Switches + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
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

func (d *pgDevices) GetDeviceUpstream(ctx context.Context, userEmail string, deviceId models.DeviceId) (*models.PowerChainGraph, error) {
	deviceRows, err := d.pool.Query(ctx, `
		SELECT `+models.DEVICE_WITH_POSITION_COLUMNS+`
		FROM `+schema.DevicesTable+`
		JOIN `+schema.SpacesTable+` ON `+schema.JoinDevices_Spaces+`
		LEFT JOIN `+schema.DevicePositionsTable+` ON `+schema.JoinDevicePositions_Devices+`
		WHERE `+schema.FullDevices_ID+` = $1
		  AND EXISTS (
			  SELECT 1 FROM `+schema.SpaceGroupACLsTable+`
			  WHERE `+schema.JoinSpaceGroupACLs_Spaces+`
			    AND `+schema.FullSpaceGroupACLs_USER_EMAIL+` IN ($2, $3)
		  )
	`, deviceId, userEmail, models.PublicAccessEmail)
	if err != nil {
		return nil, err
	}
	defer deviceRows.Close()

	chain := &models.PowerChainGraph{
		Devices:     make([]models.Device, 0),
		Switches:    make([]models.ChainSwitch, 0),
		SwitchLinks: make([]models.SwitchLink, 0),
		DeviceLinks: make([]models.DeviceLink, 0),
	}
	for deviceRows.Next() {
		var device models.Device
		if err := deviceRows.Scan(device.ValuesWithPosition()...); err != nil {
			return nil, err
		}
		device.Build()
		chain.Devices = append(chain.Devices, device)
	}
	if err := deviceRows.Err(); err != nil {
		return nil, err
	}
	if len(chain.Devices) == 0 {
		return nil, repository.ErrNotFound
	}

	switchRows, err := d.pool.Query(ctx, `
		WITH RECURSIVE switch_chain(id) AS (
			SELECT `+schema.FullDeviceSwitches_SWITCH_ID+`
			FROM `+schema.DeviceSwitchesTable+`
			WHERE `+schema.FullDeviceSwitches_DEVICE_ID+` = $1

			UNION

			SELECT `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+`
			FROM `+schema.SwitchLinksTable+`
			JOIN switch_chain ON `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+` = switch_chain.id
		)
		SELECT `+models.SWITCH_ALL_COLUMNS+`, `+models.SWITCH_PANEL_WITH_POSITION_COLUMNS+`
		FROM switch_chain
		JOIN `+schema.SwitchesTable+` ON `+schema.FullSwitches_ID+` = switch_chain.id
		JOIN `+schema.SwitchPanelsTable+` ON `+schema.JoinSwitches_SwitchPanels+`
		LEFT JOIN `+schema.SwitchPanelPositionsTable+` ON `+schema.JoinSwitchPanelPositions_SwitchPanels+`
		ORDER BY `+schema.FullSwitches_ID+`
	`, deviceId)
	if err != nil {
		return nil, err
	}
	defer switchRows.Close()

	switchIds := make([]models.SwitchId, 0)
	for switchRows.Next() {
		var sw models.ChainSwitch
		if err := switchRows.Scan(sw.ValuesWithPanel()...); err != nil {
			return nil, err
		}
		sw.Build()
		chain.Switches = append(chain.Switches, sw)
		switchIds = append(switchIds, sw.Id)
	}
	if err := switchRows.Err(); err != nil {
		return nil, err
	}

	switchLinkRows, err := d.pool.Query(ctx, `
		SELECT `+models.SWITCH_LINK_COLUMNS+`
		FROM `+schema.SwitchLinksTable+`
		WHERE `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+` = ANY($1::text[]::uuid[])
		ORDER BY `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+`, `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+`
	`, switchIds)
	if err != nil {
		return nil, err
	}
	defer switchLinkRows.Close()

	for switchLinkRows.Next() {
		var link models.SwitchLink
		if err := switchLinkRows.Scan(link.Values()...); err != nil {
			return nil, err
		}
		chain.SwitchLinks = append(chain.SwitchLinks, link)
	}
	if err := switchLinkRows.Err(); err != nil {
		return nil, err
	}

	deviceLinkRows, err := d.pool.Query(ctx, `
		SELECT `+models.DEVICE_LINK_COLUMNS+`
		FROM `+schema.DeviceSwitchesTable+`
		WHERE `+schema.FullDeviceSwitches_DEVICE_ID+` = $1
		ORDER BY `+schema.FullDeviceSwitches_SWITCH_ID+`
	`, deviceId)
	if err != nil {
		return nil, err
	}
	defer deviceLinkRows.Close()

	for deviceLinkRows.Next() {
		var link models.DeviceLink
		if err := deviceLinkRows.Scan(link.Values()...); err != nil {
			return nil, err
		}
		chain.DeviceLinks = append(chain.DeviceLinks, link)
	}

	return chain, deviceLinkRows.Err()
}
