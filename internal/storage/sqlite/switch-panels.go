package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"database/sql"
)

type sqliteSwitchPanels struct {
	db *sql.DB
}

func (panel *sqliteSwitchPanels) GetSwitchPanels(ctx context.Context, _ string, spaceId string, include string) ([]models.SwitchPanel, error) {
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

func (panel *sqliteSwitchPanels) GetSwitchPanel(ctx context.Context, _ string, switchPanelId string, include string) (*models.SwitchPanel, error) {
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
	if err := panel.db.QueryRowContext(ctx, query, switchPanelId).Scan(group.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}
	group.Build()

	return &group, nil
}

func (panel *sqliteSwitchPanels) AddSwitchPanel(ctx context.Context, _ string, spaceId string, name string, position *models.Position, attributes models.Attributes) (*models.SwitchPanel, error) {
	group := &models.SwitchPanel{Name: name, SpaceId: spaceId, Position: position, Attributes: attributes}

	err := withTx(ctx, panel.db, func(tx *sql.Tx) error {
		spaceGroup := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, spaceId).Scan(&group.SpaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		insertGroup := `
			INSERT INTO ` + schema.SwitchPanelsTable + ` (` + schema.SwitchPanels_NAME + `, ` + schema.SwitchPanels_SPACE_ID + `, ` + schema.SwitchPanels_SPACE_GROUP_ID + `, ` + schema.SwitchPanels_ATTRIBUTES + `)
			VALUES ($1, $2, $3, $4)
			RETURNING ` + schema.SwitchPanels_ID
		if err := tx.QueryRowContext(ctx, insertGroup, name, spaceId, group.SpaceGroupId, attributes).Scan(&group.Id); err != nil {
			return err
		}

		if position == nil {
			return nil
		}
		return upsertSwitchPanelPosition(ctx, tx, group.Id, *position)
	})
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *sqliteSwitchPanels) SetSwitchPanel(ctx context.Context, _ string, switchPanelId string, name *string, attributes models.Attributes) (*models.SwitchPanel, error) {
	var group *models.SwitchPanel

	err := withTx(ctx, panel.db, func(tx *sql.Tx) error {
		update := `
			UPDATE ` + schema.SwitchPanelsTable + `
			SET ` + schema.SwitchPanels_NAME + ` = COALESCE($1, ` + schema.SwitchPanels_NAME + `),
				` + schema.SwitchPanels_ATTRIBUTES + ` = COALESCE($2, ` + schema.SwitchPanels_ATTRIBUTES + `)
			WHERE ` + schema.SwitchPanels_ID + ` = $3
		`
		if err := notFoundIfNoneAffected(tx.ExecContext(ctx, update, name, attributes, switchPanelId)); err != nil {
			return err
		}

		var err error
		group, err = selectSwitchPanelWithPosition(ctx, tx, switchPanelId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (panel *sqliteSwitchPanels) SetSwitchPanelPosition(ctx context.Context, _ string, switchPanelId string, position models.Position) (*models.SwitchPanel, error) {
	var group *models.SwitchPanel

	err := withTx(ctx, panel.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SwitchPanelsTable, schema.SwitchPanels_ID, switchPanelId); err != nil {
			return err
		}

		if err := upsertSwitchPanelPosition(ctx, tx, switchPanelId, position); err != nil {
			return err
		}

		var err error
		group, err = selectSwitchPanelWithPosition(ctx, tx, switchPanelId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (panel *sqliteSwitchPanels) DelSwitchPanel(ctx context.Context, _ string, switchPanelId string) error {
	query := `DELETE FROM ` + schema.SwitchPanelsTable + ` WHERE ` + schema.SwitchPanels_ID + ` = $1`
	return notFoundIfNoneAffected(panel.db.ExecContext(ctx, query, switchPanelId))
}

func upsertSwitchPanelPosition(ctx context.Context, q querier, switchPanelId string, position models.Position) error {
	query := `
		INSERT INTO ` + schema.SwitchPanelPositionsTable + ` (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `, ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `) DO UPDATE
		SET ` + schema.SwitchPanelPositions_POSITION_X + ` = excluded.` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + ` = excluded.` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + ` = excluded.` + schema.SwitchPanelPositions_POSITION_Z
	_, err := q.ExecContext(ctx, query, switchPanelId, position.X, position.Y, position.Z)
	return err
}

func selectSwitchPanelWithPosition(ctx context.Context, q querier, switchPanelId string) (*models.SwitchPanel, error) {
	query := `
		SELECT ` + models.SWITCH_PANEL_WITH_POSITION_COLUMNS + `
		FROM ` + schema.SwitchPanelsTable + `
		LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels + `
		WHERE ` + schema.FullSwitchPanels_ID + ` = $1
	`

	group := &models.SwitchPanel{}
	err := q.QueryRowContext(ctx, query, switchPanelId).Scan(group.ValuesWithPosition()...)
	if err != nil {
		return nil, notFoundIfNoRows(err)
	}
	group.Build()

	return group, nil
}
