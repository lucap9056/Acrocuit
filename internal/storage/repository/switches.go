package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Switches interface {
	GetSwitches(ctx context.Context, userEmail string, switchPanelId int, include string) ([]models.Switch, error)
	GetSwitch(ctx context.Context, userEmail string, switchId int, include string) (*models.Switch, error)
	AddSwitch(ctx context.Context, userEmail string, switchPanelId int, name string, upstreamSwitchId *int) (*models.Switch, error)
	SetSwitch(ctx context.Context, userEmail string, switchId int, name *string) (*models.Switch, error)
	MoveSwitch(ctx context.Context, userEmail string, switchId int, displayOrder int) ([]models.Switch, error)
	SetSwitchUpstream(ctx context.Context, userEmail string, switchId int, upstreamSwitchId *int) (*models.Switch, error)
	DelSwitch(ctx context.Context, userEmail string, switchId int) error
	GetSwitchDownstream(ctx context.Context, userEmail string, switchId int) ([]models.Switch, []models.Device, error)
}
