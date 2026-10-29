//go:build !js

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	modernc "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
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

func isUniqueViolation(err error) bool {
	var sqliteErr *modernc.Error
	return errors.As(err, &sqliteErr) && (sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE || sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY)
}
