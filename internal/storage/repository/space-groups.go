package repository

import (
	"acrocuit/internal/models"
	"context"
)

type SpaceGroups interface {
	GetSpaceGroups(ctx context.Context, userEmail, include string) ([]models.SpaceGroup, error)
	GetSpaceGroup(ctx context.Context, userEmail string, spaceGroupId int, include string) (*models.SpaceGroup, error)
	AddSpaceGroup(ctx context.Context, userEmail, name string) (int, error)
	EditSpaceGroupName(ctx context.Context, userEmail string, spaceGroupId int, name string) error
	DelSpaceGroup(ctx context.Context, userEmail string, spaceGroupId int) error
}
