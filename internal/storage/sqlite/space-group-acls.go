package sqlite

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"context"
)

func (s *sqliteSpaceGroups) GetSpaceGroupMembers(context.Context, string, string) (*models.SpaceGroupMembers, error) {
	return nil, repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) AddSpaceGroupMember(context.Context, string, string, string, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) DelSpaceGroupMember(context.Context, string, string, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) SetSpaceGroupMemberRole(context.Context, string, string, string, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) TransferSpaceGroupOwner(context.Context, string, string, string) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) SetSpaceGroupPublic(context.Context, string, string, bool) error {
	return repository.ErrNotSupported
}

func (s *sqliteSpaceGroups) LeaveSpaceGroup(context.Context, string, string) error {
	return repository.ErrNotSupported
}
