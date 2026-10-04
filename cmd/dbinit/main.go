package main

import (
	"acrocuit/internal/logs"
	"acrocuit/internal/options"
	"acrocuit/internal/setup"
	"acrocuit/internal/storage/sqlite"
	"acrocuit/schema"
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	cfg := setup.Load()

	logs.InitLogger(cfg.Logging)
	defer logs.Sync()

	database := cfg.Options.Database

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch database.Driver {
	case options.DriverPostgres:
		initPostgres(ctx, database.Source)
	case options.DriverSQLite:
		initSQLite(ctx, database.Source)
	default:
		logs.Out.Fatal("unsupported database driver", zap.String("driver", string(database.Driver)))
	}
}

func initSQLite(ctx context.Context, path string) {
	db, err := sqlite.Connect(ctx, path)
	if err != nil {
		logs.Out.Fatal("connect to database failed", zap.Error(err))
	}
	defer db.Close()

	if err := sqlite.ApplySchema(ctx, db); err != nil {
		logs.Out.Fatal("apply schema failed", zap.Error(err))
	}
	logs.Out.Info("schema applied")
}

func initPostgres(ctx context.Context, dsn string) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logs.Out.Fatal("connect to database failed", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logs.Out.Fatal("ping database failed", zap.Error(err))
	}

	var schemaSQL, functionsSQL strings.Builder
	if err := schema.RenderPostgres(&schemaSQL); err != nil {
		logs.Out.Fatal("render schema failed", zap.Error(err))
	}
	if err := schema.RenderFunctions(&functionsSQL); err != nil {
		logs.Out.Fatal("render functions failed", zap.Error(err))
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		logs.Out.Fatal("acquire connection failed", zap.Error(err))
	}
	defer conn.Release()

	pgConn := conn.Conn().PgConn()

	if _, err := pgConn.Exec(ctx, schemaSQL.String()).ReadAll(); err != nil {
		logs.Out.Fatal("apply schema failed", zap.Error(err))
	}
	logs.Out.Info("schema applied")

	if _, err := pgConn.Exec(ctx, functionsSQL.String()).ReadAll(); err != nil {
		logs.Out.Fatal("apply functions failed", zap.Error(err))
	}
	logs.Out.Info("functions applied")
}
