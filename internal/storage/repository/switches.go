package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Switches interface {
	GetSwitches(ctx context.Context, userEmail string, switchPanelId string, include string) ([]models.Switch, error)
	GetSwitch(ctx context.Context, userEmail string, switchId string, include string) (*models.Switch, error)
	AddSwitch(ctx context.Context, userEmail string, switchPanelId string, name string, attributes models.Attributes) (*models.Switch, error)
	SetSwitch(ctx context.Context, userEmail string, switchId string, name *string, attributes models.Attributes) (*models.Switch, error)
	MoveSwitch(ctx context.Context, userEmail string, switchId string, displayOrder int) ([]models.Switch, error)
	DelSwitch(ctx context.Context, userEmail string, switchId string) error
	GetSwitchUpstreams(ctx context.Context, userEmail string, switchId string, include string) ([]models.LinkedSwitch, error)
	AddSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId string, attributes models.Attributes) (*models.SwitchLink, error)
	SetSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId string, attributes models.Attributes) (*models.SwitchLink, error)
	DelSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId string) error
	GetSwitchDownstream(ctx context.Context, userEmail string, switchId string) (*models.PowerChain, error)
}
