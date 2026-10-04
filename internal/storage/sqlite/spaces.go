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
	space := &models.Space{Name: name, SpaceGroupId: spaceGroupId, BackgroundImageUpdatedAt: time.Now().UnixMilli()}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := requireRow(ctx, tx, schema.SpaceGroupsTable, schema.SpaceGroups_ID, spaceGroupId); err != nil {
			return err
		}

		insert := `
			INSERT INTO ` + schema.SpacesTable + ` (` + schema.Spaces_NAME + `, ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + `, ` + schema.Spaces_BACKGROUND_IMAGE_UPDATED_AT + `)
			SELECT $1, $2, COALESCE(MAX(` + schema.Spaces_DISPLAY_ORDER + `), 0) + 1, $3
			FROM ` + schema.SpacesTable + `
			WHERE ` + schema.Spaces_SPACE_GROUP_ID + ` = $2
			RETURNING ` + schema.Spaces_ID + `, ` + schema.Spaces_DISPLAY_ORDER
		return tx.QueryRowContext(ctx, insert, name, spaceGroupId, space.BackgroundImageUpdatedAt).
			Scan(&space.Id, &space.DisplayOrder)
	})
	if err != nil {
		return nil, err
	}

	space.Build()

	return space, nil
}

func (s *sqliteSpaces) SetSpace(ctx context.Context, _ string, spaceId int, name *string, displayOrder *int) (*models.Space, error) {
	space := &models.Space{}

	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var spaceGroupId, currentOrder int
		current := `SELECT ` + schema.Spaces_SPACE_GROUP_ID + `, ` + schema.Spaces_DISPLAY_ORDER + ` FROM ` + schema.SpacesTable + ` WHERE ` + schema.Spaces_ID + ` = $1`
		if err := tx.QueryRowContext(ctx, current, spaceId).Scan(&spaceGroupId, &currentOrder); err != nil {
			return notFoundIfNoRows(err)
		}

		if displayOrder != nil && *displayOrder != currentOrder {
			if err := spacesOrder.move(ctx, tx, spaceGroupId, spaceId, currentOrder, *displayOrder); err != nil {
				return err
			}
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
