package sqlite

import (
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"strings"
)

func ApplySchema(ctx context.Context, db *sql.DB) error {
	var schemaSQL strings.Builder
	if err := schema.RenderSQLite(&schemaSQL); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, schemaSQL.String())
	return err
}

func NewSpaceGroups(db *sql.DB) repository.SpaceGroups {
	return &sqliteSpaceGroups{db: db}
}

func NewSpaces(db *sql.DB) repository.Spaces {
	return &sqliteSpaces{db: db}
}

func NewSwitchPanels(db *sql.DB) repository.SwitchPanels {
	return &sqliteSwitchPanels{db: db}
}

func NewSwitches(db *sql.DB) repository.Switches {
	return &sqliteSwitches{db: db}
}

func NewDevices(db *sql.DB) repository.Devices {
	return &sqliteDevices{db: db}
}
