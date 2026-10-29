//go:build js

package sqlite

import (
	"acrocuit/internal/storage/sqlite/jsdriver"
	"context"
	"database/sql"
	"errors"
	"syscall/js"
)

func Connect(ctx context.Context, sqlite3, handle js.Value) (*sql.DB, error) {
	db := sql.OpenDB(jsdriver.NewConnector(sqlite3, handle))
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func isUniqueViolation(err error) bool {
	var sqliteErr *jsdriver.Error
	return errors.As(err, &sqliteErr) && (sqliteErr.Code == jsdriver.ConstraintUnique || sqliteErr.Code == jsdriver.ConstraintPrimaryKey)
}
