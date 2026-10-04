package postgres

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgSpaces struct {
	pool *pgxpool.Pool
}

func (s *pgSpaces) GetSpaces(ctx context.Context, userEmail string, spaceGroupId int, include string) ([]models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_SPACE_GROUP_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1
		ORDER BY ` + schema.FullSpaces_DISPLAY_ORDER

	rows, err := s.pool.Query(ctx, query, userEmail, spaceGroupId)
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

func (s *pgSpaces) GetSpace(ctx context.Context, userEmail string, spaceId int, include string) (*models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $1`

	var space models.Space
	err = s.pool.QueryRow(ctx, query, userEmail, spaceId).Scan(space.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	space.Build()

	return &space, nil
}

func (s *pgSpaces) AddSpace(ctx context.Context, userEmail string, spaceGroupId int, name string) (*models.Space, error) {
	query := `
		WITH permitted AS (
			SELECT 1
			FROM ` + schema.SpaceGroupsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_SpaceGroups + `
			WHERE ` + schema.FullSpaceGroups_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
			FOR UPDATE OF ` + schema.SpaceGroupsTable + `
		),
		next_order AS (
			SELECT COALESCE(MAX(` + schema.Spaces_DISPLAY_ORDER + `), 0) + 1 AS ` + schema.Spaces_DISPLAY_ORDER + `
			FROM ` + schema.SpacesTable + `
			WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $1
		),
		inserted AS (
			INSERT INTO ` + schema.SpacesTable + ` (` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `)
			SELECT $3, $1, next_order.` + schema.Spaces_DISPLAY_ORDER + `, $4
			FROM permitted, next_order
			RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `
		)
		SELECT ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER + ` FROM inserted
	`

	space := &models.Space{Name: name, SpaceGroupId: spaceGroupId, BackgroundImageUpdatedAt: time.Now().UnixMilli()}

	err := s.pool.QueryRow(ctx, query, spaceGroupId, userEmail, name, space.BackgroundImageUpdatedAt).
		Scan(&space.Id, &space.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *pgSpaces) SetSpace(ctx context.Context, userEmail string, spaceId int, name *string, displayOrder *int) (*models.Space, error) {
	query := `SELECT ` + schema.FnSetSpace_OUTPUTS + ` FROM ` + schema.FnSetSpace + `($1, $2, $3, $4)`

	space := &models.Space{}
	err := s.pool.QueryRow(ctx, query, userEmail, spaceId, name, displayOrder).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *pgSpaces) DelSpace(ctx context.Context, userEmail string, spaceId int) error {
	query := `
		DELETE FROM ` + schema.SpacesTable + `
		WHERE ` + schema.FullSpaces_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
			WHERE ` + schema.JoinSpaceGroupOwners_Spaces + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	t, err := s.pool.Exec(ctx, query, spaceId, userEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSpaces) GetSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId int) (*models.BackgroundImage, error) {
	query := `
		SELECT ` + schema.FullSpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.FullSpaceBackgroundImages_DATA + `, ` + schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT + `
		FROM ` + schema.SpaceBackgroundImagesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSpaceBackgroundImages_Spaces + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
	`

	image := &models.BackgroundImage{}
	err := s.pool.QueryRow(ctx, query, spaceId, userEmail).Scan(&image.ContentType, &image.Data, &image.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *pgSpaces) SetSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId int, contentType string, data []byte) (*models.Space, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	permitted := `
		SELECT 1
		FROM ` + schema.SpacesTable + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		FOR UPDATE OF ` + schema.SpacesTable
	var found int
	err = tx.QueryRow(ctx, permitted, spaceId, userEmail).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	upsert := `
		INSERT INTO ` + schema.SpaceBackgroundImagesTable + ` (` + schema.SpaceBackgroundImages_SPACE_ID + `, ` + schema.SpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.SpaceBackgroundImages_DATA + `)
		VALUES ($1, $2, $3)
		ON CONFLICT (` + schema.SpaceBackgroundImages_SPACE_ID + `) DO UPDATE
		SET ` + schema.SpaceBackgroundImages_CONTENT_TYPE + ` = EXCLUDED.` + schema.SpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.SpaceBackgroundImages_DATA + ` = EXCLUDED.` + schema.SpaceBackgroundImages_DATA
	if _, err := tx.Exec(ctx, upsert, spaceId, contentType, data); err != nil {
		return nil, err
	}

	space, err := touchBackgroundImage(ctx, tx, spaceId)
	if err != nil {
		return nil, err
	}

	return space, tx.Commit(ctx)
}

func (s *pgSpaces) DelSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId int) (*models.Space, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		DELETE FROM ` + schema.SpaceBackgroundImagesTable + `
		WHERE ` + schema.FullSpaceBackgroundImages_SPACE_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.SpacesTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_Spaces + `
			WHERE ` + schema.JoinSpaceBackgroundImages_Spaces + ` AND ` + schema.FullSpaceGroupOwners_USER_EMAIL + ` = $2
		)
	`
	t, err := tx.Exec(ctx, query, spaceId, userEmail)
	if err != nil {
		return nil, err
	}
	if t.RowsAffected() == 0 {
		return nil, repository.ErrNotFound
	}

	space, err := touchBackgroundImage(ctx, tx, spaceId)
	if err != nil {
		return nil, err
	}

	return space, tx.Commit(ctx)
}

func touchBackgroundImage(ctx context.Context, tx pgx.Tx, spaceId int) (*models.Space, error) {
	query := `
		UPDATE ` + schema.SpacesTable + `
		SET ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + ` = $1
		WHERE ` + schema.Spaces_ID + ` = $2
		RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT

	space := &models.Space{}
	err := tx.QueryRow(ctx, query, time.Now().UnixMilli(), spaceId).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if err != nil {
		return nil, err
	}
	space.Build()

	return space, nil
}
