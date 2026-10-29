package sqlite

import (
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const connectionParams = "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_txlock=immediate"

func Connect(ctx context.Context, path string) (*sql.DB, error) {
	filePath, _, hasParams := strings.Cut(path, "?")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	separator := "?"
	if hasParams {
		separator = "&"
	}

	db, err := sql.Open("sqlite", "file:"+path+separator+connectionParams)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("%w", err)
	}

	return db, nil
}

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
