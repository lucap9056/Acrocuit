package postgres

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolation = "23505"

type pgSwitches struct {
	pool *pgxpool.Pool
}

func (s *pgSwitches) GetSwitches(ctx context.Context, userEmail string, switchPanelId string, include string) ([]models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchesTable + `
		WHERE ` + schema.FullSwitches_SWITCH_PANEL_ID + ` = $1
		  AND EXISTS (
			  SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			  WHERE ` + schema.JoinSpaceGroupACLs_Switches + `
			    AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		  )
		ORDER BY ` + schema.FullSwitches_DISPLAY_ORDER

	rows, err := s.pool.Query(ctx, query, switchPanelId, userEmail, models.PublicAccessEmail)
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

func (s *pgSwitches) GetSwitch(ctx context.Context, userEmail string, switchId string, include string) (*models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchesTable + `
		WHERE ` + schema.FullSwitches_ID + ` = $1
		  AND EXISTS (
			  SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			  WHERE ` + schema.JoinSpaceGroupACLs_Switches + `
			    AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		  )`

	var sw models.Switch
	err = s.pool.QueryRow(ctx, query, switchId, userEmail, models.PublicAccessEmail).Scan(sw.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sw.Build()

	return &sw, nil
}

func (s *pgSwitches) AddSwitch(ctx context.Context, userEmail string, switchPanelId string, name string, attributes models.Attributes) (*models.Switch, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSwitchPanels_SPACE_GROUP_ID + `
			FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_SwitchPanels + `
			WHERE ` + schema.FullSwitchPanels_ID + ` = $1 AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
			FOR UPDATE OF ` + schema.SwitchPanelsTable + `
		),
		next_order AS (
			SELECT COALESCE(MAX(` + schema.Switches_DISPLAY_ORDER + `), 0) + 1 AS ` + schema.Switches_DISPLAY_ORDER + `
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.Switches_SWITCH_PANEL_ID + ` = $1
		),
		inserted AS (
			INSERT INTO ` + schema.SwitchesTable + ` (` + schema.Switches_NAME + `, ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `, ` + schema.Switches_ATTRIBUTES + `)
			SELECT $3, $1, permitted.` + schema.SwitchPanels_SPACE_GROUP_ID + `, next_order.` + schema.Switches_DISPLAY_ORDER + `, $4::json
			FROM permitted, next_order
			RETURNING ` + schema.Switches_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `
		)
		SELECT ` + schema.Switches_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + ` FROM inserted
	`

	sw := &models.Switch{Name: name, SwitchPanelId: switchPanelId, Attributes: attributes}

	err := s.pool.QueryRow(ctx, query, switchPanelId, userEmail, name, attributes).
		Scan(&sw.Id, &sw.SpaceGroupId, &sw.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sw.Build()

	return sw, nil
}

func (s *pgSwitches) MoveSwitch(ctx context.Context, userEmail string, switchId string, displayOrder int) ([]models.Switch, error) {
	query := `SELECT ` + schema.FnMoveSwitch_OUTPUTS + ` FROM ` + schema.FnMoveSwitch + `($1, $2, $3)`

	rows, err := s.pool.Query(ctx, query, userEmail, switchId, displayOrder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switches := make([]models.Switch, 0)
	for rows.Next() {
		var sw models.Switch
		if err := rows.Scan(sw.AllValues()...); err != nil {
			return nil, err
		}
		sw.Build()
		switches = append(switches, sw)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(switches) == 0 {
		return nil, repository.ErrNotFound
	}

	return switches, nil
}

func (s *pgSwitches) SetSwitch(ctx context.Context, userEmail string, switchId string, name *string, attributes models.Attributes) (*models.Switch, error) {
	query := `SELECT ` + schema.FnSetSwitch_OUTPUTS + ` FROM ` + schema.FnSetSwitch + `($1, $2, $3, $4)`

	sw := &models.Switch{}
	err := s.pool.QueryRow(ctx, query, userEmail, switchId, name, attributes).Scan(sw.AllValues()...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sw.Build()

	return sw, nil
}

func (s *pgSwitches) DelSwitch(ctx context.Context, userEmail string, switchId string) error {
	query := `
		DELETE FROM ` + schema.SwitchesTable + `
		WHERE ` + schema.FullSwitches_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_SwitchPanels + `
			WHERE ` + schema.JoinSwitches_SwitchPanels + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
		)
	`
	t, err := s.pool.Exec(ctx, query, switchId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSwitches) GetSwitchUpstreams(ctx context.Context, userEmail string, switchId string, include string) ([]models.LinkedSwitch, error) {
	existsQuery := `
		SELECT EXISTS (
			SELECT 1 FROM ` + schema.SwitchesTable + `
			JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_Switches + `
			WHERE ` + schema.FullSwitches_ID + ` = $1 AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		)
	`
	var exists bool
	if err := s.pool.QueryRow(ctx, existsQuery, switchId, userEmail, models.PublicAccessEmail).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
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

	rows, err := s.pool.Query(ctx, query, switchId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	upstreams := make([]models.LinkedSwitch, 0)
	for rows.Next() {
		var upstream models.LinkedSwitch
		if err := rows.Scan(upstream.Values(fields)...); err != nil {
			return nil, err
		}
		upstream.Build()
		upstreams = append(upstreams, upstream)
	}

	return upstreams, rows.Err()
}

func (s *pgSwitches) AddSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId string, attributes models.Attributes) (*models.SwitchLink, error) {
	query := `SELECT ` + schema.FnAddSwitchUpstream_OUTPUTS + ` FROM ` + schema.FnAddSwitchUpstream + `($1, $2, $3, $4)`

	var link models.SwitchLink
	err := s.pool.QueryRow(ctx, query, userEmail, switchId, upstreamSwitchId, attributes).Scan(link.Values()...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case schema.FnAddSwitchUpstream_ERR_SELF_REFERENCE:
			return nil, repository.ErrSelfReference
		case schema.FnAddSwitchUpstream_ERR_CYCLE_DETECTED:
			return nil, repository.ErrCycleDetected
		case pgUniqueViolation:
			return nil, repository.ErrConflict
		}
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (s *pgSwitches) SetSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId string, attributes models.Attributes) (*models.SwitchLink, error) {
	query := `
		UPDATE ` + schema.SwitchLinksTable + `
		SET ` + schema.SwitchLinks_ATTRIBUTES + ` = $1
		WHERE ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $2 AND ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + ` = $3 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			WHERE ` + schema.JoinSpaceGroupACLs_SwitchLinks + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $4 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
		)
		RETURNING ` + models.SWITCH_LINK_COLUMNS

	var link models.SwitchLink
	err := s.pool.QueryRow(ctx, query, attributes, switchId, upstreamSwitchId, userEmail).Scan(link.Values()...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (s *pgSwitches) DelSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId string) error {
	query := `
		DELETE FROM ` + schema.SwitchLinksTable + `
		WHERE ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $1 AND ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + ` = $2 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			WHERE ` + schema.JoinSpaceGroupACLs_SwitchLinks + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
		)
	`
	t, err := s.pool.Exec(ctx, query, switchId, upstreamSwitchId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSwitches) GetSwitchDownstream(ctx context.Context, userEmail string, switchId string) (*models.PowerChain, error) {
	switchRows, err := s.pool.Query(ctx, `
		WITH RECURSIVE switch_chain(id) AS (
			SELECT `+schema.FullSwitches_ID+`
			FROM `+schema.SwitchesTable+`
			WHERE `+schema.FullSwitches_ID+` = $1 AND EXISTS (
				SELECT 1 FROM `+schema.SpaceGroupACLsTable+`
				WHERE `+schema.JoinSpaceGroupACLs_Switches+` AND `+schema.FullSpaceGroupACLs_USER_EMAIL+` IN ($2, $3)
			)

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
	`, switchId, userEmail, models.PublicAccessEmail)
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

	switchLinkRows, err := s.pool.Query(ctx, `
		SELECT `+models.SWITCH_LINK_COLUMNS+`
		FROM `+schema.SwitchLinksTable+`
		WHERE `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+` = ANY($1::text[]::uuid[])
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

	deviceRows, err := s.pool.Query(ctx, `
		SELECT `+models.DEVICE_WITH_POSITION_COLUMNS+`
		FROM `+schema.DevicesTable+`
		LEFT JOIN `+schema.DevicePositionsTable+` ON `+schema.JoinDevicePositions_Devices+`
		WHERE `+schema.FullDevices_ID+` IN (
			SELECT `+schema.FullDeviceSwitches_DEVICE_ID+`
			FROM `+schema.DeviceSwitchesTable+`
			WHERE `+schema.FullDeviceSwitches_SWITCH_ID+` = ANY($1::text[]::uuid[])
		)
		ORDER BY `+schema.FullDevices_ID+`
	`, switchIds)
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

	deviceLinkRows, err := s.pool.Query(ctx, `
		SELECT `+models.DEVICE_LINK_COLUMNS+`
		FROM `+schema.DeviceSwitchesTable+`
		WHERE `+schema.FullDeviceSwitches_SWITCH_ID+` = ANY($1::text[]::uuid[])
		ORDER BY `+schema.FullDeviceSwitches_SWITCH_ID+`, `+schema.FullDeviceSwitches_DEVICE_ID+`
	`, switchIds)
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
