package main

import (
	"acrocuit/internal/setup"
	"acrocuit/schema"
	"context"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := setup.Load()

	dsn := cfg.Database.DSN
	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	var schemaSQL, functionsSQL strings.Builder
	if err := schema.Render(&schemaSQL); err != nil {
		log.Fatalf("render schema: %v", err)
	}
	if err := schema.RenderFunctions(&functionsSQL); err != nil {
		log.Fatalf("render functions: %v", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire connection: %v", err)
	}
	defer conn.Release()

	pgConn := conn.Conn().PgConn()

	if _, err := pgConn.Exec(ctx, schemaSQL.String()).ReadAll(); err != nil {
		log.Fatalf("apply schema: %v", err)
	}
	log.Println("schema applied")

	if _, err := pgConn.Exec(ctx, functionsSQL.String()).ReadAll(); err != nil {
		log.Fatalf("apply functions: %v", err)
	}
	log.Println("functions applied")
}
