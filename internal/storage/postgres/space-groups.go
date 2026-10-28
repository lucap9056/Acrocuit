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

type pgSpaceGroups struct {
	pool *pgxpool.Pool
}

func (s *pgSpaceGroups) GetSpaceGroups(ctx context.Context, userEmail, include string) ([]models.SpaceGroup, error) {
	if userEmail == models.PublicAccessEmail {
		return nil, repository.ErrAnonymous
	}

	keys, fields, err := models.ParseSpaceGroupColumns(include, schema.FullSpaceGroupACLs_ROLE)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpaceGroupsTable + `
		JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_SpaceGroups + `
		WHERE ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $1`

	rows, err := s.pool.Query(ctx, query, userEmail)
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

func (s *pgSpaceGroups) AddSpaceGroup(ctx context.Context, userEmail, name string) (int, error) {
	if userEmail == models.PublicAccessEmail {
		return 0, repository.ErrAnonymous
	}

	query := `
		WITH new_group AS (
			INSERT INTO ` + schema.SpaceGroupsTable + ` (` + schema.SpaceGroups_NAME + `) VALUES ($1) RETURNING ` + schema.SpaceGroups_ID + `
		)
		INSERT INTO ` + schema.SpaceGroupACLsTable + ` (` + schema.SpaceGroupACLs_USER_EMAIL + `, ` + schema.SpaceGroupACLs_SPACE_GROUP_ID + "," + schema.SpaceGroupACLs_ROLE + `)
		SELECT $2, ` + schema.SpaceGroups_ID + `, $3 FROM new_group
		RETURNING ` + schema.SpaceGroupACLs_SPACE_GROUP_ID + `
	`
	var spaceGroupId int
	if err := s.pool.QueryRow(ctx, query, name, userEmail, schema.SpaceGroupACLs_ROLE_OWNER).Scan(&spaceGroupId); err != nil {
		return 0, err
	}

	return spaceGroupId, nil
}

func (s *pgSpaceGroups) GetSpaceGroup(ctx context.Context, userEmail string, spaceGroupId int, include string) (*models.SpaceGroup, error) {
	keys, fields, err := models.ParseSpaceGroupColumns(include, schema.FullSpaceGroupACLs_ROLE)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpaceGroupsTable + `
		JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.JoinSpaceGroupACLs_SpaceGroups + `
			AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
		WHERE ` + schema.FullSpaceGroups_ID + ` = $1
		ORDER BY ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3
		LIMIT 1`

	sg := &models.SpaceGroup{}
	err = s.pool.QueryRow(ctx, query, spaceGroupId, userEmail, models.PublicAccessEmail).Scan(sg.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return sg, nil
}

func (s *pgSpaceGroups) EditSpaceGroupName(ctx context.Context, userEmail string, spaceGroupId int, name string) (*models.SpaceGroup, error) {
	query := `
		WITH updated AS (
			UPDATE ` + schema.SpaceGroupsTable + `
			SET ` + schema.SpaceGroups_NAME + ` = $1
			WHERE ` + schema.FullSpaceGroups_ID + ` = $2 AND EXISTS (
				SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
				WHERE ` + schema.JoinSpaceGroupACLs_SpaceGroups + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3 AND ` + schema.FilterSpaceGroupACLs_IsEditor + `
			)
			RETURNING ` + schema.SpaceGroups_ID + `, ` + schema.SpaceGroups_NAME + `
		)
		SELECT updated.` + schema.SpaceGroups_ID + `, updated.` + schema.SpaceGroups_NAME + `, ` + schema.FullSpaceGroupACLs_ROLE + `
		FROM updated
		JOIN ` + schema.SpaceGroupACLsTable + ` ON ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = updated.` + schema.SpaceGroups_ID + `
			AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3
	`
	sg := &models.SpaceGroup{}
	err := s.pool.QueryRow(ctx, query, name, spaceGroupId, userEmail).Scan(&sg.Id, &sg.Name, &sg.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return sg, nil
}

func (s *pgSpaceGroups) DelSpaceGroup(ctx context.Context, userEmail string, spaceGroupId int) error {
	query := `
		DELETE FROM ` + schema.SpaceGroupsTable + `
		WHERE ` + schema.FullSpaceGroups_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupACLsTable + `
			WHERE ` + schema.JoinSpaceGroupACLs_SpaceGroups + ` AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 AND ` + schema.FilterSpaceGroupACLs_IsOwner + `
		)
	`
	t, err := s.pool.Exec(ctx, query, spaceGroupId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
