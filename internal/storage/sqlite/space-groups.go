package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"database/sql"
)

const sqliteSpaceGroupRole = `'` + schema.SpaceGroupACLs_ROLE_OWNER + `'`

type sqliteSpaceGroups struct {
	db *sql.DB
}

func (s *sqliteSpaceGroups) GetSpaceGroups(ctx context.Context, _, include string) ([]models.SpaceGroup, error) {
	keys, fields, err := models.ParseSpaceGroupColumns(include, sqliteSpaceGroupRole)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SpaceGroupsTable

	rows, err := s.db.QueryContext(ctx, query)
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

func (s *sqliteSpaceGroups) AddSpaceGroup(ctx context.Context, _, name string) (string, error) {
	query := `INSERT INTO ` + schema.SpaceGroupsTable + ` (` + schema.SpaceGroups_NAME + `) VALUES ($1) RETURNING ` + schema.SpaceGroups_ID

	var spaceGroupId string
	if err := s.db.QueryRowContext(ctx, query, name).Scan(&spaceGroupId); err != nil {
		return "", err
	}

	return spaceGroupId, nil
}

func (s *sqliteSpaceGroups) GetSpaceGroup(ctx context.Context, _ string, spaceGroupId string, include string) (*models.SpaceGroup, error) {
	keys, fields, err := models.ParseSpaceGroupColumns(include, sqliteSpaceGroupRole)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SpaceGroupsTable + ` WHERE ` + schema.FullSpaceGroups_ID + ` = $1`

	sg := &models.SpaceGroup{}
	if err := s.db.QueryRowContext(ctx, query, spaceGroupId).Scan(sg.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}

	return sg, nil
}

func (s *sqliteSpaceGroups) EditSpaceGroupName(ctx context.Context, _ string, spaceGroupId string, name string) (*models.SpaceGroup, error) {
	query := `UPDATE ` + schema.SpaceGroupsTable + ` SET ` + schema.SpaceGroups_NAME + ` = $1 WHERE ` + schema.SpaceGroups_ID + ` = $2
		RETURNING ` + schema.SpaceGroups_ID + `, ` + schema.SpaceGroups_NAME + `, ` + sqliteSpaceGroupRole

	sg := &models.SpaceGroup{}
	if err := s.db.QueryRowContext(ctx, query, name, spaceGroupId).Scan(&sg.Id, &sg.Name, &sg.Role); err != nil {
		return nil, notFoundIfNoRows(err)
	}

	return sg, nil
}

func (s *sqliteSpaceGroups) DelSpaceGroup(ctx context.Context, _ string, spaceGroupId string) error {
	query := `DELETE FROM ` + schema.SpaceGroupsTable + ` WHERE ` + schema.SpaceGroups_ID + ` = $1`
	return notFoundIfNoneAffected(s.db.ExecContext(ctx, query, spaceGroupId))
}
