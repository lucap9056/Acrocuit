package storage

import (
	"acrocuit/internal/options"
	"acrocuit/internal/storage/postgres"
	"acrocuit/internal/storage/repository"
	"acrocuit/internal/storage/sqlite"
	"context"
	"fmt"
)

type Storage struct {
	close func()

	SpaceGroups   repository.SpaceGroups
	Spaces        repository.Spaces
	BreakerGroups repository.BreakerGroups
	Breakers      repository.Breakers
	Devices       repository.Devices
}

func New(ctx context.Context, opts *options.Options) (*Storage, error) {
	switch opts.Database.Driver {
	case options.DriverPostgres:
		return newPostgres(ctx, opts.Database.Source)
	case options.DriverSQLite:
		return newSQLite(ctx, opts.Database.Source)
	default:
		return nil, fmt.Errorf("storage: unsupported database driver %q", opts.Database.Driver)
	}
}

func newPostgres(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &Storage{
		close:         pool.Close,
		SpaceGroups:   postgres.NewSpaceGroups(pool),
		Spaces:        postgres.NewSpaces(pool),
		BreakerGroups: postgres.NewBreakerGroups(pool),
		Breakers:      postgres.NewBreakers(pool),
		Devices:       postgres.NewDevices(pool),
	}, nil
}

func newSQLite(ctx context.Context, path string) (*Storage, error) {
	db, err := sqlite.Connect(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := sqlite.ApplySchema(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{
		close:         func() { db.Close() },
		SpaceGroups:   sqlite.NewSpaceGroups(db),
		Spaces:        sqlite.NewSpaces(db),
		BreakerGroups: sqlite.NewBreakerGroups(db),
		Breakers:      sqlite.NewBreakers(db),
		Devices:       sqlite.NewDevices(db),
	}, nil
}

func (s *Storage) Close() {
	s.close()
}
