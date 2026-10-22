package postgres

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgSwitchPanels struct {
	pool *pgxpool.Pool
}

func (panel *pgSwitchPanels) GetSwitchPanels(ctx context.Context, userEmail string, spaceId int, include string) ([]models.SwitchPanel, error) {
	keys, fields, includePosition, err := models.ParseSwitchPanelColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchPanelsTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSwitchPanels_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels
	}
	query += ` WHERE ` + schema.FullSwitchPanels_SPACE_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1`

	rows, err := panel.pool.Query(ctx, query, userEmail, spaceId)
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

func (panel *pgSwitchPanels) GetSwitchPanel(ctx context.Context, userEmail string, switchPanelId int, include string) (*models.SwitchPanel, error) {
	keys, fields, includePosition, err := models.ParseSwitchPanelColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SwitchPanelsTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSwitchPanels_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels
	}
	query += ` WHERE ` + schema.FullSwitchPanels_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1`

	var group models.SwitchPanel
	err = panel.pool.QueryRow(ctx, query, userEmail, switchPanelId).Scan(group.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	group.Build()

	return &group, nil
}

func (panel *pgSwitchPanels) AddSwitchPanel(ctx context.Context, userEmail string, spaceId int, name string, position *models.Position) (*models.SwitchPanel, error) {
	var posX, posY, posZ *int
	if position != nil {
		posX, posY, posZ = &position.X, &position.Y, &position.Z
	}

	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSpaces_SPACE_GROUP_ID + `
			FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullSpaces_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		inserted_group AS (
			INSERT INTO ` + schema.SwitchPanelsTable + ` (` + schema.SwitchPanels_NAME + `, ` + schema.SwitchPanels_SPACE_ID + `, ` + schema.SwitchPanels_SPACE_GROUP_ID + `)
			SELECT $3, $1, ` + schema.SwitchPanels_SPACE_GROUP_ID + ` FROM permitted
			RETURNING ` + schema.SwitchPanels_ID + `, ` + schema.SwitchPanels_SPACE_GROUP_ID + `
		),
		inserted_position AS (
			INSERT INTO ` + schema.SwitchPanelPositionsTable + ` (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `, ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `)
			SELECT ` + schema.SwitchPanels_ID + `, $4, $5, $6
			FROM inserted_group
			WHERE $4::integer IS NOT NULL
			RETURNING ` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `
		)
		SELECT ` + schema.SwitchPanels_ID + `, ` + schema.SwitchPanels_SPACE_GROUP_ID + ` FROM inserted_group
	`

	group := &models.SwitchPanel{Name: name, SpaceId: spaceId, Position: position}

	err := panel.pool.QueryRow(ctx, query, spaceId, userEmail, name, posX, posY, posZ).Scan(&group.Id, &group.SpaceGroupId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *pgSwitchPanels) SetSwitchPanel(ctx context.Context, userEmail string, switchPanelId int, name string) (*models.SwitchPanel, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSwitchPanels_ID + `
			FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSwitchPanels_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullSwitchPanels_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		permitted_update AS (
			UPDATE ` + schema.SwitchPanelsTable + `
			SET ` + schema.SwitchPanels_NAME + ` = $3
			WHERE ` + schema.SwitchPanels_ID + ` = $1 AND EXISTS (SELECT 1 FROM permitted)
			RETURNING ` + schema.FullSwitchPanels_ID + `, ` + schema.FullSwitchPanels_NAME + `, ` + schema.FullSwitchPanels_SPACE_ID + `, ` + schema.FullSwitchPanels_SPACE_GROUP_ID + `
		)
		SELECT
			permitted_update.` + schema.SwitchPanels_ID + `,
			permitted_update.` + schema.SwitchPanels_NAME + `,
			permitted_update.` + schema.SwitchPanels_SPACE_ID + `,
			permitted_update.` + schema.SwitchPanels_SPACE_GROUP_ID + `,
			p.` + schema.SwitchPanelPositions_POSITION_X + `,
			p.` + schema.SwitchPanelPositions_POSITION_Y + `,
			p.` + schema.SwitchPanelPositions_POSITION_Z + `
		FROM permitted_update
		LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` p ON p.` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + ` = permitted_update.` + schema.SwitchPanels_ID + `
	`

	group := &models.SwitchPanel{Id: switchPanelId}

	err := panel.pool.QueryRow(ctx, query, switchPanelId, userEmail, name).
		Scan(&group.Id, &group.Name, &group.SpaceId, &group.SpaceGroupId, &group.PositionX, &group.PositionY, &group.PositionZ)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *pgSwitchPanels) SetSwitchPanelPosition(ctx context.Context, userEmail string, switchPanelId int, position models.Position) (*models.SwitchPanel, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSwitchPanels_ID + `
			FROM ` + schema.SwitchPanelsTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSwitchPanels_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullSwitchPanels_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		group_row AS (
			SELECT ` + schema.FullSwitchPanels_ID + ` AS id, ` + schema.FullSwitchPanels_NAME + ` AS name, ` + schema.FullSwitchPanels_SPACE_ID + ` AS space_id, ` + schema.FullSwitchPanels_SPACE_GROUP_ID + ` AS space_group_id
			FROM ` + schema.SwitchPanelsTable + `
			WHERE ` + schema.SwitchPanels_ID + ` = $1 AND EXISTS (SELECT 1 FROM permitted)
		),
		upserted_position AS (
			INSERT INTO ` + schema.SwitchPanelPositionsTable + ` (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `, ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `)
			SELECT id, $3, $4, $5
			FROM group_row
			ON CONFLICT (` + schema.SwitchPanelPositions_SWITCH_PANEL_ID + `) DO UPDATE
			SET ` + schema.SwitchPanelPositions_POSITION_X + ` = EXCLUDED.` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + ` = EXCLUDED.` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + ` = EXCLUDED.` + schema.SwitchPanelPositions_POSITION_Z + `
			RETURNING ` + schema.SwitchPanelPositions_POSITION_X + `, ` + schema.SwitchPanelPositions_POSITION_Y + `, ` + schema.SwitchPanelPositions_POSITION_Z + `
		)
		SELECT
			group_row.id,
			group_row.name,
			group_row.space_id,
			group_row.space_group_id,
			upserted_position.` + schema.SwitchPanelPositions_POSITION_X + `,
			upserted_position.` + schema.SwitchPanelPositions_POSITION_Y + `,
			upserted_position.` + schema.SwitchPanelPositions_POSITION_Z + `
		FROM group_row
		LEFT JOIN upserted_position ON true
	`

	group := &models.SwitchPanel{Id: switchPanelId}

	err := panel.pool.QueryRow(ctx, query, switchPanelId, userEmail, position.X, position.Y, position.Z).
		Scan(&group.Id, &group.Name, &group.SpaceId, &group.SpaceGroupId, &group.PositionX, &group.PositionY, &group.PositionZ)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (panel *pgSwitchPanels) DelSwitchPanel(ctx context.Context, userEmail string, switchPanelId int) error {
	query := `
		DELETE FROM ` + schema.SwitchPanelsTable + `
		WHERE ` + schema.FullSwitchPanels_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.JoinSwitchPanels_Spaces + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	t, err := panel.pool.Exec(ctx, query, switchPanelId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
