package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"errors"
	"time"
)

type sqliteSpaces struct {
	db *sql.DB
}

func (s *sqliteSpaces) GetSpaces(ctx context.Context, _ string, spaceGroupId models.SpaceGroupId, include string) ([]models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		WHERE ` + schema.FullSpaces_SPACE_GROUP_ID + ` = $1
		ORDER BY ` + schema.FullSpaces_DISPLAY_ORDER

	rows, err := s.db.QueryContext(ctx, query, spaceGroupId)
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

func (s *sqliteSpaces) GetSpace(ctx context.Context, _ string, spaceId models.SpaceId, include string) (*models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.FullSpaces_ID + ` = $1`

	var space models.Space
	err = s.db.QueryRowContext(ctx, query, spaceId).Scan(space.Values(fields)...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	space.Build()

	return &space, nil
}

func (s *sqliteSpaces) AddSpace(ctx context.Context, _ string, spaceGroupId models.SpaceGroupId, name string) (*models.Space, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	spaceGroup := `SELECT 1 FROM ` + schema.SpaceGroupsTable + ` WHERE ` + schema.SpaceGroups_ID + ` = $1`
	var found int
	err = tx.QueryRowContext(ctx, spaceGroup, spaceGroupId).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	insert := `
		INSERT INTO ` + schema.SpacesTable + ` (` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `)
		SELECT $1, $2, COALESCE(MAX(` + schema.Spaces_DISPLAY_ORDER + `), 0) + 1, 0
		FROM ` + schema.SpacesTable + `
		WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $2
		RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER

	space := &models.Space{Name: name, SpaceGroupId: spaceGroupId}
	if err := tx.QueryRowContext(ctx, insert, name, spaceGroupId).Scan(&space.Id, &space.DisplayOrder); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *sqliteSpaces) MoveSpace(ctx context.Context, _ string, spaceId models.SpaceId, displayOrder int) ([]models.Space, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	current := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
	var spaceGroupId models.SpaceGroupId
	var currentOrder int
	err = tx.QueryRowContext(ctx, current, spaceId).Scan(&spaceGroupId, &currentOrder)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if displayOrder != currentOrder {
		low, high, shift := currentOrder, displayOrder, -1
		if displayOrder < currentOrder {
			low, high, shift = displayOrder, currentOrder, 1
		}

		floorQuery := `SELECT MIN(` + schema.Spaces_DISPLAY_ORDER + `) FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $1`
		var floor int
		if err := tx.QueryRowContext(ctx, floorQuery, spaceGroupId).Scan(&floor); err != nil {
			return nil, err
		}

		park := `
			UPDATE ` + schema.SpacesTable + `
			SET ` + schema.Spaces_DISPLAY_ORDER + ` = $1 - 1 - ($2 - CASE WHEN ` + schema.Spaces_ID + ` = $3 THEN $4 ELSE ` + schema.Spaces_DISPLAY_ORDER + ` + $5 END)
			WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $6 AND ` + schema.Spaces_DISPLAY_ORDER + ` BETWEEN $7 AND $2
		`
		if _, err := tx.ExecContext(ctx, park, floor, high, spaceId, displayOrder, shift, spaceGroupId, low); err != nil {
			return nil, err
		}

		restore := `
			UPDATE ` + schema.SpacesTable + `
			SET ` + schema.Spaces_DISPLAY_ORDER + ` = $1 - ($2 - 1 - ` + schema.Spaces_DISPLAY_ORDER + `)
			WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $3 AND ` + schema.Spaces_DISPLAY_ORDER + ` < $2
		`
		if _, err := tx.ExecContext(ctx, restore, high, floor, spaceGroupId); err != nil {
			return nil, err
		}
	}

	keys, fields, err := models.ParseSpaceColumns("")
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		WHERE ` + schema.FullSpaces_SPACE_GROUP_ID + ` = $1
		ORDER BY ` + schema.FullSpaces_DISPLAY_ORDER

	rows, err := tx.QueryContext(ctx, query, spaceGroupId)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return spaces, nil
}

func (s *sqliteSpaces) SetSpace(ctx context.Context, _ string, spaceId models.SpaceId, name *string) (*models.Space, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	target := `SELECT 1 FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
	var found int
	err = tx.QueryRowContext(ctx, target, spaceId).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if name != nil {
		rename := `UPDATE ` + schema.SpacesTable + ` SET ` + schema.Spaces_NAME + ` = $1 WHERE ` + schema.Spaces_ID + ` = $2`
		if _, err := tx.ExecContext(ctx, rename, *name, spaceId); err != nil {
			return nil, err
		}
	}

	result := `
		SELECT ` + schema.Spaces_ID + `, ` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `
		FROM ` + schema.SpacesTable + `
		WHERE ` + schema.Spaces_ID + ` = $1
	`
	space := &models.Space{}
	err = tx.QueryRowContext(ctx, result, spaceId).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *sqliteSpaces) DelSpace(ctx context.Context, _ string, spaceId models.SpaceId) error {
	query := `DELETE FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
	result, err := s.db.ExecContext(ctx, query, spaceId)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *sqliteSpaces) GetSpaceBackgroundImage(ctx context.Context, _ string, spaceId models.SpaceId) (*models.BackgroundImage, error) {
	query := `
		SELECT ` + schema.FullSpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.FullSpaceBackgroundImages_DATA + `, ` + schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT + `
		FROM ` + schema.SpaceBackgroundImagesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSpaceBackgroundImages_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $1
	`

	image := &models.BackgroundImage{}
	err := s.db.QueryRowContext(ctx, query, spaceId).Scan(&image.ContentType, &image.Data, &image.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *sqliteSpaces) SetSpaceBackgroundImage(ctx context.Context, _ string, spaceId models.SpaceId, contentType string, data []byte) (*models.Space, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	target := `SELECT 1 FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
	var found int
	err = tx.QueryRowContext(ctx, target, spaceId).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	upsert := `
		INSERT INTO ` + schema.SpaceBackgroundImagesTable + ` (` + schema.SpaceBackgroundImages_SPACE_ID + `, ` + schema.SpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.SpaceBackgroundImages_DATA + `)
		VALUES ($1, $2, $3)
		ON CONFLICT (` + schema.SpaceBackgroundImages_SPACE_ID + `) DO UPDATE
		SET ` + schema.SpaceBackgroundImages_CONTENT_TYPE + ` = excluded.` + schema.SpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.SpaceBackgroundImages_DATA + ` = excluded.` + schema.SpaceBackgroundImages_DATA
	if _, err := tx.ExecContext(ctx, upsert, spaceId, contentType, data); err != nil {
		return nil, err
	}

	touch := `
		UPDATE ` + schema.SpacesTable + `
		SET ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + ` = $1
		WHERE ` + schema.Spaces_ID + ` = $2
		RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT
	space := &models.Space{}
	err = tx.QueryRowContext(ctx, touch, time.Now().UnixMilli(), spaceId).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *sqliteSpaces) DelSpaceBackgroundImage(ctx context.Context, _ string, spaceId models.SpaceId) (*models.Space, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `DELETE FROM ` + schema.SpaceBackgroundImagesTable + ` WHERE ` + schema.SpaceBackgroundImages_SPACE_ID + ` = $1`
	result, err := tx.ExecContext(ctx, query, spaceId)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, repository.ErrNotFound
	}

	touch := `
		UPDATE ` + schema.SpacesTable + `
		SET ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + ` = 0
		WHERE ` + schema.Spaces_ID + ` = $1
		RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT
	space := &models.Space{}
	err = tx.QueryRowContext(ctx, touch, spaceId).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}
