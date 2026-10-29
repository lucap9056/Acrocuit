package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"context"
)

func (s *sqliteSpaceGroups) GetSpaceGroupMembers(context.Context, string, models.SpaceGroupId) (*models.SpaceGroupMembers, error) {
	return nil, repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) AddSpaceGroupMember(context.Context, string, models.SpaceGroupId, string, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) DelSpaceGroupMember(context.Context, string, models.SpaceGroupId, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) SetSpaceGroupMemberRole(context.Context, string, models.SpaceGroupId, string, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) TransferSpaceGroupOwner(context.Context, string, models.SpaceGroupId, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) SetSpaceGroupPublic(context.Context, string, models.SpaceGroupId, bool) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) LeaveSpaceGroup(context.Context, string, models.SpaceGroupId) error {
	return repository.ErrNotSupported
}
