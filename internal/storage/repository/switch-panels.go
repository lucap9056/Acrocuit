package repository

import (
	"acrocuit/internal/models"
	"context"
)

type SwitchPanels interface {
	GetSwitchPanels(ctx context.Context, userEmail string, spaceId models.SpaceId, include string) ([]models.SwitchPanel, error)
	GetSwitchPanel(ctx context.Context, userEmail string, switchPanelId models.SwitchPanelId, include string) (*models.SwitchPanel, error)
	AddSwitchPanel(ctx context.Context, userEmail string, spaceId models.SpaceId, name string, position *models.Position, attributes models.Attributes) (*models.SwitchPanel, error)
	SetSwitchPanel(ctx context.Context, userEmail string, switchPanelId models.SwitchPanelId, name *string, attributes models.Attributes) (*models.SwitchPanel, error)
	SetSwitchPanelPosition(ctx context.Context, userEmail string, switchPanelId models.SwitchPanelId, position models.Position) (*models.SwitchPanel, error)
	DelSwitchPanel(ctx context.Context, userEmail string, switchPanelId models.SwitchPanelId) error
}
