package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"encoding/json"
)

const breakerColumns = schema.FullBreakers_ID + `, ` + schema.FullBreakers_NAME + `, ` + schema.FullBreakers_BREAKER_GROUP_ID + `, ` + schema.FullBreakers_SPACE_GROUP_ID + `, ` + schema.FullBreakers_DISPLAY_ORDER + `, ` + schema.FullBreakers_UPSTREAM_BREAKER_ID

type sqliteBreakers struct {
	db *sql.DB
}

func (b *sqliteBreakers) GetBreakers(ctx context.Context, _ string, breakerGroupId int, include string) ([]models.Breaker, error) {
	keys, fields, err := models.ParseBreakerColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.BreakersTable + `
		WHERE ` + schema.FullBreakers_BREAKER_GROUP_ID + ` = $1
		ORDER BY ` + schema.FullBreakers_DISPLAY_ORDER

	return queryBreakers(ctx, b.db, fields, query, breakerGroupId)
}

func (b *sqliteBreakers) GetBreaker(ctx context.Context, _ string, breakerId int, include string) (*models.Breaker, error) {
	keys, fields, err := models.ParseBreakerColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.BreakersTable + ` WHERE ` + schema.FullBreakers_ID + ` = $1`

	var breaker models.Breaker
	if err := b.db.QueryRowContext(ctx, query, breakerId).Scan(breaker.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}
	breaker.Build()

	return &breaker, nil
}

func (b *sqliteBreakers) AddBreaker(ctx context.Context, _ string, breakerGroupId int, name string, upstreamBreakerId *int) (*models.Breaker, error) {
	breaker := &models.Breaker{Name: name, BreakerGroupId: breakerGroupId, UpstreamBreakerId: upstreamBreakerId}

	err := withTx(ctx, b.db, func(tx *sql.Tx) error {
		spaceGroup := `SELECT ` + schema.BreakerGroups_SPACE_GROUP_ID + ` FROM ` + schema.BreakerGroupsTable + ` WHERE ` + schema.BreakerGroups_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, breakerGroupId).Scan(&breaker.SpaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if upstreamBreakerId != nil {
			if err := requireBreakerInSpaceGroup(ctx, tx, *upstreamBreakerId, breaker.SpaceGroupId); err != nil {
				return err
			}
		}

		insert := `
			INSERT INTO ` + schema.BreakersTable + ` (` + schema.Breakers_NAME + `, ` + schema.Breakers_BREAKER_GROUP_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + `, ` + schema.Breakers_UPSTREAM_BREAKER_ID + `)
			SELECT $1, $2, $3, COALESCE(MAX(` + schema.Breakers_DISPLAY_ORDER + `), 0) + 1, $4
			FROM ` + schema.BreakersTable + `
			WHERE ` + schema.Breakers_BREAKER_GROUP_ID + ` = $2
			RETURNING ` + schema.Breakers_ID + `, ` + schema.Breakers_DISPLAY_ORDER
		return tx.QueryRowContext(ctx, insert, name, breakerGroupId, breaker.SpaceGroupId, upstreamBreakerId).
			Scan(&breaker.Id, &breaker.DisplayOrder)
	})
	if err != nil {
		return nil, err
	}

	breaker.Build()

	return breaker, nil
}

func (b *sqliteBreakers) SetBreaker(ctx context.Context, _ string, breakerId int, name *string, displayOrder *int) (*models.Breaker, error) {
	var breaker *models.Breaker

	err := withTx(ctx, b.db, func(tx *sql.Tx) error {
		var breakerGroupId, currentOrder int
		current := `SELECT ` + schema.Breakers_BREAKER_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + ` FROM ` + schema.BreakersTable + ` WHERE ` + schema.Breakers_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, current, breakerId).Scan(&breakerGroupId, &currentOrder); err != nil {
			return notFoundIfNoRows(err)
		}

		if displayOrder != nil && *displayOrder != currentOrder {
			if err := breakersOrder.move(ctx, tx, breakerGroupId, breakerId, currentOrder, *displayOrder); err != nil {
				return err
			}
		}

		if name != nil {
			rename := `UPDATE ` + schema.BreakersTable + ` SET ` + schema.Breakers_NAME + ` = $1 WHERE ` + schema.Breakers_ID + ` = $2`
			if _, err := tx.ExecContext(ctx, rename, *name, breakerId); err != nil {
				return err
			}
		}

		var err error
		breaker, err = selectBreaker(ctx, tx, breakerId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return breaker, nil
}

func (b *sqliteBreakers) SetBreakerUpstream(ctx context.Context, _ string, breakerId int, upstreamBreakerId *int) (*models.Breaker, error) {
	var breaker *models.Breaker

	err := withTx(ctx, b.db, func(tx *sql.Tx) error {
		var spaceGroupId int
		spaceGroup := `SELECT ` + schema.Breakers_SPACE_GROUP_ID + ` FROM ` + schema.BreakersTable + ` WHERE ` + schema.Breakers_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, breakerId).Scan(&spaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if upstreamBreakerId != nil {
			if *upstreamBreakerId == breakerId {
				return repository.ErrSelfReference
			}

			if err := requireBreakerInSpaceGroup(ctx, tx, *upstreamBreakerId, spaceGroupId); err != nil {
				return err
			}

			var hasCycle bool
			cycle := `
				WITH RECURSIVE upstream_chain(id, upstream_id) AS (
					SELECT ` + schema.Breakers_ID + `, ` + schema.Breakers_UPSTREAM_BREAKER_ID + `
					FROM ` + schema.BreakersTable + `
					WHERE ` + schema.Breakers_ID + ` = $1

					UNION

					SELECT ` + schema.FullBreakers_ID + `, ` + schema.FullBreakers_UPSTREAM_BREAKER_ID + `
					FROM ` + schema.BreakersTable + `
					JOIN upstream_chain ON ` + schema.FullBreakers_ID + ` = upstream_chain.upstream_id
				)
				SELECT EXISTS (SELECT 1 FROM upstream_chain WHERE id = $2)
			`
			if err := tx.QueryRowContext(ctx, cycle, *upstreamBreakerId, breakerId).Scan(&hasCycle); err != nil {
				return err
			}
			if hasCycle {
				return repository.ErrCycleDetected
			}
		}

		update := `UPDATE ` + schema.BreakersTable + ` SET ` + schema.Breakers_UPSTREAM_BREAKER_ID + ` = $1 WHERE ` + schema.Breakers_ID + ` = $2`
		if _, err := tx.ExecContext(ctx, update, upstreamBreakerId, breakerId); err != nil {
			return err
		}

		var err error
		breaker, err = selectBreaker(ctx, tx, breakerId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return breaker, nil
}

func (b *sqliteBreakers) DelBreaker(ctx context.Context, _ string, breakerId int) error {
	query := `DELETE FROM ` + schema.BreakersTable + ` WHERE ` + schema.Breakers_ID + ` = $1`
	return notFoundIfNoneAffected(b.db.ExecContext(ctx, query, breakerId))
}

func (b *sqliteBreakers) GetBreakerDownstream(ctx context.Context, _ string, breakerId int) ([]models.Breaker, []models.Device, error) {
	_, fields, err := models.ParseBreakerColumns("")
	if err != nil {
		return nil, nil, err
	}

	query := `
		WITH RECURSIVE breaker_chain AS (
			SELECT ` + breakerColumns + `, 0 AS depth
			FROM ` + schema.BreakersTable + `
			WHERE ` + schema.FullBreakers_ID + ` = $1

			UNION ALL

			SELECT ` + breakerColumns + `, breaker_chain.depth + 1
			FROM ` + schema.BreakersTable + `
			JOIN breaker_chain ON ` + schema.FullBreakers_UPSTREAM_BREAKER_ID + ` = breaker_chain.` + schema.Breakers_ID + `
		)
		SELECT ` + schema.Breakers_ID + `, ` + schema.Breakers_NAME + `, ` + schema.Breakers_BREAKER_GROUP_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + `, ` + schema.Breakers_UPSTREAM_BREAKER_ID + `
		FROM breaker_chain
		ORDER BY depth
	`
	breakers, err := queryBreakers(ctx, b.db, fields, query, breakerId)
	if err != nil {
		return nil, nil, err
	}
	if len(breakers) == 0 {
		return nil, nil, repository.ErrNotFound
	}

	ids := make([]int, len(breakers))
	for i, breaker := range breakers {
		ids[i] = breaker.Id
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return nil, nil, err
	}

	deviceQuery := `
		SELECT ` + schema.FullDevices_ID + `, ` + schema.FullDevices_NAME + `, ` + schema.FullDevices_SPACE_ID + `, ` + schema.FullDeviceBreakers_BREAKER_ID + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.DeviceBreakersTable + ` ON ` + schema.JoinDeviceBreakers_Devices + `
		WHERE ` + schema.FullDeviceBreakers_BREAKER_ID + ` IN (SELECT value FROM json_each($1))
		ORDER BY ` + schema.FullDeviceBreakers_BREAKER_ID + `, ` + schema.FullDevices_ID + `
	`
	deviceRows, err := b.db.QueryContext(ctx, deviceQuery, string(idsJSON))
	if err != nil {
		return nil, nil, err
	}
	defer deviceRows.Close()

	devices := make([]models.Device, 0)
	for deviceRows.Next() {
		var d models.Device
		if err := deviceRows.Scan(&d.Id, &d.Name, &d.SpaceId, &d.BreakerId); err != nil {
			return nil, nil, err
		}
		d.Build()
		devices = append(devices, d)
	}

	return breakers, devices, deviceRows.Err()
}

func requireBreakerInSpaceGroup(ctx context.Context, q querier, breakerId, spaceGroupId int) error {
	found, err := exists(ctx, q, `
		SELECT 1 FROM `+schema.BreakersTable+`
		WHERE `+schema.Breakers_ID+` = $1 AND `+schema.Breakers_SPACE_GROUP_ID+` = $2
	`, breakerId, spaceGroupId)
	if err != nil {
		return err
	}
	if !found {
		return repository.ErrNotFound
	}
	return nil
}

func selectBreaker(ctx context.Context, q querier, breakerId int) (*models.Breaker, error) {
	query := `SELECT ` + breakerColumns + ` FROM ` + schema.BreakersTable + ` WHERE ` + schema.FullBreakers_ID + ` = $1`

	breaker := &models.Breaker{}
	err := q.QueryRowContext(ctx, query, breakerId).
		Scan(&breaker.Id, &breaker.Name, &breaker.BreakerGroupId, &breaker.SpaceGroupId, &breaker.DisplayOrder, &breaker.UpstreamBreakerId)
	if err != nil {
		return nil, notFoundIfNoRows(err)
	}
	breaker.Build()

	return breaker, nil
}

func queryBreakers(ctx context.Context, q querier, fields []uint8, query string, args ...any) ([]models.Breaker, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakers := make([]models.Breaker, 0)
	for rows.Next() {
		var breaker models.Breaker
		if err := rows.Scan(breaker.Values(fields)...); err != nil {
			return nil, err
		}
		breaker.Build()
		breakers = append(breakers, breaker)
	}

	return breakers, rows.Err()
}
