package storage

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Spaces interface {
	GetSpaces(ctx context.Context, userId, spaceGroupId int, include string) ([]models.Space, error)
	GetSpace(ctx context.Context, userId, spaceId int, include string) (*models.Space, error)
	AddSpace(ctx context.Context, userId, spaceGroupId int, name string) (*models.Space, error)
	SetSpace(ctx context.Context, userId, spaceId int, name *string, displayOrder *int) (*models.Space, error)
	DelSpace(ctx context.Context, userId, spaceId int) error
}

type pgSpaces struct {
	pool *pgxpool.Pool
}

func (s *pgSpaces) GetSpaces(ctx context.Context, userId, spaceGroupId int, include string) ([]models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_SPACE_GROUP_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1
		ORDER BY ` + schema.FullSpaces_DISPLAY_ORDER

	rows, err := s.pool.Query(ctx, query, userId, spaceGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spaces := make([]models.Space, 0)
	for rows.Next() {
		var space models.Space
		if err := rows.Scan(space.Values(fields)...); err != nil {
			return nil, err
		}
		space.Build()
		spaces = append(spaces, space)
	}

	return spaces, rows.Err()
}

func (s *pgSpaces) GetSpace(ctx context.Context, userId, spaceId int, include string) (*models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1`

	var space models.Space
	err = s.pool.QueryRow(ctx, query, userId, spaceId).Scan(space.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	space.Build()

	return &space, nil
}

func (s *pgSpaces) AddSpace(ctx context.Context, userId, spaceGroupId int, name string) (*models.Space, error) {
	query := `
		WITH permitted AS (
			SELECT 1
			FROM ` + schema.SpaceGroupsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SpaceGroups + `
			WHERE ` + schema.FullSpaceGroups_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
			FOR UPDATE OF ` + schema.SpaceGroupsTable + `
		),
		next_order AS (
			SELECT COALESCE(MAX(` + schema.Spaces_DISPLAY_ORDER + `), 0) + 1 AS ` + schema.Spaces_DISPLAY_ORDER + `
			FROM ` + schema.SpacesTable + `
			WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $1
		),
		inserted AS (
			INSERT INTO ` + schema.SpacesTable + ` (` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `)
			SELECT $3, $1, next_order.` + schema.Spaces_DISPLAY_ORDER + `
			FROM permitted, next_order
			RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `
		)
		SELECT ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + ` FROM inserted
	`

	space := &models.Space{Name: name, SpaceGroupId: spaceGroupId}

	err := s.pool.QueryRow(ctx, query, spaceGroupId, userId, name).
		Scan(&space.Id, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *pgSpaces) SetSpace(ctx context.Context, userId, spaceId int, name *string, displayOrder *int) (*models.Space, error) {
	query := `SELECT ` + schema.FnSetSpace_OUTPUTS + ` FROM ` + schema.FnSetSpace + `($1, $2, $3, $4)`

	space := &models.Space{}
	err := s.pool.QueryRow(ctx, query, userId, spaceId, name, displayOrder).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *pgSpaces) DelSpace(ctx context.Context, userId, spaceId int) error {
	query := `
		DELETE FROM ` + schema.SpacesTable + `
		WHERE ` + schema.FullSpaces_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
			WHERE ` + schema.JoinSpaceGroupOwners_Spaces + ` AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
		)
	`
	t, err := s.pool.Exec(ctx, query, spaceId, userId)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
