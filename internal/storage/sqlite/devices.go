package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"encoding/json"
)

type sqliteDevices struct {
	db *sql.DB
}

func (d *sqliteDevices) GetDevices(ctx context.Context, _ string, spaceId string, include string) ([]models.Device, error) {
	keys, fields, includePosition, err := models.ParseDeviceColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.DevicesTable
	if includePosition {
		query += ` LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices
	}
	query += ` WHERE ` + schema.FullDevices_SPACE_ID + ` = $1`

	rows, err := d.db.QueryContext(ctx, query, spaceId)
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

func (d *sqliteDevices) GetDevice(ctx context.Context, _ string, deviceId string, include string) (*models.Device, error) {
	keys, fields, includePosition, err := models.ParseDeviceColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.DevicesTable
	if includePosition {
		query += ` LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices
	}
	query += ` WHERE ` + schema.FullDevices_ID + ` = $1`

	var device models.Device
	if err := d.db.QueryRowContext(ctx, query, deviceId).Scan(device.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}
	device.Build()

	return &device, nil
}

func (d *sqliteDevices) AddDevice(ctx context.Context, _ string, spaceId string, name string, position *models.Position, attributes models.Attributes) (*models.Device, error) {
	device := &models.Device{Name: name, SpaceId: spaceId, Position: position, Attributes: attributes}

	err := withTx(ctx, d.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SpacesTable, schema.Spaces_ID, spaceId); err != nil {
			return err
		}

		insert := `INSERT INTO ` + schema.DevicesTable + ` (` + schema.Devices_NAME + `, ` + schema.Devices_SPACE_ID + `, ` + schema.Devices_ATTRIBUTES + `) VALUES ($1, $2, $3) RETURNING ` + schema.Devices_ID
		if err := tx.QueryRowContext(ctx, insert, name, spaceId, attributes).Scan(&device.Id); err != nil {
			return err
		}

		if position == nil {
			return nil
		}
		return upsertDevicePosition(ctx, tx, device.Id, *position)
	})
	if err != nil {
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *sqliteDevices) SetDevice(ctx context.Context, _ string, deviceId string, name *string, position *models.Position, attributes models.Attributes) (*models.Device, error) {
	device := &models.Device{}

	err := withTx(ctx, d.db, func(tx *sql.Tx) error {
		update := `
			UPDATE ` + schema.DevicesTable + `
			SET ` + schema.Devices_NAME + ` = COALESCE($1, ` + schema.Devices_NAME + `),
				` + schema.Devices_ATTRIBUTES + ` = COALESCE($2, ` + schema.Devices_ATTRIBUTES + `)
			WHERE ` + schema.Devices_ID + ` = $3
		`
		if err := notFoundIfNoneAffected(tx.ExecContext(ctx, update, name, attributes, deviceId)); err != nil {
			return err
		}

		if position != nil {
			if err := upsertDevicePosition(ctx, tx, deviceId, *position); err != nil {
				return err
			}
		}

		result := `
			SELECT ` + models.DEVICE_WITH_POSITION_COLUMNS + `
			FROM ` + schema.DevicesTable + `
			LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices + `
			WHERE ` + schema.FullDevices_ID + ` = $1
		`
		return tx.QueryRowContext(ctx, result, deviceId).Scan(device.ValuesWithPosition()...)
	})
	if err != nil {
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *sqliteDevices) SetDevicePosition(ctx context.Context, userEmail string, deviceId string, position models.Position) (*models.Device, error) {
	return d.SetDevice(ctx, userEmail, deviceId, nil, &position, nil)
}

func (d *sqliteDevices) DelDevice(ctx context.Context, _ string, deviceId string) error {
	query := `DELETE FROM ` + schema.DevicesTable + ` WHERE ` + schema.Devices_ID + ` = $1`
	return notFoundIfNoneAffected(d.db.ExecContext(ctx, query, deviceId))
}

func (d *sqliteDevices) GetDeviceSwitches(ctx context.Context, _ string, deviceId string, include string) ([]models.LinkedSwitch, error) {
	if err := requireRow(ctx, d.db, schema.DevicesTable, schema.Devices_ID, deviceId); err != nil {
		return nil, err
	}

	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `,` + schema.FullDeviceSwitches_ATTRIBUTES + `
		FROM ` + schema.DeviceSwitchesTable + `
		JOIN ` + schema.SwitchesTable + ` ON ` + schema.JoinDeviceSwitches_Switches + `
		WHERE ` + schema.FullDeviceSwitches_DEVICE_ID + ` = $1`

	rows, err := d.db.QueryContext(ctx, query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	linkedSwitches := make([]models.LinkedSwitch, 0)
	for rows.Next() {
		var ls models.LinkedSwitch
		if err := rows.Scan(ls.Values(fields)...); err != nil {
			return nil, err
		}
		ls.Build()
		linkedSwitches = append(linkedSwitches, ls)
	}

	return linkedSwitches, rows.Err()
}

func (d *sqliteDevices) AddDeviceSwitch(ctx context.Context, _ string, deviceId, switchId string, attributes models.Attributes) (*models.DeviceLink, error) {
	link := &models.DeviceLink{DeviceId: deviceId, SwitchId: switchId, Attributes: attributes}

	err := withTx(ctx, d.db, func(tx *sql.Tx) error {
		var spaceGroupId string
		spaceGroup := `
			SELECT ` + schema.FullSpaces_SPACE_GROUP_ID + `
			FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1
		`
		if err := tx.QueryRowContext(ctx, spaceGroup, deviceId).Scan(&spaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if err := requireSwitchInSpaceGroup(ctx, tx, switchId, spaceGroupId); err != nil {
			return err
		}

		insert := `INSERT INTO ` + schema.DeviceSwitchesTable + ` (` + schema.DeviceSwitches_DEVICE_ID + `, ` + schema.DeviceSwitches_SWITCH_ID + `, ` + schema.DeviceSwitches_ATTRIBUTES + `) VALUES ($1, $2, $3)`
		_, err := tx.ExecContext(ctx, insert, deviceId, switchId, attributes)
		if isUniqueViolation(err) {
			return repository.ErrConflict
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (d *sqliteDevices) SetDeviceSwitch(ctx context.Context, _ string, deviceId, switchId string, attributes models.Attributes) (*models.DeviceLink, error) {
	query := `
		UPDATE ` + schema.DeviceSwitchesTable + `
		SET ` + schema.DeviceSwitches_ATTRIBUTES + ` = $1
		WHERE ` + schema.DeviceSwitches_DEVICE_ID + ` = $2 AND ` + schema.DeviceSwitches_SWITCH_ID + ` = $3
		RETURNING ` + models.DEVICE_LINK_COLUMNS

	var link models.DeviceLink
	if err := d.db.QueryRowContext(ctx, query, attributes, deviceId, switchId).Scan(link.Values()...); err != nil {
		return nil, notFoundIfNoRows(err)
	}

	return &link, nil
}

func (d *sqliteDevices) DelDeviceSwitch(ctx context.Context, _ string, deviceId, switchId string) error {
	query := `
		DELETE FROM ` + schema.DeviceSwitchesTable + `
		WHERE ` + schema.DeviceSwitches_DEVICE_ID + ` = $1 AND ` + schema.DeviceSwitches_SWITCH_ID + ` = $2
	`
	return notFoundIfNoneAffected(d.db.ExecContext(ctx, query, deviceId, switchId))
}

func (d *sqliteDevices) GetDeviceUpstream(ctx context.Context, _ string, deviceId string) (*models.PowerChain, error) {
	deviceRows, err := d.db.QueryContext(ctx, `
		SELECT `+models.DEVICE_WITH_POSITION_COLUMNS+`
		FROM `+schema.DevicesTable+`
		LEFT JOIN `+schema.DevicePositionsTable+` ON `+schema.JoinDevicePositions_Devices+`
		WHERE `+schema.FullDevices_ID+` = $1
	`, deviceId)
	if err != nil {
		return nil, err
	}
	defer deviceRows.Close()

	chain := &models.PowerChain{
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

	switchRows, err := d.db.QueryContext(ctx, `
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

	switchIds := make([]string, 0)
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

	switchIdsJSON, err := json.Marshal(switchIds)
	if err != nil {
		return nil, err
	}

	switchLinkRows, err := d.db.QueryContext(ctx, `
		SELECT `+models.SWITCH_LINK_COLUMNS+`
		FROM `+schema.SwitchLinksTable+`
		WHERE `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+` IN (SELECT value FROM json_each($1))
		ORDER BY `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+`, `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+`
	`, string(switchIdsJSON))
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

	deviceLinkRows, err := d.db.QueryContext(ctx, `
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

func upsertDevicePosition(ctx context.Context, q querier, deviceId string, position models.Position) error {
	query := `
		INSERT INTO ` + schema.DevicePositionsTable + ` (` + schema.DevicePositions_DEVICE_ID + `, ` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + `)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (` + schema.DevicePositions_DEVICE_ID + `) DO UPDATE
		SET ` + schema.DevicePositions_POSITION_X + ` = excluded.` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + ` = excluded.` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + ` = excluded.` + schema.DevicePositions_POSITION_Z
	_, err := q.ExecContext(ctx, query, deviceId, position.X, position.Y, position.Z)
	return err
}
