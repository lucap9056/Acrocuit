package storage

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BreakerGroups interface {
	GetBreakerGroups(ctx context.Context, userId, spaceId int, include string) ([]models.BreakerGroup, error)
	GetBreakerGroup(ctx context.Context, userId, breakerGroupId int, include string) (*models.BreakerGroup, error)
	AddBreakerGroup(ctx context.Context, userId, spaceId int, name string, position *models.Position) (*models.BreakerGroup, error)
	SetBreakerGroup(ctx context.Context, userId, breakerGroupId int, name string) (*models.BreakerGroup, error)
	SetBreakerGroupPosition(ctx context.Context, userId, breakerGroupId int, position models.Position) (*models.BreakerGroup, error)
	DelBreakerGroup(ctx context.Context, userId, breakerGroupId int) error
}

type pgBreakerGroups struct {
	pool *pgxpool.Pool
}

func (bg *pgBreakerGroups) GetBreakerGroups(ctx context.Context, userId, spaceId int, include string) ([]models.BreakerGroup, error) {
	keys, fields, includePosition, err := models.ParseBreakerGroupColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.BreakerGroupsTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinBreakerGroups_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.BreakerGroupPositionsTable + ` ON ` + schema.JoinBreakerGroupPositions_BreakerGroups
	}
	query += ` WHERE ` + schema.FullBreakerGroups_SPACE_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1`

	rows, err := bg.pool.Query(ctx, query, userId, spaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakerGroups := make([]models.BreakerGroup, 0)
	for rows.Next() {
		var group models.BreakerGroup
		if err := rows.Scan(group.Values(fields)...); err != nil {
			return nil, err
		}
		group.Build()
		breakerGroups = append(breakerGroups, group)
	}

	return breakerGroups, rows.Err()
}

func (bg *pgBreakerGroups) GetBreakerGroup(ctx context.Context, userId, breakerGroupId int, include string) (*models.BreakerGroup, error) {
	keys, fields, includePosition, err := models.ParseBreakerGroupColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.BreakerGroupsTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinBreakerGroups_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces
	if includePosition {
		query += ` LEFT JOIN ` + schema.BreakerGroupPositionsTable + ` ON ` + schema.JoinBreakerGroupPositions_BreakerGroups
	}
	query += ` WHERE ` + schema.FullBreakerGroups_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1`

	var group models.BreakerGroup
	err = bg.pool.QueryRow(ctx, query, userId, breakerGroupId).Scan(group.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	group.Build()

	return &group, nil
}

func (bg *pgBreakerGroups) AddBreakerGroup(ctx context.Context, userId, spaceId int, name string, position *models.Position) (*models.BreakerGroup, error) {
	var posX, posY, posZ *int
	if position != nil {
		posX, posY, posZ = &position.X, &position.Y, &position.Z
	}

	query := `
		WITH permitted AS (
			SELECT ` + schema.FullSpaces_SPACE_GROUP_ID + `
			FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullSpaces_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		inserted_group AS (
			INSERT INTO ` + schema.BreakerGroupsTable + ` (` + schema.BreakerGroups_NAME + `, ` + schema.BreakerGroups_SPACE_ID + `, ` + schema.BreakerGroups_SPACE_GROUP_ID + `)
			SELECT $3, $1, ` + schema.BreakerGroups_SPACE_GROUP_ID + ` FROM permitted
			RETURNING ` + schema.BreakerGroups_ID + `, ` + schema.BreakerGroups_SPACE_GROUP_ID + `
		),
		inserted_position AS (
			INSERT INTO ` + schema.BreakerGroupPositionsTable + ` (` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + `, ` + schema.BreakerGroupPositions_POSITION_X + `, ` + schema.BreakerGroupPositions_POSITION_Y + `, ` + schema.BreakerGroupPositions_POSITION_Z + `)
			SELECT ` + schema.BreakerGroups_ID + `, $4, $5, $6
			FROM inserted_group
			WHERE $4::integer IS NOT NULL
			RETURNING ` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + `
		)
		SELECT ` + schema.BreakerGroups_ID + `, ` + schema.BreakerGroups_SPACE_GROUP_ID + ` FROM inserted_group
	`

	group := &models.BreakerGroup{Name: name, SpaceId: spaceId, Position: position}

	err := bg.pool.QueryRow(ctx, query, spaceId, userId, name, posX, posY, posZ).Scan(&group.Id, &group.SpaceGroupId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (bg *pgBreakerGroups) SetBreakerGroup(ctx context.Context, userId, breakerGroupId int, name string) (*models.BreakerGroup, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullBreakerGroups_ID + `
			FROM ` + schema.BreakerGroupsTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinBreakerGroups_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullBreakerGroups_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		permitted_update AS (
			UPDATE ` + schema.BreakerGroupsTable + `
			SET ` + schema.BreakerGroups_NAME + ` = $3
			WHERE ` + schema.BreakerGroups_ID + ` = $1 AND EXISTS (SELECT 1 FROM permitted)
			RETURNING ` + schema.FullBreakerGroups_ID + `, ` + schema.FullBreakerGroups_NAME + `, ` + schema.FullBreakerGroups_SPACE_ID + `, ` + schema.FullBreakerGroups_SPACE_GROUP_ID + `
		)
		SELECT
			permitted_update.` + schema.BreakerGroups_ID + `,
			permitted_update.` + schema.BreakerGroups_NAME + `,
			permitted_update.` + schema.BreakerGroups_SPACE_ID + `,
			permitted_update.` + schema.BreakerGroups_SPACE_GROUP_ID + `,
			p.` + schema.BreakerGroupPositions_POSITION_X + `,
			p.` + schema.BreakerGroupPositions_POSITION_Y + `,
			p.` + schema.BreakerGroupPositions_POSITION_Z + `
		FROM permitted_update
		LEFT JOIN ` + schema.BreakerGroupPositionsTable + ` p ON p.` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + ` = permitted_update.` + schema.BreakerGroups_ID + `
	`

	group := &models.BreakerGroup{Id: breakerGroupId}

	err := bg.pool.QueryRow(ctx, query, breakerGroupId, userId, name).
		Scan(&group.Id, &group.Name, &group.SpaceId, &group.SpaceGroupId, &group.PositionX, &group.PositionY, &group.PositionZ)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (bg *pgBreakerGroups) SetBreakerGroupPosition(ctx context.Context, userId, breakerGroupId int, position models.Position) (*models.BreakerGroup, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullBreakerGroups_ID + `
			FROM ` + schema.BreakerGroupsTable + `
			JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinBreakerGroups_Spaces + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.FullBreakerGroups_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
			FOR UPDATE OF ` + schema.SpacesTable + `
		),
		group_row AS (
			SELECT ` + schema.FullBreakerGroups_ID + ` AS id, ` + schema.FullBreakerGroups_NAME + ` AS name, ` + schema.FullBreakerGroups_SPACE_ID + ` AS space_id, ` + schema.FullBreakerGroups_SPACE_GROUP_ID + ` AS space_group_id
			FROM ` + schema.BreakerGroupsTable + `
			WHERE ` + schema.BreakerGroups_ID + ` = $1 AND EXISTS (SELECT 1 FROM permitted)
		),
		upserted_position AS (
			INSERT INTO ` + schema.BreakerGroupPositionsTable + ` (` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + `, ` + schema.BreakerGroupPositions_POSITION_X + `, ` + schema.BreakerGroupPositions_POSITION_Y + `, ` + schema.BreakerGroupPositions_POSITION_Z + `)
			SELECT id, $3, $4, $5
			FROM group_row
			ON CONFLICT (` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + `) DO UPDATE
			SET ` + schema.BreakerGroupPositions_POSITION_X + ` = EXCLUDED.` + schema.BreakerGroupPositions_POSITION_X + `, ` + schema.BreakerGroupPositions_POSITION_Y + ` = EXCLUDED.` + schema.BreakerGroupPositions_POSITION_Y + `, ` + schema.BreakerGroupPositions_POSITION_Z + ` = EXCLUDED.` + schema.BreakerGroupPositions_POSITION_Z + `
			RETURNING ` + schema.BreakerGroupPositions_POSITION_X + `, ` + schema.BreakerGroupPositions_POSITION_Y + `, ` + schema.BreakerGroupPositions_POSITION_Z + `
		)
		SELECT
			group_row.id,
			group_row.name,
			group_row.space_id,
			group_row.space_group_id,
			upserted_position.` + schema.BreakerGroupPositions_POSITION_X + `,
			upserted_position.` + schema.BreakerGroupPositions_POSITION_Y + `,
			upserted_position.` + schema.BreakerGroupPositions_POSITION_Z + `
		FROM group_row
		LEFT JOIN upserted_position ON true
	`

	group := &models.BreakerGroup{Id: breakerGroupId}

	err := bg.pool.QueryRow(ctx, query, breakerGroupId, userId, position.X, position.Y, position.Z).
		Scan(&group.Id, &group.Name, &group.SpaceId, &group.SpaceGroupId, &group.PositionX, &group.PositionY, &group.PositionZ)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (bg *pgBreakerGroups) DelBreakerGroup(ctx context.Context, userId, breakerGroupId int) error {
	query := `
		DELETE FROM ` + schema.BreakerGroupsTable + `
		WHERE ` + schema.FullBreakerGroups_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.JoinBreakerGroups_Spaces + ` AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
		)
	`
	t, err := bg.pool.Exec(ctx, query, breakerGroupId, userId)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
