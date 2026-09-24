package main

import (
	"acrocuit/internal/logs"
	"acrocuit/internal/setup"
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

	dsn := cfg.Database.DSN
	if dsn == "" {
		logs.Out.Fatal("DATABASE_DSN is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logs.Out.Fatal("connect to database failed", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logs.Out.Fatal("ping database failed", zap.Error(err))
	}

	var schemaSQL, functionsSQL strings.Builder
	if err := schema.Render(&schemaSQL); err != nil {
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
