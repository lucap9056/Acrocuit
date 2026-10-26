package postgres

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const selectSwitchChainWithPanel = `
	SELECT ` + models.SWITCH_ALL_COLUMNS + `, ` + models.SWITCH_PANEL_WITH_POSITION_COLUMNS + `
	FROM switch_chain
	JOIN ` + schema.SwitchesTable + ` ON ` + schema.FullSwitches_ID + ` = switch_chain.id
	JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.JoinSwitches_SwitchPanels + `
	LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels + `
	ORDER BY ` + schema.FullSwitches_ID

func switchIdsOf(switches []models.Switch) []int {
	ids := make([]int, len(switches))
	for i, sw := range switches {
		ids[i] = sw.Id
	}
	return ids
}

func scanAll[T any](rows pgx.Rows, err error, values func(*T) []any) ([]T, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]T, 0)
	for rows.Next() {
		var item T
		if err := rows.Scan(values(&item)...); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func querySwitchesWithPanel(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]models.Switch, error) {
	rows, err := pool.Query(ctx, query, args...)
	switches, err := scanAll(rows, err, (*models.Switch).ValuesWithPanel)
	for i := range switches {
		switches[i].Build()
	}
	return switches, err
}

func queryDevicesWithPosition(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]models.Device, error) {
	rows, err := pool.Query(ctx, query, args...)
	devices, err := scanAll(rows, err, (*models.Device).ValuesWithPosition)
	for i := range devices {
		devices[i].Build()
	}
	return devices, err
}

func querySwitchLinks(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]models.SwitchLink, error) {
	rows, err := pool.Query(ctx, query, args...)
	return scanAll(rows, err, (*models.SwitchLink).Values)
}

func queryDeviceLinks(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]models.DeviceLink, error) {
	rows, err := pool.Query(ctx, query, args...)
	return scanAll(rows, err, (*models.DeviceLink).Values)
}
