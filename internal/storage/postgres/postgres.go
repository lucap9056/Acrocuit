package postgres

import (
	"acrocuit/internal/storage/repository"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w", err)
	}

	return pool, nil
}

func NewSpaceGroups(pool *pgxpool.Pool) repository.SpaceGroups {
	return &pgSpaceGroups{pool: pool}
}

func NewSpaces(pool *pgxpool.Pool) repository.Spaces {
	return &pgSpaces{pool: pool}
}

func NewBreakerGroups(pool *pgxpool.Pool) repository.BreakerGroups {
	return &pgBreakerGroups{pool: pool}
}

func NewBreakers(pool *pgxpool.Pool) repository.Breakers {
	return &pgBreakers{pool: pool}
}

func NewDevices(pool *pgxpool.Pool) repository.Devices {
	return &pgDevices{pool: pool}
}
