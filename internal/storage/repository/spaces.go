package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Spaces interface {
	GetSpaces(ctx context.Context, userEmail string, spaceGroupId int, include string) ([]models.Space, error)
	GetSpace(ctx context.Context, userEmail string, spaceId int, include string) (*models.Space, error)
	AddSpace(ctx context.Context, userEmail string, spaceGroupId int, name string) (*models.Space, error)
	SetSpace(ctx context.Context, userEmail string, spaceId int, name *string, displayOrder *int) (*models.Space, error)
	DelSpace(ctx context.Context, userEmail string, spaceId int) error
}
