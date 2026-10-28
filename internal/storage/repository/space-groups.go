package repository

import (
	"acrocuit/internal/models"
	"context"
)

type SpaceGroups interface {
	GetSpaceGroups(ctx context.Context, userEmail, include string) ([]models.SpaceGroup, error)
	GetSpaceGroup(ctx context.Context, userEmail string, spaceGroupId string, include string) (*models.SpaceGroup, error)
	AddSpaceGroup(ctx context.Context, userEmail, name string) (string, error)
	EditSpaceGroupName(ctx context.Context, userEmail string, spaceGroupId string, name string) (*models.SpaceGroup, error)
	DelSpaceGroup(ctx context.Context, userEmail string, spaceGroupId string) error

	GetSpaceGroupMembers(ctx context.Context, userEmail string, spaceGroupId string) (*models.SpaceGroupMembers, error)
	AddSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId string, memberEmail, role string) error
	DelSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId string, memberEmail string) error
	SetSpaceGroupMemberRole(ctx context.Context, userEmail string, spaceGroupId string, memberEmail, role string) error
	TransferSpaceGroupOwner(ctx context.Context, userEmail string, spaceGroupId string, memberEmail string) error
	SetSpaceGroupPublic(ctx context.Context, userEmail string, spaceGroupId string, public bool) error
	LeaveSpaceGroup(ctx context.Context, userEmail string, spaceGroupId string) error
}
