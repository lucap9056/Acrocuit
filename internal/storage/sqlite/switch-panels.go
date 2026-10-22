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

func (panel *sqliteSwitchPanels) GetSwitchPanels(ctx context.Context, _ string, spaceId int, include string) ([]models.SwitchPanel, error) {
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

func (panel *sqliteSwitchPanels) GetSwitchPanel(ctx context.Context, _ string, switchPanelId int, include string) (*models.SwitchPanel, error) {
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

func (panel *sqliteSwitchPanels) AddSwitchPanel(ctx context.Context, _ string, spaceId int, name string, position *models.Position) (*models.SwitchPanel, error) {
	group := &models.SwitchPanel{Name: name, SpaceId: spaceId, Position: position}

	err := withTx(ctx, panel.db, func(tx *sql.Tx) error {
		spaceGroup := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, spaceId).Scan(&group.SpaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		insertGroup := `
			INSERT INTO ` + schema.SwitchPanelsTable + ` (` + schema.SwitchPanels_NAME + `, ` + schema.SwitchPanels_SPACE_ID + `, ` + schema.SwitchPanels_SPACE_GROUP_ID + `)
			VALUES ($1, $2, $3)
			RETURNING ` + schema.SwitchPanels_ID
		if err := tx.QueryRowContext(ctx, insertGroup, name, spaceId, group.SpaceGroupId).Scan(&group.Id); err != nil {
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

func (panel *sqliteSwitchPanels) SetSwitchPanel(ctx context.Context, _ string, switchPanelId int, name string) (*models.SwitchPanel, error) {
	var group *models.SwitchPanel

	err := withTx(ctx, panel.db, func(tx *sql.Tx) error {
		rename := `UPDATE ` + schema.SwitchPanelsTable + ` SET ` + schema.SwitchPanels_NAME + ` = $1 WHERE ` + schema.SwitchPanels_ID + ` = $2`
		if err := notFoundIfNoneAffected(tx.ExecContext(ctx, rename, name, switchPanelId)); err != nil {
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

func (panel *sqliteSwitchPanels) SetSwitchPanelPosition(ctx context.Context, _ string, switchPanelId int, position models.Position) (*models.SwitchPanel, error) {
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

func (panel *sqliteSwitchPanels) DelSwitchPanel(ctx context.Context, _ string, switchPanelId int) error {
	query := `DELETE FROM ` + schema.SwitchPanelsTable + ` WHERE ` + schema.SwitchPanels_ID + ` = $1`
	return notFoundIfNoneAffected(panel.db.ExecContext(ctx, query, switchPanelId))
}

func upsertSwitchPanelPosition(ctx context.Context, q querier, switchPanelId int, position models.Position) error {
	query := `
		INSERT INTO ` + schema.SwitchPanelPositionsTable + ` (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `, ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `) DO UPDATE
		SET ` + schema.SwitchPanelPositions_POSITION_X + ` = excluded.` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + ` = excluded.` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + ` = excluded.` + schema.SwitchPanelPositions_POSITION_Z
	_, err := q.ExecContext(ctx, query, switchPanelId, position.X, position.Y, position.Z)
	return err
}

func selectSwitchPanelWithPosition(ctx context.Context, q querier, switchPanelId int) (*models.SwitchPanel, error) {
	query := `
		SELECT ` + schema.FullSwitchPanels_ID + `, ` + schema.FullSwitchPanels_NAME + `, ` + schema.FullSwitchPanels_SPACE_ID + `, ` + schema.FullSwitchPanels_SPACE_GROUP_ID + `,
			` + schema.FullSwitchPanelPositions_POSITION_X + `, ` + schema.FullSwitchPanelPositions_POSITION_Y + `, ` + schema.FullSwitchPanelPositions_POSITION_Z + `
		FROM ` + schema.SwitchPanelsTable + `
		LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels + `
		WHERE ` + schema.FullSwitchPanels_ID + ` = $1
	`

	group := &models.SwitchPanel{}
	err := q.QueryRowContext(ctx, query, switchPanelId).
		Scan(&group.Id, &group.Name, &group.SpaceId, &group.SpaceGroupId, &group.PositionX, &group.PositionY, &group.PositionZ)
	if err != nil {
		return nil, notFoundIfNoRows(err)
	}
	group.Build()

	return group, nil
}
