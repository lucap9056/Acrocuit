package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"database/sql"
	"time"
)

type sqliteSpaces struct {
	db *sql.DB
}

func (s *sqliteSpaces) GetSpaces(ctx context.Context, _ string, spaceGroupId int, include string) ([]models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.SpacesTable + `
		WHERE ` + schema.FullSpaces_SPACE_GROUP_ID + ` = $1
		ORDER BY ` + schema.FullSpaces_DISPLAY_ORDER

	return querySpaces(ctx, s.db, fields, query, spaceGroupId)
}

func querySpaces(ctx context.Context, q querier, fields []uint8, query string, args ...any) ([]models.Space, error) {
	rows, err := q.QueryContext(ctx, query, args...)
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

func (s *sqliteSpaces) GetSpace(ctx context.Context, _ string, spaceId int, include string) (*models.Space, error) {
	keys, fields, err := models.ParseSpaceColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.FullSpaces_ID + ` = $1`

	var space models.Space
	if err := s.db.QueryRowContext(ctx, query, spaceId).Scan(space.Values(fields)...); err != nil {
		return nil, notFoundIfNoRows(err)
	}
	space.Build()

	return &space, nil
}

func (s *sqliteSpaces) AddSpace(ctx context.Context, _ string, spaceGroupId int, name string) (*models.Space, error) {
	space := &models.Space{Name: name, SpaceGroupId: spaceGroupId}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SpaceGroupsTable, schema.SpaceGroups_ID, spaceGroupId); err != nil {
			return err
		}

		insert := `
			INSERT INTO ` + schema.SpacesTable + ` (` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `)
			SELECT $1, $2, COALESCE(MAX(` + schema.Spaces_DISPLAY_ORDER + `), 0) + 1, 0
			FROM ` + schema.SpacesTable + `
			WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $2
			RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER
		return tx.QueryRowContext(ctx, insert, name, spaceGroupId).
			Scan(&space.Id, &space.DisplayOrder)
	})
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *sqliteSpaces) MoveSpace(ctx context.Context, _ string, spaceId int, displayOrder int) ([]models.Space, error) {
	var spaces []models.Space

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var spaceGroupId, currentOrder int
		current := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, current, spaceId).Scan(&spaceGroupId, &currentOrder); err != nil {
			return notFoundIfNoRows(err)
		}

		if displayOrder != currentOrder {
			if err := spacesOrder.move(ctx, tx, spaceGroupId, spaceId, currentOrder, displayOrder); err != nil {
				return err
			}
		}

		keys, fields, err := models.ParseSpaceColumns("")
		if err != nil {
			return err
		}
		query := `SELECT ` + keys + `
			FROM ` + schema.SpacesTable + `
			WHERE ` + schema.FullSpaces_SPACE_GROUP_ID + ` = $1
			ORDER BY ` + schema.FullSpaces_DISPLAY_ORDER
		spaces, err = querySpaces(ctx, tx, fields, query, spaceGroupId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (s *sqliteSpaces) SetSpace(ctx context.Context, _ string, spaceId int, name *string) (*models.Space, error) {
	space := &models.Space{}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SpacesTable, schema.Spaces_ID, spaceId); err != nil {
			return err
		}

		if name != nil {
			rename := `UPDATE ` + schema.SpacesTable + ` SET ` + schema.Spaces_NAME + ` = $1 WHERE ` + schema.Spaces_ID + ` = $2`
			if _, err := tx.ExecContext(ctx, rename, *name, spaceId); err != nil {
				return err
			}
		}

		result := `
			SELECT ` + schema.Spaces_ID + `, ` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `
			FROM ` + schema.SpacesTable + `
			WHERE ` + schema.Spaces_ID + ` = $1
		`
		return tx.QueryRowContext(ctx, result, spaceId).
			Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	})
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *sqliteSpaces) DelSpace(ctx context.Context, _ string, spaceId int) error {
	query := `DELETE FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
	return notFoundIfNoneAffected(s.db.ExecContext(ctx, query, spaceId))
}

func (s *sqliteSpaces) GetSpaceBackgroundImage(ctx context.Context, _ string, spaceId int) (*models.BackgroundImage, error) {
	query := `
		SELECT ` + schema.FullSpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.FullSpaceBackgroundImages_DATA + `, ` + schema.FullSpaces_BACKGROUND_IMAGE_UPDATED_AT + `
		FROM ` + schema.SpaceBackgroundImagesTable + `
		JOIN ` + schema.SpacesTable + ` ON ` + schema.JoinSpaceBackgroundImages_Spaces + `
		WHERE ` + schema.FullSpaces_ID + ` = $1
	`

	image := &models.BackgroundImage{}
	if err := s.db.QueryRowContext(ctx, query, spaceId).Scan(&image.ContentType, &image.Data, &image.UpdatedAt); err != nil {
		return nil, notFoundIfNoRows(err)
	}

	return image, nil
}

func (s *sqliteSpaces) SetSpaceBackgroundImage(ctx context.Context, _ string, spaceId int, contentType string, data []byte) (*models.Space, error) {
	var space *models.Space

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SpacesTable, schema.Spaces_ID, spaceId); err != nil {
			return err
		}

		upsert := `
			INSERT INTO ` + schema.SpaceBackgroundImagesTable + ` (` + schema.SpaceBackgroundImages_SPACE_ID + `, ` + schema.SpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.SpaceBackgroundImages_DATA + `)
			VALUES ($1, $2, $3)
			ON CONFLICT (` + schema.SpaceBackgroundImages_SPACE_ID + `) DO UPDATE
			SET ` + schema.SpaceBackgroundImages_CONTENT_TYPE + ` = excluded.` + schema.SpaceBackgroundImages_CONTENT_TYPE + `, ` + schema.SpaceBackgroundImages_DATA + ` = excluded.` + schema.SpaceBackgroundImages_DATA
		if _, err := tx.ExecContext(ctx, upsert, spaceId, contentType, data); err != nil {
			return err
		}

		var err error
		space, err = setBackgroundImageUpdatedAt(ctx, tx, spaceId, time.Now().UnixMilli())
		return err
	})
	if err != nil {
		return nil, err
	}

	return space, nil
}

func (s *sqliteSpaces) DelSpaceBackgroundImage(ctx context.Context, _ string, spaceId int) (*models.Space, error) {
	var space *models.Space

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		query := `DELETE FROM ` + schema.SpaceBackgroundImagesTable + ` WHERE ` + schema.SpaceBackgroundImages_SPACE_ID + ` = $1`
		if err := notFoundIfNoneAffected(tx.ExecContext(ctx, query, spaceId)); err != nil {
			return err
		}

		var err error
		space, err = setBackgroundImageUpdatedAt(ctx, tx, spaceId, 0)
		return err
	})
	if err != nil {
		return nil, err
	}

	return space, nil
}

func setBackgroundImageUpdatedAt(ctx context.Context, tx *sql.Tx, spaceId int, updatedAt int64) (*models.Space, error) {
	query := `
		UPDATE ` + schema.SpacesTable + `
		SET ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + ` = $1
		WHERE ` + schema.Spaces_ID + ` = $2
		RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT

	space := &models.Space{}
	err := tx.QueryRowContext(ctx, query, updatedAt, spaceId).
		Scan(&space.Id, &space.Name, &space.SpaceGroupId, &space.DisplayOrder, &space.BackgroundImageUpdatedAt)
	if err != nil {
		return nil, err
	}
	space.Build()

	return space, nil
}
