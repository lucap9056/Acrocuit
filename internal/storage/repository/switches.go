package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Switches interface {
	GetSwitches(ctx context.Context, userEmail string, switchPanelId int, include string) ([]models.Switch, error)
	GetSwitch(ctx context.Context, userEmail string, switchId int, include string) (*models.Switch, error)
	AddSwitch(ctx context.Context, userEmail string, switchPanelId int, name string, attributes models.Attributes) (*models.Switch, error)
	SetSwitch(ctx context.Context, userEmail string, switchId int, name *string, attributes models.Attributes) (*models.Switch, error)
	MoveSwitch(ctx context.Context, userEmail string, switchId int, displayOrder int) ([]models.Switch, error)
	DelSwitch(ctx context.Context, userEmail string, switchId int) error
	GetSwitchUpstreams(ctx context.Context, userEmail string, switchId int, include string) ([]models.LinkedSwitch, error)
	AddSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId int, attributes models.Attributes) (*models.SwitchLink, error)
	SetSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId int, attributes models.Attributes) (*models.SwitchLink, error)
	DelSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId int) error
	GetSwitchDownstream(ctx context.Context, userEmail string, switchId int) (*models.PowerChain, error)
}
