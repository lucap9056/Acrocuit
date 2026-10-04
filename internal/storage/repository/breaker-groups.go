package repository

import (
	"acrocuit/internal/models"
	"context"
)

type BreakerGroups interface {
	GetBreakerGroups(ctx context.Context, userEmail string, spaceId int, include string) ([]models.BreakerGroup, error)
	GetBreakerGroup(ctx context.Context, userEmail string, breakerGroupId int, include string) (*models.BreakerGroup, error)
	AddBreakerGroup(ctx context.Context, userEmail string, spaceId int, name string, position *models.Position) (*models.BreakerGroup, error)
	SetBreakerGroup(ctx context.Context, userEmail string, breakerGroupId int, name string) (*models.BreakerGroup, error)
	SetBreakerGroupPosition(ctx context.Context, userEmail string, breakerGroupId int, position models.Position) (*models.BreakerGroup, error)
	DelBreakerGroup(ctx context.Context, userEmail string, breakerGroupId int) error
}
