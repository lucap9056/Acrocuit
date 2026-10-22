package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"encoding/json"
)

const switchColumns = schema.FullSwitches_ID + `, ` + schema.FullSwitches_NAME + `, ` + schema.FullSwitches_SWITCH_PANEL_ID + `, ` + schema.FullSwitches_SPACE_GROUP_ID + `, ` + schema.FullSwitches_DISPLAY_ORDER + `, ` + schema.FullSwitches_UPSTREAM_SWITCH_ID

const (
	switchChainWithPanelColumns = `switch_chain.` + schema.Switches_ID + `, switch_chain.` + schema.Switches_NAME + `, switch_chain.` + schema.Switches_SWITCH_PANEL_ID + `, switch_chain.` + schema.Switches_SPACE_GROUP_ID + `, switch_chain.` + schema.Switches_DISPLAY_ORDER + `, switch_chain.` + schema.Switches_UPSTREAM_SWITCH_ID + `, ` + models.SWITCH_PANEL_WITH_POSITION_COLUMNS
	switchChainJoinPanel        = `
		FROM switch_chain
		JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.FullSwitchPanels_ID + ` = switch_chain.` + schema.Switches_SWITCH_PANEL_ID + `
		LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels
)

type sqliteSwitches struct {
	db *sql.DB
}

func (s *sqliteSwitches) GetSwitches(ctx context.Context, _ string, switchPanelId int, include string) ([]models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchesTable + `
		WHERE ` + schema.FullSwitches_SWITCH_PANEL_ID + ` = $1
		ORDER BY ` + schema.FullSwitches_DISPLAY_ORDER

	return querySwitches(ctx, s.db, fields, query, switchPanelId)
}

func (s *sqliteSwitches) GetSwitch(ctx context.Context, _ string, switchId int, include string) (*models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.FullSwitches_ID + ` = $1`

	var sw models.Switch
	if err := s.db.QueryRowContext(ctx, query, switchId).Scan(sw.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}
	sw.Build()

	return &sw, nil
}

func (s *sqliteSwitches) AddSwitch(ctx context.Context, _ string, switchPanelId int, name string, upstreamSwitchId *int) (*models.Switch, error) {
	sw := &models.Switch{Name: name, SwitchPanelId: switchPanelId, UpstreamSwitchId: upstreamSwitchId}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		spaceGroup := `SELECT ` + schema.SwitchPanels_SPACE_GROUP_ID + ` FROM ` + schema.SwitchPanelsTable + ` WHERE ` + schema.SwitchPanels_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, switchPanelId).Scan(&sw.SpaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if upstreamSwitchId != nil {
			if err := requireSwitchInSpaceGroup(ctx, tx, *upstreamSwitchId, sw.SpaceGroupId); err != nil {
				return err
			}
		}

		insert := `
			INSERT INTO ` + schema.SwitchesTable + ` (` + schema.Switches_NAME + `, ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `, ` + schema.Switches_UPSTREAM_SWITCH_ID + `)
			SELECT $1, $2, $3, COALESCE(MAX(` + schema.Switches_DISPLAY_ORDER + `), 0) + 1, $4
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.Switches_SWITCH_PANEL_ID + ` = $2
			RETURNING ` + schema.Switches_ID + `, ` + schema.Switches_DISPLAY_ORDER
		return tx.QueryRowContext(ctx, insert, name, switchPanelId, sw.SpaceGroupId, upstreamSwitchId).
			Scan(&sw.Id, &sw.DisplayOrder)
	})
	if err != nil {
		return nil, err
	}

	sw.Build()

	return sw, nil
}

func (s *sqliteSwitches) MoveSwitch(ctx context.Context, _ string, switchId int, displayOrder int) ([]models.Switch, error) {
	var switches []models.Switch

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var switchPanelId, currentOrder int
		current := `SELECT ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_DISPLAY_ORDER + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, current, switchId).Scan(&switchPanelId, &currentOrder); err != nil {
			return notFoundIfNoRows(err)
		}

		if displayOrder != currentOrder {
			if err := switchesOrder.move(ctx, tx, switchPanelId, switchId, currentOrder, displayOrder); err != nil {
				return err
			}
		}

		keys, fields, err := models.ParseSwitchColumns("")
		if err != nil {
			return err
		}
		query := `SELECT ` + keys + `
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.FullSwitches_SWITCH_PANEL_ID + ` = $1
			ORDER BY ` + schema.FullSwitches_DISPLAY_ORDER
		switches, err = querySwitches(ctx, tx, fields, query, switchPanelId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return switches, nil
}

func (s *sqliteSwitches) SetSwitch(ctx context.Context, _ string, switchId int, name *string) (*models.Switch, error) {
	var sw *models.Switch

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SwitchesTable, schema.Switches_ID, switchId); err != nil {
			return err
		}

		if name != nil {
			rename := `UPDATE ` + schema.SwitchesTable + ` SET ` + schema.Switches_NAME + ` = $1 WHERE ` + schema.Switches_ID + ` = $2`
			if _, err := tx.ExecContext(ctx, rename, *name, switchId); err != nil {
				return err
			}
		}

		var err error
		sw, err = selectSwitch(ctx, tx, switchId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return sw, nil
}

func (s *sqliteSwitches) SetSwitchUpstream(ctx context.Context, _ string, switchId int, upstreamSwitchId *int) (*models.Switch, error) {
	var sw *models.Switch

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var spaceGroupId int
		spaceGroup := `SELECT ` + schema.Switches_SPACE_GROUP_ID + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, switchId).Scan(&spaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if upstreamSwitchId != nil {
			if *upstreamSwitchId == switchId {
				return repository.ErrSelfReference
			}

			if err := requireSwitchInSpaceGroup(ctx, tx, *upstreamSwitchId, spaceGroupId); err != nil {
				return err
			}

			var hasCycle bool
			cycle := `
				WITH RECURSIVE upstream_chain(id, upstream_id) AS (
					SELECT ` + schema.Switches_ID + `, ` + schema.Switches_UPSTREAM_SWITCH_ID + `
					FROM ` + schema.SwitchesTable + `
					WHERE ` + schema.Switches_ID + ` = $1

					UNION

					SELECT ` + schema.FullSwitches_ID + `, ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + `
					FROM ` + schema.SwitchesTable + `
					JOIN upstream_chain ON ` + schema.FullSwitches_ID + ` = upstream_chain.upstream_id
				)
				SELECT EXISTS (SELECT 1 FROM upstream_chain WHERE id = $2)
			`
			if err := tx.QueryRowContext(ctx, cycle, *upstreamSwitchId, switchId).Scan(&hasCycle); err != nil {
				return err
			}
			if hasCycle {
				return repository.ErrCycleDetected
			}
		}

		update := `UPDATE ` + schema.SwitchesTable + ` SET ` + schema.Switches_UPSTREAM_SWITCH_ID + ` = $1 WHERE ` + schema.Switches_ID + ` = $2`
		if _, err := tx.ExecContext(ctx, update, upstreamSwitchId, switchId); err != nil {
			return err
		}

		var err error
		sw, err = selectSwitch(ctx, tx, switchId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return sw, nil
}

func (s *sqliteSwitches) DelSwitch(ctx context.Context, _ string, switchId int) error {
	query := `DELETE FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $1`
	return notFoundIfNoneAffected(s.db.ExecContext(ctx, query, switchId))
}

func (s *sqliteSwitches) GetSwitchDownstream(ctx context.Context, _ string, switchId int) ([]models.Switch, []models.Device, error) {
	query := `
		WITH RECURSIVE switch_chain AS (
			SELECT ` + switchColumns + `, 0 AS depth
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.FullSwitches_ID + ` = $1

			UNION ALL

			SELECT ` + switchColumns + `, switch_chain.depth + 1
			FROM ` + schema.SwitchesTable + `
			JOIN switch_chain ON ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + ` = switch_chain.` + schema.Switches_ID + `
		)
		SELECT ` + switchChainWithPanelColumns + switchChainJoinPanel + `
		ORDER BY switch_chain.depth
	`
	switches, err := querySwitchesWithPanel(ctx, s.db, query, switchId)
	if err != nil {
		return nil, nil, err
	}
	if len(switches) == 0 {
		return nil, nil, repository.ErrNotFound
	}

	ids := make([]int, len(switches))
	for i, sw := range switches {
		ids[i] = sw.Id
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return nil, nil, err
	}

	deviceQuery := `
		SELECT ` + schema.FullDevices_ID + `, ` + schema.FullDevices_NAME + `, ` + schema.FullDevices_SPACE_ID + `, ` + schema.FullDeviceSwitches_SWITCH_ID + `, ` + schema.FullDevicePositions_POSITION_X + `, ` + schema.FullDevicePositions_POSITION_Y + `, ` + schema.FullDevicePositions_POSITION_Z + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.DeviceSwitchesTable + ` ON ` + schema.JoinDeviceSwitches_Devices + `
		LEFT JOIN ` + schema.DevicePositionsTable + ` ON ` + schema.JoinDevicePositions_Devices + `
		WHERE ` + schema.FullDeviceSwitches_SWITCH_ID + ` IN (SELECT value FROM json_each($1))
		ORDER BY ` + schema.FullDeviceSwitches_SWITCH_ID + `, ` + schema.FullDevices_ID + `
	`
	deviceRows, err := s.db.QueryContext(ctx, deviceQuery, string(idsJSON))
	if err != nil {
		return nil, nil, err
	}
	defer deviceRows.Close()

	devices := make([]models.Device, 0)
	for deviceRows.Next() {
		var d models.Device
		if err := deviceRows.Scan(&d.Id, &d.Name, &d.SpaceId, &d.SwitchId, &d.PositionX, &d.PositionY, &d.PositionZ); err != nil {
			return nil, nil, err
		}
		d.Build()
		devices = append(devices, d)
	}

	return switches, devices, deviceRows.Err()
}

func requireSwitchInSpaceGroup(ctx context.Context, q querier, switchId, spaceGroupId int) error {
	found, err := exists(ctx, q, `
		SELECT 1 FROM `+schema.SwitchesTable+`
		WHERE `+schema.Switches_ID+` = $1 AND `+schema.Switches_SPACE_GROUP_ID+` = $2
	`, switchId, spaceGroupId)
	if err != nil {
		return err
	}
	if !found {
		return repository.ErrNotFound
	}
	return nil
}

func selectSwitch(ctx context.Context, q querier, switchId int) (*models.Switch, error) {
	query := `SELECT ` + switchColumns + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.FullSwitches_ID + ` = $1`

	sw := &models.Switch{}
	err := q.QueryRowContext(ctx, query, switchId).
		Scan(&sw.Id, &sw.Name, &sw.SwitchPanelId, &sw.SpaceGroupId, &sw.DisplayOrder, &sw.UpstreamSwitchId)
	if err != nil {
		return nil, notFoundIfNoRows(err)
	}
	sw.Build()

	return sw, nil
}

func querySwitches(ctx context.Context, q querier, fields []uint8, query string, args ...any) ([]models.Switch, error) {
	rows, err := q.QueryContext(ctx, query, args...)
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

func querySwitchesWithPanel(ctx context.Context, q querier, query string, args ...any) ([]models.Switch, error) {
	rows, err := q.QueryContext(ctx, query, args...)
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
