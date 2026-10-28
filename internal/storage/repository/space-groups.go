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

	GetSpaceGroupMembers(ctx context.Context, userEmail string, spaceGroupId int) (*models.SpaceGroupMembers, error)
	AddSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId int, memberEmail, role string) error
	DelSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId int, memberEmail string) error
	SetSpaceGroupMemberRole(ctx context.Context, userEmail string, spaceGroupId int, memberEmail, role string) error
	TransferSpaceGroupOwner(ctx context.Context, userEmail string, spaceGroupId int, memberEmail string) error
	SetSpaceGroupPublic(ctx context.Context, userEmail string, spaceGroupId int, public bool) error
	LeaveSpaceGroup(ctx context.Context, userEmail string, spaceGroupId int) error
}
