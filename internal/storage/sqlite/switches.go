package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"encoding/json"
)

type sqliteSwitches struct {
	db *sql.DB
}

func (s *sqliteSwitches) GetSwitches(ctx context.Context, _ string, switchPanelId string, include string) ([]models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchesTable + `
		WHERE ` + schema.FullSwitches_SWITCH_PANEL_ID + ` = $1
		ORDER BY ` + schema.FullSwitches_DISPLAY_ORDER

	rows, err := s.db.QueryContext(ctx, query, switchPanelId)
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

func (s *sqliteSwitches) GetSwitch(ctx context.Context, _ string, switchId string, include string) (*models.Switch, error) {
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

func (s *sqliteSwitches) AddSwitch(ctx context.Context, _ string, switchPanelId string, name string, attributes models.Attributes) (*models.Switch, error) {
	sw := &models.Switch{Name: name, SwitchPanelId: switchPanelId, Attributes: attributes}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		spaceGroup := `SELECT ` + schema.SwitchPanels_SPACE_GROUP_ID + ` FROM ` + schema.SwitchPanelsTable + ` WHERE ` + schema.SwitchPanels_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, switchPanelId).Scan(&sw.SpaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		insert := `
			INSERT INTO ` + schema.SwitchesTable + ` (` + schema.Switches_NAME + `, ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `, ` + schema.Switches_ATTRIBUTES + `)
			SELECT $1, $2, $3, COALESCE(MAX(` + schema.Switches_DISPLAY_ORDER + `), 0) + 1, $4
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.Switches_SWITCH_PANEL_ID + ` = $2
			RETURNING ` + schema.Switches_ID + `, ` + schema.Switches_DISPLAY_ORDER
		return tx.QueryRowContext(ctx, insert, name, switchPanelId, sw.SpaceGroupId, attributes).
			Scan(&sw.Id, &sw.DisplayOrder)
	})
	if err != nil {
		return nil, err
	}

	sw.Build()

	return sw, nil
}

func (s *sqliteSwitches) MoveSwitch(ctx context.Context, _ string, switchId string, displayOrder int) ([]models.Switch, error) {
	var switches []models.Switch

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var switchPanelId string
		var currentOrder int
		current := `SELECT ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_DISPLAY_ORDER + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, current, switchId).Scan(&switchPanelId, &currentOrder); err != nil {
			return notFoundIfNoRows(err)
		}

		if displayOrder != currentOrder {
			if err := switchesOrder.move(ctx, tx, switchPanelId, switchId, currentOrder, displayOrder); err != nil {
				return err
			}
		}

		query := `SELECT ` + models.SWITCH_ALL_COLUMNS + `
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.FullSwitches_SWITCH_PANEL_ID + ` = $1
			ORDER BY ` + schema.FullSwitches_DISPLAY_ORDER
		rows, err := tx.QueryContext(ctx, query, switchPanelId)
		if err != nil {
			return err
		}
		defer rows.Close()

		switches = make([]models.Switch, 0)
		for rows.Next() {
			var sw models.Switch
			if err := rows.Scan(sw.AllValues()...); err != nil {
				return err
			}
			sw.Build()
			switches = append(switches, sw)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return switches, nil
}

func (s *sqliteSwitches) SetSwitch(ctx context.Context, _ string, switchId string, name *string, attributes models.Attributes) (*models.Switch, error) {
	sw := &models.Switch{}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		update := `
			UPDATE ` + schema.SwitchesTable + `
			SET ` + schema.Switches_NAME + ` = COALESCE($1, ` + schema.Switches_NAME + `),
				` + schema.Switches_ATTRIBUTES + ` = COALESCE($2, ` + schema.Switches_ATTRIBUTES + `)
			WHERE ` + schema.Switches_ID + ` = $3
		`
		if err := notFoundIfNoneAffected(tx.ExecContext(ctx, update, name, attributes, switchId)); err != nil {
			return err
		}

		query := `SELECT ` + models.SWITCH_ALL_COLUMNS + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.FullSwitches_ID + ` = $1`
		return notFoundIfNoRows(tx.QueryRowContext(ctx, query, switchId).Scan(sw.AllValues()...))
	})
	if err != nil {
		return nil, err
	}
	sw.Build()

	return sw, nil
}

func (s *sqliteSwitches) DelSwitch(ctx context.Context, _ string, switchId string) error {
	query := `DELETE FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $1`
	return notFoundIfNoneAffected(s.db.ExecContext(ctx, query, switchId))
}

func (s *sqliteSwitches) GetSwitchUpstreams(ctx context.Context, _ string, switchId string, include string) ([]models.LinkedSwitch, error) {
	if err := requireRow(ctx, s.db, schema.SwitchesTable, schema.Switches_ID, switchId); err != nil {
		return nil, err
	}

	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `,` + schema.FullSwitchLinks_ATTRIBUTES + `
		FROM ` + schema.SwitchLinksTable + `
		JOIN ` + schema.SwitchesTable + ` ON ` + schema.FullSwitches_ID + ` = ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + `
		WHERE ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $1
		ORDER BY ` + schema.FullSwitches_ID

	rows, err := s.db.QueryContext(ctx, query, switchId)
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

func (s *sqliteSwitches) AddSwitchUpstream(ctx context.Context, _ string, switchId, upstreamSwitchId string, attributes models.Attributes) (*models.SwitchLink, error) {
	link := &models.SwitchLink{UpstreamSwitchId: upstreamSwitchId, DownstreamSwitchId: switchId, Attributes: attributes}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var spaceGroupId string
		spaceGroup := `SELECT ` + schema.Switches_SPACE_GROUP_ID + ` FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, switchId).Scan(&spaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		if upstreamSwitchId == switchId {
			return repository.ErrSelfReference
		}

		if err := requireSwitchInSpaceGroup(ctx, tx, upstreamSwitchId, spaceGroupId); err != nil {
			return err
		}

		var hasCycle bool
		cycle := `
			WITH RECURSIVE upstream_chain(id) AS (
				SELECT $1

				UNION

				SELECT ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + `
				FROM ` + schema.SwitchLinksTable + `
				JOIN upstream_chain ON ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + ` = upstream_chain.id
			)
			SELECT EXISTS (SELECT 1 FROM upstream_chain WHERE id = $2)
		`
		if err := tx.QueryRowContext(ctx, cycle, upstreamSwitchId, switchId).Scan(&hasCycle); err != nil {
			return err
		}
		if hasCycle {
			return repository.ErrCycleDetected
		}

		insert := `
			INSERT INTO ` + schema.SwitchLinksTable + ` (` + schema.SwitchLinks_UPSTREAM_SWITCH_ID + `, ` + schema.SwitchLinks_DOWNSTREAM_SWITCH_ID + `, ` + schema.SwitchLinks_SPACE_GROUP_ID + `, ` + schema.SwitchLinks_ATTRIBUTES + `)
			VALUES ($1, $2, $3, $4)
		`
		_, err := tx.ExecContext(ctx, insert, upstreamSwitchId, switchId, spaceGroupId, attributes)
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

func (s *sqliteSwitches) SetSwitchUpstream(ctx context.Context, _ string, switchId, upstreamSwitchId string, attributes models.Attributes) (*models.SwitchLink, error) {
	query := `
		UPDATE ` + schema.SwitchLinksTable + `
		SET ` + schema.SwitchLinks_ATTRIBUTES + ` = $1
		WHERE ` + schema.SwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $2 AND ` + schema.SwitchLinks_UPSTREAM_SWITCH_ID + ` = $3
		RETURNING ` + models.SWITCH_LINK_COLUMNS

	var link models.SwitchLink
	if err := s.db.QueryRowContext(ctx, query, attributes, switchId, upstreamSwitchId).Scan(link.Values()...); err != nil {
		return nil, notFoundIfNoRows(err)
	}

	return &link, nil
}

func (s *sqliteSwitches) DelSwitchUpstream(ctx context.Context, _ string, switchId, upstreamSwitchId string) error {
	query := `
		DELETE FROM ` + schema.SwitchLinksTable + `
		WHERE ` + schema.SwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $1 AND ` + schema.SwitchLinks_UPSTREAM_SWITCH_ID + ` = $2
	`
	return notFoundIfNoneAffected(s.db.ExecContext(ctx, query, switchId, upstreamSwitchId))
}

func (s *sqliteSwitches) GetSwitchDownstream(ctx context.Context, _ string, switchId string) (*models.PowerChain, error) {
	switchRows, err := s.db.QueryContext(ctx, `
		WITH RECURSIVE switch_chain(id) AS (
			SELECT `+schema.FullSwitches_ID+`
			FROM `+schema.SwitchesTable+`
			WHERE `+schema.FullSwitches_ID+` = $1

			UNION

			SELECT `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+`
			FROM `+schema.SwitchLinksTable+`
			JOIN switch_chain ON `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+` = switch_chain.id
		)
		SELECT `+models.SWITCH_ALL_COLUMNS+`, `+models.SWITCH_PANEL_WITH_POSITION_COLUMNS+`
		FROM switch_chain
		JOIN `+schema.SwitchesTable+` ON `+schema.FullSwitches_ID+` = switch_chain.id
		JOIN `+schema.SwitchPanelsTable+` ON `+schema.JoinSwitches_SwitchPanels+`
		LEFT JOIN `+schema.SwitchPanelPositionsTable+` ON `+schema.JoinSwitchPanelPositions_SwitchPanels+`
		ORDER BY `+schema.FullSwitches_ID+`
	`, switchId)
	if err != nil {
		return nil, err
	}
	defer switchRows.Close()

	chain := &models.PowerChain{
		Switches:    make([]models.ChainSwitch, 0),
		SwitchLinks: make([]models.SwitchLink, 0),
		Devices:     make([]models.Device, 0),
		DeviceLinks: make([]models.DeviceLink, 0),
	}
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
	if len(chain.Switches) == 0 {
		return nil, repository.ErrNotFound
	}

	switchIdsJSON, err := json.Marshal(switchIds)
	if err != nil {
		return nil, err
	}

	switchLinkRows, err := s.db.QueryContext(ctx, `
		SELECT `+models.SWITCH_LINK_COLUMNS+`
		FROM `+schema.SwitchLinksTable+`
		WHERE `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+` IN (SELECT value FROM json_each($1))
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

	deviceRows, err := s.db.QueryContext(ctx, `
		SELECT `+models.DEVICE_WITH_POSITION_COLUMNS+`
		FROM `+schema.DevicesTable+`
		LEFT JOIN `+schema.DevicePositionsTable+` ON `+schema.JoinDevicePositions_Devices+`
		WHERE `+schema.FullDevices_ID+` IN (
			SELECT `+schema.FullDeviceSwitches_DEVICE_ID+`
			FROM `+schema.DeviceSwitchesTable+`
			WHERE `+schema.FullDeviceSwitches_SWITCH_ID+` IN (SELECT value FROM json_each($1))
		)
		ORDER BY `+schema.FullDevices_ID+`
	`, string(switchIdsJSON))
	if err != nil {
		return nil, err
	}
	defer deviceRows.Close()

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

	deviceLinkRows, err := s.db.QueryContext(ctx, `
		SELECT `+models.DEVICE_LINK_COLUMNS+`
		FROM `+schema.DeviceSwitchesTable+`
		WHERE `+schema.FullDeviceSwitches_SWITCH_ID+` IN (SELECT value FROM json_each($1))
		ORDER BY `+schema.FullDeviceSwitches_SWITCH_ID+`, `+schema.FullDeviceSwitches_DEVICE_ID+`
	`, string(switchIdsJSON))
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

func requireSwitchInSpaceGroup(ctx context.Context, q querier, switchId, spaceGroupId string) error {
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
