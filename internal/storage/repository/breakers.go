package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Breakers interface {
	GetBreakers(ctx context.Context, userEmail string, breakerGroupId int, include string) ([]models.Breaker, error)
	GetBreaker(ctx context.Context, userEmail string, breakerId int, include string) (*models.Breaker, error)
	AddBreaker(ctx context.Context, userEmail string, breakerGroupId int, name string, upstreamBreakerId *int) (*models.Breaker, error)
	SetBreaker(ctx context.Context, userEmail string, breakerId int, name *string, displayOrder *int) (*models.Breaker, error)
	SetBreakerUpstream(ctx context.Context, userEmail string, breakerId int, upstreamBreakerId *int) (*models.Breaker, error)
	DelBreaker(ctx context.Context, userEmail string, breakerId int) error
	GetBreakerDownstream(ctx context.Context, userEmail string, breakerId int) ([]models.Breaker, []models.Device, error)
}
