package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Switches interface {
	GetSwitches(ctx context.Context, userEmail string, switchPanelId models.SwitchPanelId, include string) ([]models.Switch, error)
	GetSwitch(ctx context.Context, userEmail string, switchId models.SwitchId, include string) (*models.Switch, error)
	AddSwitch(ctx context.Context, userEmail string, switchPanelId models.SwitchPanelId, name string, attributes models.Attributes) (*models.Switch, error)
	SetSwitch(ctx context.Context, userEmail string, switchId models.SwitchId, name *string, attributes models.Attributes) (*models.Switch, error)
	MoveSwitch(ctx context.Context, userEmail string, switchId models.SwitchId, displayOrder int) ([]models.Switch, error)
	DelSwitch(ctx context.Context, userEmail string, switchId models.SwitchId) error
	GetSwitchUpstreams(ctx context.Context, userEmail string, switchId models.SwitchId, include string) ([]models.LinkedSwitch, error)
	AddSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId models.SwitchId, attributes models.Attributes) (*models.SwitchLink, error)
	SetSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId models.SwitchId, attributes models.Attributes) (*models.SwitchLink, error)
	DelSwitchUpstream(ctx context.Context, userEmail string, switchId, upstreamSwitchId models.SwitchId) error
	GetSwitchDownstream(ctx context.Context, userEmail string, switchId models.SwitchId) (*models.PowerChainGraph, error)
}
