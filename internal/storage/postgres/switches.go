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

func (s *pgSwitches) GetSwitches(ctx context.Context, userEmail string, switchPanelId int, include string) ([]models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchesTable + `
		JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.JoinSwitches_SwitchPanels + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SwitchPanels + `
		WHERE ` + schema.FullSwitches_SWITCH_PANEL_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1
		ORDER BY ` + schema.FullSwitches_DISPLAY_ORDER

	return querySwitches(ctx, s.pool, fields, query, userEmail, switchPanelId)
}

func (s *pgSwitches) GetSwitch(ctx context.Context, userEmail string, switchId int, include string) (*models.Switch, error) {
	keys, fields, err := models.ParseSwitchColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchesTable + `
		JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.JoinSwitches_SwitchPanels + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SwitchPanels + `
		WHERE ` + schema.FullSwitches_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1`

	var sw models.Switch
	err = s.pool.QueryRow(ctx, query, userEmail, switchId).Scan(sw.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sw.Build()

	return &sw, nil
}

func (s *pgSwitches) AddSwitch(ctx context.Context, userEmail string, switchPanelId int, name string, attributes models.Attributes) (*models.Switch, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSwitchPanels_SPACE_GROUP_ID + `
			FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SwitchPanels + `
			WHERE ` + schema.FullSwitchPanels_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
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

func (s *pgSwitches) MoveSwitch(ctx context.Context, userEmail string, switchId int, displayOrder int) ([]models.Switch, error) {
	query := `SELECT ` + schema.FnMoveSwitch_OUTPUTS + ` FROM ` + schema.FnMoveSwitch + `($1, $2, $3)`

	rows, err := s.pool.Query(ctx, query, userEmail, switchId, displayOrder)
	switches, err := scanAll(rows, err, (*models.Switch).AllValues)
	if err != nil {
		return nil, err
	}

	if len(switches) == 0 {
		return nil, repository.ErrNotFound
	}

	for i := range switches {
		switches[i].Build()
	}

	return switches, nil
}

func (s *pgSwitches) SetSwitch(ctx context.Context, userEmail string, switchId int, name *string, attributes models.Attributes) (*models.Switch, error) {
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

func (s *pgSwitches) DelSwitch(ctx context.Context, userEmail string, switchId int) error {
	query := `
		DELETE FROM ` + schema.SwitchesTable + `
		WHERE ` + schema.FullSwitches_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SwitchPanels + `
			WHERE ` + schema.JoinSwitches_SwitchPanels + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
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

func (s *pgSwitches) GetSwitchUpstreams(ctx context.Context, userEmail string, switchId int, include string) ([]models.LinkedSwitch, error) {
	existsQuery := `
		SELECT EXISTS (
			SELECT 1 FROM ` + schema.SwitchesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Switches + `
			WHERE ` + schema.FullSwitches_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	var exists bool
	if err := s.pool.QueryRow(ctx, existsQuery, switchId, userEmail).Scan(&exists); err != nil {
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

	return queryLinkedSwitches(ctx, s.pool, fields, query, switchId)
}

func (s *pgSwitches) AddSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId int, attributes models.Attributes) (*models.SwitchLink, error) {
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

func (s *pgSwitches) SetSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId int, attributes models.Attributes) (*models.SwitchLink, error) {
	query := `
		UPDATE ` + schema.SwitchLinksTable + `
		SET ` + schema.SwitchLinks_ATTRIBUTES + ` = $1
		WHERE ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $2 AND ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + ` = $3 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
			WHERE ` + schema.JoinSpaceGroupOwners_SwitchLinks + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $4
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

func (s *pgSwitches) DelSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId int) error {
	query := `
		DELETE FROM ` + schema.SwitchLinksTable + `
		WHERE ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + ` = $1 AND ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + ` = $2 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
			WHERE ` + schema.JoinSpaceGroupOwners_SwitchLinks + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $3
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

func (s *pgSwitches) GetSwitchDownstream(ctx context.Context, userEmail string, switchId int) (*models.PowerChain, error) {
	query := `
		WITH RECURSIVE switch_chain(id) AS (
			SELECT ` + schema.FullSwitches_ID + `
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.FullSwitches_ID + ` = $1 AND EXISTS (
				SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
				WHERE ` + schema.JoinSpaceGroupOwners_Switches + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			)

			UNION

			SELECT ` + schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID + `
			FROM ` + schema.SwitchLinksTable + `
			JOIN switch_chain ON ` + schema.FullSwitchLinks_UPSTREAM_SWITCH_ID + ` = switch_chain.id
		)
	` + selectSwitchChainWithPanel
	switches, err := querySwitchesWithPanel(ctx, s.pool, query, switchId, userEmail)
	if err != nil {
		return nil, err
	}
	if len(switches) == 0 {
		return nil, repository.ErrNotFound
	}

	ids := switchIdsOf(switches)
	chain := &models.PowerChain{Switches: switches}

	chain.SwitchLinks, err = querySwitchLinks(ctx, s.pool, `
		SELECT `+models.SWITCH_LINK_COLUMNS+`
		FROM `+schema.SwitchLinksTable+`
		WHERE `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+` = ANY($1)
		ORDER BY `+schema.FullSwitchLinks_UPSTREAM_SWITCH_ID+`, `+schema.FullSwitchLinks_DOWNSTREAM_SWITCH_ID+`
	`, ids)
	if err != nil {
		return nil, err
	}

	chain.Devices, err = queryDevicesWithPosition(ctx, s.pool, `
		SELECT `+models.DEVICE_WITH_POSITION_COLUMNS+`
		FROM `+schema.DevicesTable+`
		LEFT JOIN `+schema.DevicePositionsTable+` ON `+schema.JoinDevicePositions_Devices+`
		WHERE `+schema.FullDevices_ID+` IN (
			SELECT `+schema.FullDeviceSwitches_DEVICE_ID+`
			FROM `+schema.DeviceSwitchesTable+`
			WHERE `+schema.FullDeviceSwitches_SWITCH_ID+` = ANY($1)
		)
		ORDER BY `+schema.FullDevices_ID+`
	`, ids)
	if err != nil {
		return nil, err
	}

	chain.DeviceLinks, err = queryDeviceLinks(ctx, s.pool, `
		SELECT `+models.DEVICE_LINK_COLUMNS+`
		FROM `+schema.DeviceSwitchesTable+`
		WHERE `+schema.FullDeviceSwitches_SWITCH_ID+` = ANY($1)
		ORDER BY `+schema.FullDeviceSwitches_SWITCH_ID+`, `+schema.FullDeviceSwitches_DEVICE_ID+`
	`, ids)
	if err != nil {
		return nil, err
	}

	return chain, nil
}

func querySwitches(ctx context.Context, pool *pgxpool.Pool, fields []uint8, query string, args ...any) ([]models.Switch, error) {
	rows, err := pool.Query(ctx, query, args...)
	switches, err := scanAll(rows, err, func(sw *models.Switch) []any { return sw.Values(fields) })
	for i := range switches {
		switches[i].Build()
	}
	return switches, err
}

func queryLinkedSwitches(ctx context.Context, pool *pgxpool.Pool, fields []uint8, query string, args ...any) ([]models.LinkedSwitch, error) {
	rows, err := pool.Query(ctx, query, args...)
	switches, err := scanAll(rows, err, func(ls *models.LinkedSwitch) []any { return ls.Values(fields) })
	for i := range switches {
		switches[i].Build()
	}
	return switches, err
}
