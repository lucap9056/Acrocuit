package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool

	SpaceGroups   SpaceGroups
	Spaces        Spaces
	BreakerGroups BreakerGroups
	Breakers      Breakers
	Devices       Devices
	Users         Users
}

func New(ctx context.Context, dsn string) (*Storage, error) {

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &Storage{
		pool:          pool,
		SpaceGroups:   &pgSpaceGroups{pool: pool},
		Spaces:        &pgSpaces{pool: pool},
		BreakerGroups: &pgBreakerGroups{pool: pool},
		Breakers:      &pgBreakers{pool: pool},
		Devices:       &pgDevices{pool: pool},
		Users:         &pgUsers{pool: pool},
	}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
