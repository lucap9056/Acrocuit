package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Spaces interface {
	GetSpaces(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, include string) ([]models.Space, error)
	GetSpace(ctx context.Context, userEmail string, spaceId models.SpaceId, include string) (*models.Space, error)
	AddSpace(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, name string) (*models.Space, error)
	SetSpace(ctx context.Context, userEmail string, spaceId models.SpaceId, name *string) (*models.Space, error)
	MoveSpace(ctx context.Context, userEmail string, spaceId models.SpaceId, displayOrder int) ([]models.Space, error)
	DelSpace(ctx context.Context, userEmail string, spaceId models.SpaceId) error
	GetSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId models.SpaceId) (*models.BackgroundImage, error)
	SetSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId models.SpaceId, contentType string, data []byte) (*models.Space, error)
	DelSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId models.SpaceId) (*models.Space, error)
}
