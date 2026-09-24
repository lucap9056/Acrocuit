package storage

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpaceGroups interface {
	GetSpaceGroups(ctx context.Context, userId int, include string) ([]models.SpaceGroup, error)
	GetSpaceGroup(ctx context.Context, userId, spaceGroupId int, include string) (*models.SpaceGroup, error)
	AddSpaceGroup(ctx context.Context, userId int, name string) (int, error)
	EditSpaceGroupName(ctx context.Context, userId, spaceGroupId int, name string) error
	DelSpaceGroup(ctx context.Context, userId, spaceGroupId int) error
}

type pgSpaceGroups struct {
	pool *pgxpool.Pool
}

func (s *pgSpaceGroups) GetSpaceGroups(ctx context.Context, userId int, include string) ([]models.SpaceGroup, error) {
	keys, fields, err := models.ParseSpaceGroupColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpaceGroupsTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SpaceGroups + `
		WHERE ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1`

	rows, err := s.pool.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spaceGroups := make([]models.SpaceGroup, 0)
	for rows.Next() {
		var sg models.SpaceGroup
		if err := rows.Scan(sg.Values(fields)...); err != nil {
			return nil, err
		}
		spaceGroups = append(spaceGroups, sg)
	}

	return spaceGroups, rows.Err()
}

func (s *pgSpaceGroups) AddSpaceGroup(ctx context.Context, userId int, name string) (int, error) {
	query := `
		WITH new_group AS (
			INSERT INTO ` + schema.SpaceGroupsTable + ` (` + schema.SpaceGroups_NAME + `) VALUES ($2) RETURNING ` + schema.SpaceGroups_ID + `
		)
		INSERT INTO ` + schema.SpaceGroupOwnersTable + ` (` + schema.SpaceGroupOwners_USER_ID + `, ` + schema.SpaceGroupOwners_SPACE_GROUP_ID + `)
		SELECT $1, ` + schema.SpaceGroups_ID + ` FROM new_group
		RETURNING ` + schema.SpaceGroupOwners_SPACE_GROUP_ID + `
	`
	var spaceGroupId int
	if err := s.pool.QueryRow(ctx, query, userId, name).Scan(&spaceGroupId); err != nil {
		return 0, err
	}

	return spaceGroupId, nil
}

func (s *pgSpaceGroups) GetSpaceGroup(ctx context.Context, userId, spaceGroupId int, include string) (*models.SpaceGroup, error) {
	keys, fields, err := models.ParseSpaceGroupColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpaceGroupsTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SpaceGroups + `
		WHERE ` + schema.FullSpaceGroups_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1`

	sg := &models.SpaceGroup{}
	err = s.pool.QueryRow(ctx, query, userId, spaceGroupId).Scan(sg.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return sg, nil
}

func (s *pgSpaceGroups) EditSpaceGroupName(ctx context.Context, userId, spaceGroupId int, name string) error {
	query := `
		UPDATE ` + schema.SpaceGroupsTable + `
		SET ` + schema.SpaceGroups_NAME + ` = $1
		WHERE ` + schema.FullSpaceGroups_ID + ` = $2 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
			WHERE ` + schema.JoinSpaceGroupOwners_SpaceGroups + ` AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $3
		)
	`
	t, err := s.pool.Exec(ctx, query, name, spaceGroupId, userId)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *pgSpaceGroups) DelSpaceGroup(ctx context.Context, userId, spaceGroupId int) error {
	query := `
		DELETE FROM ` + schema.SpaceGroupsTable + `
		WHERE ` + schema.FullSpaceGroups_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
			WHERE ` + schema.JoinSpaceGroupOwners_SpaceGroups + ` AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
		)
	`
	t, err := s.pool.Exec(ctx, query, spaceGroupId, userId)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
