package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
)

type sqliteDevices struct {
	db *sql.DB
}

func (d *sqliteDevices) GetDevices(ctx context.Context, _ string, spaceId int, include string) ([]models.Device, error) {
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

func (d *sqliteDevices) GetDevice(ctx context.Context, _ string, deviceId int, include string) (*models.Device, error) {
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

func (d *sqliteDevices) AddDevice(ctx context.Context, _ string, spaceId int, name string, position *models.Position) (*models.Device, error) {
	device := &models.Device{Name: name, SpaceId: spaceId, Position: position}

	err := withTx(ctx, d.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SpacesTable, schema.Spaces_ID, spaceId); err != nil {
			return err
		}

		insert := `INSERT INTO ` + schema.DevicesTable + ` (` + schema.Devices_NAME + `, ` + schema.Devices_SPACE_ID + `) VALUES ($1, $2) RETURNING ` + schema.Devices_ID
		if err := tx.QueryRowContext(ctx, insert, name, spaceId).Scan(&device.Id); err != nil {
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

func (d *sqliteDevices) SetDevice(ctx context.Context, _ string, deviceId int, name *string, position *models.Position) (*models.Device, error) {
	device := &models.Device{}

	err := withTx(ctx, d.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.DevicesTable, schema.Devices_ID, deviceId); err != nil {
			return err
		}

		if name != nil {
			rename := `UPDATE ` + schema.DevicesTable + ` SET ` + schema.Devices_NAME + ` = $1 WHERE ` + schema.Devices_ID + ` = $2`
			if _, err := tx.ExecContext(ctx, rename, *name, deviceId); err != nil {
				return err
			}
		}

		if position != nil {
			if err := upsertDevicePosition(ctx, tx, deviceId, *position); err != nil {
				return err
			}
		}

		result := `
			SELECT ` + schema.FullDevices_ID + `, ` + schema.FullDevices_NAME + `, ` + schema.FullDevices_SPACE_ID + `,
				` + schema.FullDevicePositions_POSITION_X + `, ` + schema.FullDevicePositions_POSITION_Y + `, ` + schema.FullDevicePositions_POSITION_Z + `
			FROM ` + schema.DevicesTable + `
			LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices + `
			WHERE ` + schema.FullDevices_ID + ` = $1
		`
		return tx.QueryRowContext(ctx, result, deviceId).
			Scan(&device.Id, &device.Name, &device.SpaceId, &device.PositionX, &device.PositionY, &device.PositionZ)
	})
	if err != nil {
		return nil, err
	}

	device.Build()

	return device, nil
}

func (d *sqliteDevices) SetDevicePosition(ctx context.Context, userEmail string, deviceId int, position models.Position) (*models.Device, error) {
	return d.SetDevice(ctx, userEmail, deviceId, nil, &position)
}

func (d *sqliteDevices) DelDevice(ctx context.Context, _ string, deviceId int) error {
	query := `DELETE FROM ` + schema.DevicesTable + ` WHERE ` + schema.Devices_ID + ` = $1`
	return notFoundIfNoneAffected(d.db.ExecContext(ctx, query, deviceId))
}

func (d *sqliteDevices) GetDeviceBreakers(ctx context.Context, _ string, deviceId int, include string) ([]models.Breaker, error) {
	if err := requireRow(ctx, d.db, schema.DevicesTable, schema.Devices_ID, deviceId); err != nil {
		return nil, err
	}

	keys, fields, err := models.ParseBreakerColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.DeviceBreakersTable + `
		JOIN ` + schema.BreakersTable + ` ON ` + schema.JoinDeviceBreakers_Breakers + `
		WHERE ` + schema.FullDeviceBreakers_DEVICE_ID + ` = $1`

	return queryBreakers(ctx, d.db, fields, query, deviceId)
}

func (d *sqliteDevices) AddDeviceBreaker(ctx context.Context, _ string, deviceId, breakerId int) error {
	return withTx(ctx, d.db, func(tx *sql.Tx) error {
		var spaceGroupId int
		spaceGroup := `
			SELECT ` + schema.FullSpaces_SPACE_GROUP_ID + `
			FROM ` + schema.DevicesTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinDevices_Spaces + `
			WHERE ` + schema.FullDevices_ID + ` = $1
		`
		if err := tx.QueryRowContext(ctx, spaceGroup, deviceId).Scan(&spaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if err := requireBreakerInSpaceGroup(ctx, tx, breakerId, spaceGroupId); err != nil {
			return err
		}

		insert := `INSERT INTO ` + schema.DeviceBreakersTable + ` (` + schema.DeviceBreakers_DEVICE_ID + `, ` + schema.DeviceBreakers_BREAKER_ID + `) VALUES ($1, $2)`
		_, err := tx.ExecContext(ctx, insert, deviceId, breakerId)
		if isUniqueViolation(err) {
			return repository.ErrConflict
		}
		return err
	})
}

func (d *sqliteDevices) DelDeviceBreaker(ctx context.Context, _ string, deviceId, breakerId int) error {
	query := `
		DELETE FROM ` + schema.DeviceBreakersTable + `
		WHERE ` + schema.DeviceBreakers_DEVICE_ID + ` = $1 AND ` + schema.DeviceBreakers_BREAKER_ID + ` = $2
	`
	return notFoundIfNoneAffected(d.db.ExecContext(ctx, query, deviceId, breakerId))
}

func (d *sqliteDevices) GetDeviceUpstream(ctx context.Context, _ string, deviceId int) ([]models.Breaker, error) {
	if err := requireRow(ctx, d.db, schema.DevicesTable, schema.Devices_ID, deviceId); err != nil {
		return nil, err
	}

	_, fields, err := models.ParseBreakerColumns("")
	if err != nil {
		return nil, err
	}

	query := `
		WITH RECURSIVE breaker_chain AS (
			SELECT ` + breakerColumns + `
			FROM ` + schema.DeviceBreakersTable + `
			JOIN ` + schema.BreakersTable + ` ON ` + schema.JoinDeviceBreakers_Breakers + `
			WHERE ` + schema.FullDeviceBreakers_DEVICE_ID + ` = $1

			UNION ALL

			SELECT ` + breakerColumns + `
			FROM ` + schema.BreakersTable + `
			JOIN breaker_chain ON ` + schema.FullBreakers_ID + ` = breaker_chain.` + schema.Breakers_UPSTREAM_BREAKER_ID + `
		)
		SELECT DISTINCT ` + schema.Breakers_ID + `, ` + schema.Breakers_NAME + `, ` + schema.Breakers_BREAKER_GROUP_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + `, ` + schema.Breakers_UPSTREAM_BREAKER_ID + `
		FROM breaker_chain
		ORDER BY ` + schema.Breakers_ID + `
	`
	return queryBreakers(ctx, d.db, fields, query, deviceId)
}

func upsertDevicePosition(ctx context.Context, q querier, deviceId int, position models.Position) error {
	query := `
		INSERT INTO ` + schema.DevicePositionsTable + ` (` + schema.DevicePositions_DEVICE_ID + `, ` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + `)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (` + schema.DevicePositions_DEVICE_ID + `) DO UPDATE
		SET ` + schema.DevicePositions_POSITION_X + ` = excluded.` + schema.DevicePositions_POSITION_X + `, ` + schema.DevicePositions_POSITION_Y + ` = excluded.` + schema.DevicePositions_POSITION_Y + `, ` + schema.DevicePositions_POSITION_Z + ` = excluded.` + schema.DevicePositions_POSITION_Z
	_, err := q.ExecContext(ctx, query, deviceId, position.X, position.Y, position.Z)
	return err
}
