package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Spaces interface {
	GetSpaces(ctx context.Context, userEmail string, spaceGroupId string, include string) ([]models.Space, error)
	GetSpace(ctx context.Context, userEmail string, spaceId string, include string) (*models.Space, error)
	AddSpace(ctx context.Context, userEmail string, spaceGroupId string, name string) (*models.Space, error)
	SetSpace(ctx context.Context, userEmail string, spaceId string, name *string) (*models.Space, error)
	MoveSpace(ctx context.Context, userEmail string, spaceId string, displayOrder int) ([]models.Space, error)
	DelSpace(ctx context.Context, userEmail string, spaceId string) error
	GetSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId string) (*models.BackgroundImage, error)
	SetSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId string, contentType string, data []byte) (*models.Space, error)
	DelSpaceBackgroundImage(ctx context.Context, userEmail string, spaceId string) (*models.Space, error)
}
