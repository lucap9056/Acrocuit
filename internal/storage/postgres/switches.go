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

	rows, err := s.pool.Query(ctx, query, userEmail, switchPanelId)
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

func (s *pgSwitches) AddSwitch(ctx context.Context, userEmail string, switchPanelId int, name string, upstreamSwitchId *int) (*models.Switch, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSwitchPanels_SPACE_GROUP_ID + `
			FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SwitchPanels + `
			WHERE ` + schema.FullSwitchPanels_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			FOR UPDATE OF ` + schema.SwitchPanelsTable + `
		),
		upstream_ok AS (
			SELECT 1 FROM permitted
			WHERE $4::integer IS NULL OR EXISTS (
				SELECT 1 FROM ` + schema.SwitchesTable + ` WHERE ` + schema.Switches_ID + ` = $4 AND ` + schema.Switches_SPACE_GROUP_ID + ` = (SELECT ` + schema.SwitchPanels_SPACE_GROUP_ID + ` FROM permitted)
			)
		),
		next_order AS (
			SELECT COALESCE(MAX(` + schema.Switches_DISPLAY_ORDER + `), 0) + 1 AS ` + schema.Switches_DISPLAY_ORDER + `
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.Switches_SWITCH_PANEL_ID + ` = $1
		),
		inserted AS (
			INSERT INTO ` + schema.SwitchesTable + ` (` + schema.Switches_NAME + `, ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `, ` + schema.Switches_UPSTREAM_SWITCH_ID + `)
			SELECT $3, $1, permitted.` + schema.SwitchPanels_SPACE_GROUP_ID + `, next_order.` + schema.Switches_DISPLAY_ORDER + `, $4
			FROM permitted, next_order
			WHERE EXISTS (SELECT 1 FROM upstream_ok)
			RETURNING ` + schema.Switches_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `
		)
		SELECT ` + schema.Switches_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + ` FROM inserted
	`

	sw := &models.Switch{Name: name, SwitchPanelId: switchPanelId}

	err := s.pool.QueryRow(ctx, query, switchPanelId, userEmail, name, upstreamSwitchId).
		Scan(&sw.Id, &sw.SpaceGroupId, &sw.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sw.UpstreamSwitchId = upstreamSwitchId
	sw.Build()

	return sw, nil
}

func (s *pgSwitches) MoveSwitch(ctx context.Context, userEmail string, switchId int, displayOrder int) ([]models.Switch, error) {
	query := `SELECT ` + schema.FnMoveSwitch_OUTPUTS + ` FROM ` + schema.FnMoveSwitch + `($1, $2, $3)`

	rows, err := s.pool.Query(ctx, query, userEmail, switchId, displayOrder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switches := make([]models.Switch, 0)
	for rows.Next() {
		var sw models.Switch
		if err := rows.Scan(&sw.Id, &sw.Name, &sw.SwitchPanelId, &sw.SpaceGroupId, &sw.DisplayOrder, &sw.UpstreamSwitchId); err != nil {
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

func (s *pgSwitches) SetSwitch(ctx context.Context, userEmail string, switchId int, name *string) (*models.Switch, error) {
	query := `SELECT ` + schema.FnSetSwitch_OUTPUTS + ` FROM ` + schema.FnSetSwitch + `($1, $2, $3)`

	sw := &models.Switch{}
	err := s.pool.QueryRow(ctx, query, userEmail, switchId, name).
		Scan(&sw.Id, &sw.Name, &sw.SwitchPanelId, &sw.SpaceGroupId, &sw.DisplayOrder, &sw.UpstreamSwitchId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	sw.Build()

	return sw, nil
}

func (s *pgSwitches) SetSwitchUpstream(ctx context.Context, userEmail string, switchId int, upstreamSwitchId *int) (*models.Switch, error) {
	query := `SELECT ` + schema.FnSetSwitchUpstream_OUTPUTS + ` FROM ` + schema.FnSetSwitchUpstream + `($1, $2, $3)`

	sw := &models.Switch{}
	err := s.pool.QueryRow(ctx, query, userEmail, switchId, upstreamSwitchId).
		Scan(&sw.Id, &sw.Name, &sw.SwitchPanelId, &sw.SpaceGroupId, &sw.DisplayOrder, &sw.UpstreamSwitchId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case schema.FnSetSwitchUpstream_ERR_SELF_REFERENCE:
			return nil, repository.ErrSelfReference
		case schema.FnSetSwitchUpstream_ERR_CYCLE_DETECTED:
			return nil, repository.ErrCycleDetected
		}
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

func (s *pgSwitches) GetSwitchDownstream(ctx context.Context, userEmail string, switchId int) ([]models.Switch, []models.Device, error) {
	_, fields, err := models.ParseSwitchColumns("")
	if err != nil {
		return nil, nil, err
	}

	query := `
		WITH RECURSIVE switch_chain AS (
			SELECT ` + schema.FullSwitches_ID + `, ` + schema.FullSwitches_NAME + `, ` + schema.FullSwitches_SWITCH_PANEL_ID + `, ` + schema.FullSwitches_SPACE_GROUP_ID + `, ` + schema.FullSwitches_DISPLAY_ORDER + `, ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + `, 0 AS depth
			FROM ` + schema.SwitchesTable + `
			WHERE ` + schema.FullSwitches_ID + ` = $1 AND EXISTS (
				SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
				WHERE ` + schema.JoinSpaceGroupOwners_Switches + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			)

			UNION ALL

			SELECT ` + schema.FullSwitches_ID + `, ` + schema.FullSwitches_NAME + `, ` + schema.FullSwitches_SWITCH_PANEL_ID + `, ` + schema.FullSwitches_SPACE_GROUP_ID + `, ` + schema.FullSwitches_DISPLAY_ORDER + `, ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + `, switch_chain.depth + 1
			FROM ` + schema.SwitchesTable + `
			JOIN switch_chain ON ` + schema.FullSwitches_UPSTREAM_SWITCH_ID + ` = switch_chain.` + schema.Switches_ID + `
		)
		SELECT ` + schema.Switches_ID + `, ` + schema.Switches_NAME + `, ` + schema.Switches_SWITCH_PANEL_ID + `, ` + schema.Switches_SPACE_GROUP_ID + `, ` + schema.Switches_DISPLAY_ORDER + `, ` + schema.Switches_UPSTREAM_SWITCH_ID + `
		FROM switch_chain
		ORDER BY depth
	`
	rows, err := s.pool.Query(ctx, query, switchId, userEmail)
	if err != nil {
		return nil, nil, err
	}

	switches := make([]models.Switch, 0)
	ids := make([]int, 0)
	for rows.Next() {
		var sw models.Switch
		if err := rows.Scan(sw.Values(fields)...); err != nil {
			rows.Close()
			return nil, nil, err
		}
		sw.Build()
		switches = append(switches, sw)
		ids = append(ids, sw.Id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(switches) == 0 {
		return nil, nil, repository.ErrNotFound
	}

	deviceQuery := `
		SELECT ` + schema.FullDevices_ID + `, ` + schema.FullDevices_NAME + `, ` + schema.FullDevices_SPACE_ID + `, ` + schema.FullDeviceSwitches_SWITCH_ID + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.DeviceSwitchesTable + ` ON ` + schema.JoinDeviceSwitches_Devices + `
		WHERE ` + schema.FullDeviceSwitches_SWITCH_ID + ` = ANY($1)
		ORDER BY ` + schema.FullDeviceSwitches_SWITCH_ID + `, ` + schema.FullDevices_ID + `
	`
	deviceRows, err := s.pool.Query(ctx, deviceQuery, ids)
	if err != nil {
		return nil, nil, err
	}
	defer deviceRows.Close()

	devices := make([]models.Device, 0)
	for deviceRows.Next() {
		var d models.Device
		if err := deviceRows.Scan(&d.Id, &d.Name, &d.SpaceId, &d.SwitchId); err != nil {
			return nil, nil, err
		}
		d.Build()
		devices = append(devices, d)
	}

	return switches, devices, deviceRows.Err()
}
