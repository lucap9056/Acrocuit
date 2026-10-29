package repository

import (
	"acrocuit/internal/models"
	"context"
)

type SpaceGroups interface {
	GetSpaceGroups(ctx context.Context, userEmail, include string) ([]models.SpaceGroup, error)
	GetSpaceGroup(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, include string) (*models.SpaceGroup, error)
	AddSpaceGroup(ctx context.Context, userEmail, name string) (models.SpaceGroupId, error)
	EditSpaceGroupName(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, name string) (*models.SpaceGroup, error)
	DelSpaceGroup(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId) error

	GetSpaceGroupMembers(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId) (*models.SpaceGroupMembers, error)
	AddSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, memberEmail, role string) error
	DelSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, memberEmail string) error
	SetSpaceGroupMemberRole(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, memberEmail, role string) error
	TransferSpaceGroupOwner(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, memberEmail string) error
	SetSpaceGroupPublic(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId, public bool) error
	LeaveSpaceGroup(ctx context.Context, userEmail string, spaceGroupId models.SpaceGroupId) error
}
