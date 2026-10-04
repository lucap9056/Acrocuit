package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"database/sql"
)

type sqliteBreakerGroups struct {
	db *sql.DB
}

func (bg *sqliteBreakerGroups) GetBreakerGroups(ctx context.Context, _ string, spaceId int, include string) ([]models.BreakerGroup, error) {
	keys, fields, includePosition, err := models.ParseBreakerGroupColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.BreakerGroupsTable
	if includePosition {
		query += ` LEFT JOIN ` + schema.BreakerGroupPositionsTable + ` ON ` + schema.JoinBreakerGroupPositions_BreakerGroups
	}
	query += ` WHERE ` + schema.FullBreakerGroups_SPACE_ID + ` = $1`

	rows, err := bg.db.QueryContext(ctx, query, spaceId)
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

func (bg *sqliteBreakerGroups) GetBreakerGroup(ctx context.Context, _ string, breakerGroupId int, include string) (*models.BreakerGroup, error) {
	keys, fields, includePosition, err := models.ParseBreakerGroupColumns(include, true)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.BreakerGroupsTable
	if includePosition {
		query += ` LEFT JOIN ` + schema.BreakerGroupPositionsTable + ` ON ` + schema.JoinBreakerGroupPositions_BreakerGroups
	}
	query += ` WHERE ` + schema.FullBreakerGroups_ID + ` = $1`

	var group models.BreakerGroup
	if err := bg.db.QueryRowContext(ctx, query, breakerGroupId).Scan(group.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}
	group.Build()

	return &group, nil
}

func (bg *sqliteBreakerGroups) AddBreakerGroup(ctx context.Context, _ string, spaceId int, name string, position *models.Position) (*models.BreakerGroup, error) {
	group := &models.BreakerGroup{Name: name, SpaceId: spaceId, Position: position}

	err := withTx(ctx, bg.db, func(tx *sql.Tx) error {
		spaceGroup := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, spaceGroup, spaceId).Scan(&group.SpaceGroupId); err != nil {
			return notFoundIfNoRows(err)
		}

		insertGroup := `
			INSERT INTO ` + schema.BreakerGroupsTable + ` (` + schema.BreakerGroups_NAME + `, ` + schema.BreakerGroups_SPACE_ID + `, ` + schema.BreakerGroups_SPACE_GROUP_ID + `)
			VALUES ($1, $2, $3)
			RETURNING ` + schema.BreakerGroups_ID
		if err := tx.QueryRowContext(ctx, insertGroup, name, spaceId, group.SpaceGroupId).Scan(&group.Id); err != nil {
			return err
		}

		if position == nil {
			return nil
		}
		return upsertBreakerGroupPosition(ctx, tx, group.Id, *position)
	})
	if err != nil {
		return nil, err
	}

	group.Build()

	return group, nil
}

func (bg *sqliteBreakerGroups) SetBreakerGroup(ctx context.Context, _ string, breakerGroupId int, name string) (*models.BreakerGroup, error) {
	var group *models.BreakerGroup

	err := withTx(ctx, bg.db, func(tx *sql.Tx) error {
		rename := `UPDATE ` + schema.BreakerGroupsTable + ` SET ` + schema.BreakerGroups_NAME + ` = $1 WHERE ` + schema.BreakerGroups_ID + ` = $2`
		if err := notFoundIfNoneAffected(tx.ExecContext(ctx, rename, name, breakerGroupId)); err != nil {
			return err
		}

		var err error
		group, err = selectBreakerGroupWithPosition(ctx, tx, breakerGroupId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (bg *sqliteBreakerGroups) SetBreakerGroupPosition(ctx context.Context, _ string, breakerGroupId int, position models.Position) (*models.BreakerGroup, error) {
	var group *models.BreakerGroup

	err := withTx(ctx, bg.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.BreakerGroupsTable, schema.BreakerGroups_ID, breakerGroupId); err != nil {
			return err
		}

		if err := upsertBreakerGroupPosition(ctx, tx, breakerGroupId, position); err != nil {
			return err
		}

		var err error
		group, err = selectBreakerGroupWithPosition(ctx, tx, breakerGroupId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (bg *sqliteBreakerGroups) DelBreakerGroup(ctx context.Context, _ string, breakerGroupId int) error {
	query := `DELETE FROM ` + schema.BreakerGroupsTable + ` WHERE ` + schema.BreakerGroups_ID + ` = $1`
	return notFoundIfNoneAffected(bg.db.ExecContext(ctx, query, breakerGroupId))
}

func upsertBreakerGroupPosition(ctx context.Context, q querier, breakerGroupId int, position models.Position) error {
	query := `
		INSERT INTO ` + schema.BreakerGroupPositionsTable + ` (` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + `, ` + schema.BreakerGroupPositions_POSITION_X + `, ` + schema.BreakerGroupPositions_POSITION_Y + `, ` + schema.BreakerGroupPositions_POSITION_Z + `)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (` + schema.BreakerGroupPositions_BREAKER_GROUP_ID + `) DO UPDATE
		SET ` + schema.BreakerGroupPositions_POSITION_X + ` = excluded.` + schema.BreakerGroupPositions_POSITION_X + `, ` + schema.BreakerGroupPositions_POSITION_Y + ` = excluded.` + schema.BreakerGroupPositions_POSITION_Y + `, ` + schema.BreakerGroupPositions_POSITION_Z + ` = excluded.` + schema.BreakerGroupPositions_POSITION_Z
	_, err := q.ExecContext(ctx, query, breakerGroupId, position.X, position.Y, position.Z)
	return err
}

func selectBreakerGroupWithPosition(ctx context.Context, q querier, breakerGroupId int) (*models.BreakerGroup, error) {
	query := `
		SELECT ` + schema.FullBreakerGroups_ID + `, ` + schema.FullBreakerGroups_NAME + `, ` + schema.FullBreakerGroups_SPACE_ID + `, ` + schema.FullBreakerGroups_SPACE_GROUP_ID + `,
			` + schema.FullBreakerGroupPositions_POSITION_X + `, ` + schema.FullBreakerGroupPositions_POSITION_Y + `, ` + schema.FullBreakerGroupPositions_POSITION_Z + `
		FROM ` + schema.BreakerGroupsTable + `
		LEFT JOIN ` + schema.BreakerGroupPositionsTable + ` ON ` + schema.JoinBreakerGroupPositions_BreakerGroups + `
		WHERE ` + schema.FullBreakerGroups_ID + ` = $1
	`

	group := &models.BreakerGroup{}
	err := q.QueryRowContext(ctx, query, breakerGroupId).
		Scan(&group.Id, &group.Name, &group.SpaceId, &group.SpaceGroupId, &group.PositionX, &group.PositionY, &group.PositionZ)
	if err != nil {
		return nil, notFoundIfNoRows(err)
	}
	group.Build()

	return group, nil
}
