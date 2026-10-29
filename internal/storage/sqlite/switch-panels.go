package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"errors"
)

type sqliteSwitchPanels struct {
	db *sql.DB
}

func (panel *sqliteSwitchPanels) GetSwitchPanels(ctx context.Context, _ string, spaceId models.SpaceId, include string) ([]models.SwitchPanel, error) {
	keys, fields, includePosition, err := models.ParseSwitchPanelColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SwitchPanelsTable
	if includePosition {
		query += ` LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels
	}
	query += ` WHERE ` + schema.FullSwitchPanels_SPACE_ID + ` = $1`

	rows, err := panel.db.QueryContext(ctx, query, spaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switchPanels := make([]models.SwitchPanel, 0)
	for rows.Next() {
		var group models.SwitchPanel
		if err := rows.Scan(group.Values(fields)...); err != nil {
			return nil, err
		}
		group.Build()
		switchPanels = append(switchPanels, group)
	}

	return switchPanels, rows.Err()
}

func (panel *sqliteSwitchPanels) GetSwitchPanel(ctx context.Context, _ string, switchPanelId models.SwitchPanelId, include string) (*models.SwitchPanel, error) {
	keys, fields, includePosition, err := models.ParseSwitchPanelColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SwitchPanelsTable
	if includePosition {
		query += ` LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels
	}
	query += ` WHERE ` + schema.FullSwitchPanels_ID + ` = $1`

	var group models.SwitchPanel
	err = panel.db.QueryRowContext(ctx, query, switchPanelId).Scan(group.Values(fields)...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	group.Build()

	return &group, nil
}

func (panel *sqliteSwitchPanels) AddSwitchPanel(ctx context.Context, _ string, spaceId models.SpaceId, name string, position *models.Position, attributes models.Attributes) (*models.SwitchPanel, error) {
	tx, err := panel.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	group := &models.SwitchPanel{Name: name, SpaceId: spaceId, Position: position, Attributes: attributes}

	spaceGroup := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
	err = tx.QueryRowContext(ctx, spaceGroup, spaceId).Scan(&group.SpaceGroupId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	insertGroup := `
		INSERT INTO ` + schema.SwitchPanelsTable + ` (` + schema.SwitchPanels_NAME + `, ` + schema.SwitchPanels_SPACE_ID + `, ` + schema.SwitchPanels_SPACE_GROUP_ID + `, ` + schema.SwitchPanels_ATTRIBUTES + `)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + schema.SwitchPanels_ID
	if err := tx.QueryRowContext(ctx, insertGroup, name, spaceId, group.SpaceGroupId, attributes).Scan(&group.Id); err != nil {
		return nil, err
	}

	if position != nil {
		insertPosition := `
			INSERT INTO ` + schema.SwitchPanelPositionsTable + ` (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `, ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `)
			VALUES ($1, $2, $3, $4)
		`
		if _, err := tx.ExecContext(ctx, insertPosition, group.Id, position.X, position.Y, position.Z); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *sqliteSwitchPanels) SetSwitchPanel(ctx context.Context, _ string, switchPanelId models.SwitchPanelId, name *string, attributes models.Attributes) (*models.SwitchPanel, error) {
	tx, err := panel.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	update := `
		UPDATE ` + schema.SwitchPanelsTable + `
		SET ` + schema.SwitchPanels_NAME + ` = COALESCE($1, ` + schema.SwitchPanels_NAME + `),
			` + schema.SwitchPanels_ATTRIBUTES + ` = COALESCE($2, ` + schema.SwitchPanels_ATTRIBUTES + `)
		WHERE ` + schema.SwitchPanels_ID + ` = $3
	`
	result, err := tx.ExecContext(ctx, update, name, attributes, switchPanelId)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, repository.ErrNotFound
	}

	query := `
		SELECT ` + models.SWITCH_PANEL_WITH_POSITION_COLUMNS + `
		FROM ` + schema.SwitchPanelsTable + `
		LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels + `
		WHERE ` + schema.FullSwitchPanels_ID + ` = $1
	`
	group := &models.SwitchPanel{}
	if err := tx.QueryRowContext(ctx, query, switchPanelId).Scan(group.ValuesWithPosition()...); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *sqliteSwitchPanels) SetSwitchPanelPosition(ctx context.Context, _ string, switchPanelId models.SwitchPanelId, position models.Position) (*models.SwitchPanel, error) {
	tx, err := panel.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	target := `SELECT 1 FROM ` + schema.SwitchPanelsTable + ` WHERE ` + schema.SwitchPanels_ID + ` = $1`
	var found int
	err = tx.QueryRowContext(ctx, target, switchPanelId).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	upsert := `
		INSERT INTO ` + schema.SwitchPanelPositionsTable + ` (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `, ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `) DO UPDATE
		SET ` + schema.SwitchPanelPositions_POSITION_X + ` = excluded.` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + ` = excluded.` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + ` = excluded.` + schema.SwitchPanelPositions_POSITION_Z
	if _, err := tx.ExecContext(ctx, upsert, switchPanelId, position.X, position.Y, position.Z); err != nil {
		return nil, err
	}

	query := `
		SELECT ` + models.SWITCH_PANEL_WITH_POSITION_COLUMNS + `
		FROM ` + schema.SwitchPanelsTable + `
		LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels + `
		WHERE ` + schema.FullSwitchPanels_ID + ` = $1
	`
	group := &models.SwitchPanel{}
	if err := tx.QueryRowContext(ctx, query, switchPanelId).Scan(group.ValuesWithPosition()...); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *sqliteSwitchPanels) DelSwitchPanel(ctx context.Context, _ string, switchPanelId models.SwitchPanelId) error {
	query := `DELETE FROM ` + schema.SwitchPanelsTable + ` WHERE ` + schema.SwitchPanels_ID + ` = $1`
	result, err := panel.db.ExecContext(ctx, query, switchPanelId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrNotFound
	}

	return nil
}
