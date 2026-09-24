package storage

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Breakers interface {
	GetBreakers(ctx context.Context, userId, breakerGroupId int, include string) ([]models.Breaker, error)
	GetBreaker(ctx context.Context, userId, breakerId int, include string) (*models.Breaker, error)
	AddBreaker(ctx context.Context, userId, breakerGroupId int, name string, upstreamBreakerId *int) (*models.Breaker, error)
	SetBreaker(ctx context.Context, userId, breakerId int, name *string, displayOrder *int) (*models.Breaker, error)
	SetBreakerUpstream(ctx context.Context, userId, breakerId int, upstreamBreakerId *int) (*models.Breaker, error)
	DelBreaker(ctx context.Context, userId, breakerId int) error
	GetBreakerDownstream(ctx context.Context, userId, breakerId int) ([]models.Breaker, []models.Device, error)
}

type pgBreakers struct {
	pool *pgxpool.Pool
}

func (b *pgBreakers) GetBreakers(ctx context.Context, userId, breakerGroupId int, include string) ([]models.Breaker, error) {
	keys, fields, err := models.ParseBreakerColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.BreakersTable + `
		JOIN ` + schema.BreakerGroupsTable + ` ON ` + schema.JoinBreakers_BreakerGroups + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_BreakerGroups + `
		WHERE ` + schema.FullBreakers_BREAKER_GROUP_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1
		ORDER BY ` + schema.FullBreakers_DISPLAY_ORDER

	rows, err := b.pool.Query(ctx, query, userId, breakerGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakers := make([]models.Breaker, 0)
	for rows.Next() {
		var breaker models.Breaker
		if err := rows.Scan(breaker.Values(fields)...); err != nil {
			return nil, err
		}
		breaker.Build()
		breakers = append(breakers, breaker)
	}

	return breakers, rows.Err()
}

func (b *pgBreakers) GetBreaker(ctx context.Context, userId, breakerId int, include string) (*models.Breaker, error) {
	keys, fields, err := models.ParseBreakerColumns(include)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + keys + `
		FROM ` + schema.BreakersTable + `
		JOIN ` + schema.BreakerGroupsTable + ` ON ` + schema.JoinBreakers_BreakerGroups + `
		JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_BreakerGroups + `
		WHERE ` + schema.FullBreakers_ID + ` = $2 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $1`

	var breaker models.Breaker
	err = b.pool.QueryRow(ctx, query, userId, breakerId).Scan(breaker.Values(fields)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	breaker.Build()

	return &breaker, nil
}

func (b *pgBreakers) AddBreaker(ctx context.Context, userId, breakerGroupId int, name string, upstreamBreakerId *int) (*models.Breaker, error) {
	query := `
		WITH permitted AS (
			SELECT ` + schema.FullBreakerGroups_SPACE_GROUP_ID + `
			FROM ` + schema.BreakerGroupsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_BreakerGroups + `
			WHERE ` + schema.FullBreakerGroups_ID + ` = $1 AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
			FOR UPDATE OF ` + schema.BreakerGroupsTable + `
		),
		upstream_ok AS (
			SELECT 1 FROM permitted
			WHERE $4::integer IS NULL OR EXISTS (
				SELECT 1 FROM ` + schema.BreakersTable + ` WHERE ` + schema.Breakers_ID + ` = $4 AND ` + schema.Breakers_SPACE_GROUP_ID + ` = (SELECT ` + schema.BreakerGroups_SPACE_GROUP_ID + ` FROM permitted)
			)
		),
		next_order AS (
			SELECT COALESCE(MAX(` + schema.Breakers_DISPLAY_ORDER + `), 0) + 1 AS ` + schema.Breakers_DISPLAY_ORDER + `
			FROM ` + schema.BreakersTable + `
			WHERE ` + schema.Breakers_BREAKER_GROUP_ID + ` = $1
		),
		inserted AS (
			INSERT INTO ` + schema.BreakersTable + ` (` + schema.Breakers_NAME + `, ` + schema.Breakers_BREAKER_GROUP_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + `, ` + schema.Breakers_UPSTREAM_BREAKER_ID + `)
			SELECT $3, $1, permitted.` + schema.BreakerGroups_SPACE_GROUP_ID + `, next_order.` + schema.Breakers_DISPLAY_ORDER + `, $4
			FROM permitted, next_order
			WHERE EXISTS (SELECT 1 FROM upstream_ok)
			RETURNING ` + schema.Breakers_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + `
		)
		SELECT ` + schema.Breakers_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + ` FROM inserted
	`

	breaker := &models.Breaker{Name: name, BreakerGroupId: breakerGroupId}

	err := b.pool.QueryRow(ctx, query, breakerGroupId, userId, name, upstreamBreakerId).
		Scan(&breaker.Id, &breaker.SpaceGroupId, &breaker.DisplayOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	breaker.UpstreamBreakerId = upstreamBreakerId
	breaker.Build()

	return breaker, nil
}

func (b *pgBreakers) SetBreaker(ctx context.Context, userId, breakerId int, name *string, displayOrder *int) (*models.Breaker, error) {
	query := `SELECT ` + schema.FnSetBreaker_OUTPUTS + ` FROM ` + schema.FnSetBreaker + `($1, $2, $3, $4)`

	breaker := &models.Breaker{}
	err := b.pool.QueryRow(ctx, query, userId, breakerId, name, displayOrder).
		Scan(&breaker.Id, &breaker.Name, &breaker.BreakerGroupId, &breaker.SpaceGroupId, &breaker.DisplayOrder, &breaker.UpstreamBreakerId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	breaker.Build()

	return breaker, nil
}

func (b *pgBreakers) SetBreakerUpstream(ctx context.Context, userId, breakerId int, upstreamBreakerId *int) (*models.Breaker, error) {
	query := `SELECT ` + schema.FnSetBreakerUpstream_OUTPUTS + ` FROM ` + schema.FnSetBreakerUpstream + `($1, $2, $3)`

	breaker := &models.Breaker{}
	err := b.pool.QueryRow(ctx, query, userId, breakerId, upstreamBreakerId).
		Scan(&breaker.Id, &breaker.Name, &breaker.BreakerGroupId, &breaker.SpaceGroupId, &breaker.DisplayOrder, &breaker.UpstreamBreakerId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case schema.FnSetBreakerUpstream_ERR_SELF_REFERENCE:
			return nil, ErrSelfReference
		case schema.FnSetBreakerUpstream_ERR_CYCLE_DETECTED:
			return nil, ErrCycleDetected
		}
	}
	if err != nil {
		return nil, err
	}

	breaker.Build()

	return breaker, nil
}

func (b *pgBreakers) DelBreaker(ctx context.Context, userId, breakerId int) error {
	query := `
		DELETE FROM ` + schema.BreakersTable + `
		WHERE ` + schema.FullBreakers_ID + ` = $1 AND EXISTS (
			SELECT 1 FROM ` + schema.BreakerGroupsTable + `
			JOIN ` + schema.SpaceGroupOwnersTable + ` ON ` + schema.JoinSpaceGroupOwners_BreakerGroups + `
			WHERE ` + schema.JoinBreakers_BreakerGroups + ` AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
		)
	`
	t, err := b.pool.Exec(ctx, query, breakerId, userId)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (b *pgBreakers) GetBreakerDownstream(ctx context.Context, userId, breakerId int) ([]models.Breaker, []models.Device, error) {
	_, fields, err := models.ParseBreakerColumns("")
	if err != nil {
		return nil, nil, err
	}

	query := `
		WITH RECURSIVE breaker_chain AS (
			SELECT ` + schema.FullBreakers_ID + `, ` + schema.FullBreakers_NAME + `, ` + schema.FullBreakers_BREAKER_GROUP_ID + `, ` + schema.FullBreakers_SPACE_GROUP_ID + `, ` + schema.FullBreakers_DISPLAY_ORDER + `, ` + schema.FullBreakers_UPSTREAM_BREAKER_ID + `, 0 AS depth
			FROM ` + schema.BreakersTable + `
			WHERE ` + schema.FullBreakers_ID + ` = $1 AND EXISTS (
				SELECT 1 FROM ` + schema.SpaceGroupOwnersTable + `
				WHERE ` + schema.JoinSpaceGroupOwners_Breakers + ` AND ` + schema.FullSpaceGroupOwners_USER_ID + ` = $2
			)

			UNION ALL

			SELECT ` + schema.FullBreakers_ID + `, ` + schema.FullBreakers_NAME + `, ` + schema.FullBreakers_BREAKER_GROUP_ID + `, ` + schema.FullBreakers_SPACE_GROUP_ID + `, ` + schema.FullBreakers_DISPLAY_ORDER + `, ` + schema.FullBreakers_UPSTREAM_BREAKER_ID + `, breaker_chain.depth + 1
			FROM ` + schema.BreakersTable + `
			JOIN breaker_chain ON ` + schema.FullBreakers_UPSTREAM_BREAKER_ID + ` = breaker_chain.` + schema.Breakers_ID + `
		)
		SELECT ` + schema.Breakers_ID + `, ` + schema.Breakers_NAME + `, ` + schema.Breakers_BREAKER_GROUP_ID + `, ` + schema.Breakers_SPACE_GROUP_ID + `, ` + schema.Breakers_DISPLAY_ORDER + `, ` + schema.Breakers_UPSTREAM_BREAKER_ID + `
		FROM breaker_chain
		ORDER BY depth
	`
	rows, err := b.pool.Query(ctx, query, breakerId, userId)
	if err != nil {
		return nil, nil, err
	}

	breakers := make([]models.Breaker, 0)
	ids := make([]int, 0)
	for rows.Next() {
		var breaker models.Breaker
		if err := rows.Scan(breaker.Values(fields)...); err != nil {
			rows.Close()
			return nil, nil, err
		}
		breaker.Build()
		breakers = append(breakers, breaker)
		ids = append(ids, breaker.Id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(breakers) == 0 {
		return nil, nil, ErrNotFound
	}

	deviceQuery := `
		SELECT ` + schema.FullDevices_ID + `, ` + schema.FullDevices_NAME + `, ` + schema.FullDevices_SPACE_ID + `, ` + schema.FullDeviceBreakers_BREAKER_ID + `
		FROM ` + schema.DevicesTable + `
		JOIN ` + schema.DeviceBreakersTable + ` ON ` + schema.JoinDeviceBreakers_Devices + `
		WHERE ` + schema.FullDeviceBreakers_BREAKER_ID + ` = ANY($1)
		ORDER BY ` + schema.FullDeviceBreakers_BREAKER_ID + `, ` + schema.FullDevices_ID + `
	`
	deviceRows, err := b.pool.Query(ctx, deviceQuery, ids)
	if err != nil {
		return nil, nil, err
	}
	defer deviceRows.Close()

	devices := make([]models.Device, 0)
	for deviceRows.Next() {
		var d models.Device
		if err := deviceRows.Scan(&d.Id, &d.Name, &d.SpaceId, &d.BreakerId); err != nil {
			return nil, nil, err
		}
		d.Build()
		devices = append(devices, d)
	}

	return breakers, devices, deviceRows.Err()
}
