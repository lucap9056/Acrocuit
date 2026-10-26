package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"database/sql"
	"encoding/json"
)

const selectSwitchChainWithPanel = `
	SELECT ` + models.SWITCH_ALL_COLUMNS + `, ` + models.SWITCH_PANEL_WITH_POSITION_COLUMNS + `
	FROM switch_chain
	JOIN ` + schema.SwitchesTable + ` ON ` + schema.FullSwitches_ID + ` = switch_chain.id
	JOIN ` + schema.SwitchPanelsTable + ` ON ` + schema.JoinSwitches_SwitchPanels + `
	LEFT JOIN ` + schema.SwitchPanelPositionsTable + ` ON ` + schema.JoinSwitchPanelPositions_SwitchPanels + `
	ORDER BY ` + schema.FullSwitches_ID

func switchIdsJSON(switches []models.ChainSwitch) (string, error) {
	ids := make([]int, len(switches))
	for i, sw := range switches {
		ids[i] = sw.Id
	}
	raw, err := json.Marshal(ids)
	return string(raw), err
}

func scanAll[T any](rows *sql.Rows, err error, values func(*T) []any) ([]T, error) {
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

func querySwitchesWithPanel(ctx context.Context, q querier, query string, args ...any) ([]models.ChainSwitch, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	switches, err := scanAll(rows, err, (*models.ChainSwitch).ValuesWithPanel)
	for i := range switches {
		switches[i].Build()
	}
	return switches, err
}

func queryDevicesWithPosition(ctx context.Context, q querier, query string, args ...any) ([]models.Device, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	devices, err := scanAll(rows, err, (*models.Device).ValuesWithPosition)
	for i := range devices {
		devices[i].Build()
	}
	return devices, err
}

func querySwitchLinks(ctx context.Context, q querier, query string, args ...any) ([]models.SwitchLink, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	return scanAll(rows, err, (*models.SwitchLink).Values)
}

func queryDeviceLinks(ctx context.Context, q querier, query string, args ...any) ([]models.DeviceLink, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	return scanAll(rows, err, (*models.DeviceLink).Values)
}
