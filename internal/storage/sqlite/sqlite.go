package sqlite

import (
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
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

type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

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

func NewBreakerGroups(db *sql.DB) repository.BreakerGroups {
	return &sqliteBreakerGroups{db: db}
}

func NewBreakers(db *sql.DB) repository.Breakers {
	return &sqliteBreakers{db: db}
}

func NewDevices(db *sql.DB) repository.Devices {
	return &sqliteDevices{db: db}
}

func withTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func exists(ctx context.Context, q querier, query string, args ...any) (bool, error) {
	var found bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (`+query+`)`, args...).Scan(&found)
	return found, err
}

func requireRow(ctx context.Context, q querier, table, idColumn string, id int) error {
	found, err := exists(ctx, q, `SELECT 1 FROM `+table+` WHERE `+idColumn+` = $1`, id)
	if err != nil {
		return err
	}
	if !found {
		return repository.ErrNotFound
	}
	return nil
}

func notFoundIfNoRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrNotFound
	}
	return err
}

func notFoundIfNoneAffected(result sql.Result, err error) error {
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

func isUniqueViolation(err error) bool {
	var sqliteErr *modernc.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	code := sqliteErr.Code()
	return code == sqlite3.SQLITE_CONSTRAINT_UNIQUE || code == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY
}

type orderedTable struct {
	table, id, group, order string
}

var (
	spacesOrder   = orderedTable{schema.SpacesTable, schema.Spaces_ID, schema.Spaces_SPACE_GROUP_ID, schema.Spaces_DISPLAY_ORDER}
	breakersOrder = orderedTable{schema.BreakersTable, schema.Breakers_ID, schema.Breakers_BREAKER_GROUP_ID, schema.Breakers_DISPLAY_ORDER}
)

func (o orderedTable) move(ctx context.Context, tx *sql.Tx, groupId, id, from, to int) error {
	low, high, shift := from, to, -1
	if to < from {
		low, high, shift = to, from, 1
	}

	var floor int
	err := tx.QueryRowContext(ctx, `SELECT MIN(`+o.order+`) FROM `+o.table+` WHERE `+o.group+` = $1`, groupId).Scan(&floor)
	if err != nil {
		return err
	}

	park := `
		UPDATE ` + o.table + `
		SET ` + o.order + ` = $1 - 1 - ($2 - CASE WHEN ` + o.id + ` = $3 THEN $4 ELSE ` + o.order + ` + $5 END)
		WHERE ` + o.group + ` = $6 AND ` + o.order + ` BETWEEN $7 AND $2
	`
	if _, err := tx.ExecContext(ctx, park, floor, high, id, to, shift, groupId, low); err != nil {
		return err
	}

	restore := `
		UPDATE ` + o.table + `
		SET ` + o.order + ` = $1 - ($2 - 1 - ` + o.order + `)
		WHERE ` + o.group + ` = $3 AND ` + o.order + ` < $2
	`
	_, err = tx.ExecContext(ctx, restore, high, floor, groupId)
	return err
}
