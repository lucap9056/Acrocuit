package repository

import (
	"acrocuit/internal/models"
	"context"
)

type SwitchPanels interface {
	GetSwitchPanels(ctx context.Context, userEmail string, spaceId int, include string) ([]models.SwitchPanel, error)
	GetSwitchPanel(ctx context.Context, userEmail string, switchPanelId int, include string) (*models.SwitchPanel, error)
	AddSwitchPanel(ctx context.Context, userEmail string, spaceId int, name string, position *models.Position) (*models.SwitchPanel, error)
	SetSwitchPanel(ctx context.Context, userEmail string, switchPanelId int, name string) (*models.SwitchPanel, error)
	SetSwitchPanelPosition(ctx context.Context, userEmail string, switchPanelId int, position models.Position) (*models.SwitchPanel, error)
	DelSwitchPanel(ctx context.Context, userEmail string, switchPanelId int) error
}
