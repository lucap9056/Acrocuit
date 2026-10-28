package postgres

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"acrocuit/schema"
	"context"
)

const spaceGroupOwnedBy = `EXISTS (
	SELECT 1 FROM ` + schema.SpaceGroupACLsTable + ` owner_acl
	WHERE owner_acl.` + schema.SpaceGroupACLs_SPACE_GROUP_ID + ` = $1
		AND owner_acl.` + schema.SpaceGroupACLs_USER_EMAIL + ` = $2
		AND owner_acl.` + schema.SpaceGroupACLs_ROLE + ` = '` + schema.SpaceGroupACLs_ROLE_OWNER + `'
)`

func (s *pgSpaceGroups) GetSpaceGroupMembers(ctx context.Context, userEmail string, spaceGroupId string) (*models.SpaceGroupMembers, error) {
	query := `
		SELECT ` + schema.FullSpaceGroupACLs_USER_EMAIL + `, ` + schema.FullSpaceGroupACLs_ROLE + `
		FROM ` + schema.SpaceGroupACLsTable + `
		WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1 AND ` + spaceGroupOwnedBy + `
		ORDER BY
			CASE ` + schema.FullSpaceGroupACLs_ROLE + `
				WHEN '` + schema.SpaceGroupACLs_ROLE_OWNER + `' THEN 0
				WHEN '` + schema.SpaceGroupACLs_ROLE_EDITOR + `' THEN 1
				ELSE 2
			END,
			` + schema.FullSpaceGroupACLs_USER_EMAIL

	rows, err := s.pool.Query(ctx, query, spaceGroupId, userEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &models.SpaceGroupMembers{Members: make([]models.SpaceGroupMember, 0)}
	found := false
	for rows.Next() {
		var member models.SpaceGroupMember
		if err := rows.Scan(&member.Email, &member.Role); err != nil {
			return nil, err
		}
		found = true
		if member.Email == models.PublicAccessEmail {
			result.Public = true
			continue
		}
		result.Members = append(result.Members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, repository.ErrNotFound
	}

	return result, nil
}

func (s *pgSpaceGroups) AddSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId string, memberEmail, role string) error {
	if memberEmail == models.PublicAccessEmail {
		return repository.ErrInvalidMember
	}
	if role != schema.SpaceGroupACLs_ROLE_VIEWER && role != schema.SpaceGroupACLs_ROLE_EDITOR {
		return repository.ErrInvalidRole
	}

	query := `
		WITH permitted AS (
			SELECT 1 WHERE ` + spaceGroupOwnedBy + `
		),
		inserted AS (
			INSERT INTO ` + schema.SpaceGroupACLsTable + ` (` + schema.SpaceGroupACLs_USER_EMAIL + `, ` + schema.SpaceGroupACLs_SPACE_GROUP_ID + `, ` + schema.SpaceGroupACLs_ROLE + `)
			SELECT $3, $1, $4 FROM permitted
			ON CONFLICT (` + schema.SpaceGroupACLs_USER_EMAIL + `, ` + schema.SpaceGroupACLs_SPACE_GROUP_ID + `) DO NOTHING
			RETURNING 1
		)
		SELECT EXISTS (SELECT 1 FROM permitted), EXISTS (SELECT 1 FROM inserted)
	`
	var permitted, inserted bool
	if err := s.pool.QueryRow(ctx, query, spaceGroupId, userEmail, memberEmail, role).Scan(&permitted, &inserted); err != nil {
		return err
	}
	if !permitted {
		return repository.ErrNotFound
	}
	if !inserted {
		return repository.ErrConflict
	}

	return nil
}

func (s *pgSpaceGroups) DelSpaceGroupMember(ctx context.Context, userEmail string, spaceGroupId string, memberEmail string) error {
	if memberEmail == models.PublicAccessEmail {
		return repository.ErrInvalidMember
	}

	query := `
		DELETE FROM ` + schema.SpaceGroupACLsTable + `
		WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
			AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3
			AND ` + schema.FullSpaceGroupACLs_ROLE + ` <> '` + schema.SpaceGroupACLs_ROLE_OWNER + `'
			AND ` + spaceGroupOwnedBy
	t, err := s.pool.Exec(ctx, query, spaceGroupId, userEmail, memberEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSpaceGroups) SetSpaceGroupMemberRole(ctx context.Context, userEmail string, spaceGroupId string, memberEmail, role string) error {
	if memberEmail == models.PublicAccessEmail {
		return repository.ErrInvalidMember
	}
	if role != schema.SpaceGroupACLs_ROLE_VIEWER && role != schema.SpaceGroupACLs_ROLE_EDITOR {
		return repository.ErrInvalidRole
	}

	query := `
		UPDATE ` + schema.SpaceGroupACLsTable + `
		SET ` + schema.SpaceGroupACLs_ROLE + ` = $4
		WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
			AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3
			AND ` + schema.FullSpaceGroupACLs_ROLE + ` <> '` + schema.SpaceGroupACLs_ROLE_OWNER + `'
			AND ` + spaceGroupOwnedBy
	t, err := s.pool.Exec(ctx, query, spaceGroupId, userEmail, memberEmail, role)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSpaceGroups) TransferSpaceGroupOwner(ctx context.Context, userEmail string, spaceGroupId string, memberEmail string) error {
	if memberEmail == models.PublicAccessEmail {
		return repository.ErrInvalidMember
	}

	query := `
		UPDATE ` + schema.SpaceGroupACLsTable + `
		SET ` + schema.SpaceGroupACLs_ROLE + ` = CASE
			WHEN ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2 THEN '` + schema.SpaceGroupACLs_ROLE_EDITOR + `'
			ELSE '` + schema.SpaceGroupACLs_ROLE_OWNER + `'
		END
		WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
			AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` IN ($2, $3)
			AND ` + spaceGroupOwnedBy + `
			AND EXISTS (
				SELECT 1 FROM ` + schema.SpaceGroupACLsTable + ` target_acl
				WHERE target_acl.` + schema.SpaceGroupACLs_SPACE_GROUP_ID + ` = $1
					AND target_acl.` + schema.SpaceGroupACLs_USER_EMAIL + ` = $3
					AND target_acl.` + schema.SpaceGroupACLs_ROLE + ` <> '` + schema.SpaceGroupACLs_ROLE_OWNER + `'
			)`
	t, err := s.pool.Exec(ctx, query, spaceGroupId, userEmail, memberEmail)
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSpaceGroups) SetSpaceGroupPublic(ctx context.Context, userEmail string, spaceGroupId string, public bool) error {
	changePublicAccess := `
		DELETE FROM ` + schema.SpaceGroupACLsTable + `
		WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
			AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $3
			AND EXISTS (SELECT 1 FROM permitted)`
	if public {
		changePublicAccess = `
		INSERT INTO ` + schema.SpaceGroupACLsTable + ` (` + schema.SpaceGroupACLs_USER_EMAIL + `, ` + schema.SpaceGroupACLs_SPACE_GROUP_ID + `, ` + schema.SpaceGroupACLs_ROLE + `)
		SELECT $3, $1, '` + schema.SpaceGroupACLs_ROLE_VIEWER + `' FROM permitted
		ON CONFLICT (` + schema.SpaceGroupACLs_USER_EMAIL + `, ` + schema.SpaceGroupACLs_SPACE_GROUP_ID + `) DO NOTHING`
	}

	query := `
		WITH permitted AS (
			SELECT 1 WHERE ` + spaceGroupOwnedBy + `
		),
		changed AS (` + changePublicAccess + `
		)
		SELECT EXISTS (SELECT 1 FROM permitted)
	`
	var permitted bool
	if err := s.pool.QueryRow(ctx, query, spaceGroupId, userEmail, models.PublicAccessEmail).Scan(&permitted); err != nil {
		return err
	}
	if !permitted {
		return repository.ErrNotFound
	}

	return nil
}

func (s *pgSpaceGroups) LeaveSpaceGroup(ctx context.Context, userEmail string, spaceGroupId string) error {
	if userEmail == models.PublicAccessEmail {
		return repository.ErrAnonymous
	}

	query := `
		WITH member AS (
			SELECT ` + schema.FullSpaceGroupACLs_ROLE + `
			FROM ` + schema.SpaceGroupACLsTable + `
			WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1 AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2
			FOR UPDATE
		),
		successor AS (
			SELECT ` + schema.FullSpaceGroupACLs_USER_EMAIL + `
			FROM ` + schema.SpaceGroupACLsTable + `
			WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
				AND ` + schema.FullSpaceGroupACLs_ROLE + ` = '` + schema.SpaceGroupACLs_ROLE_EDITOR + `'
				AND EXISTS (SELECT 1 FROM member WHERE ` + schema.SpaceGroupACLs_ROLE + ` = '` + schema.SpaceGroupACLs_ROLE_OWNER + `')
			ORDER BY random()
			LIMIT 1
			FOR UPDATE
		),
		promoted AS (
			UPDATE ` + schema.SpaceGroupACLsTable + `
			SET ` + schema.SpaceGroupACLs_ROLE + ` = '` + schema.SpaceGroupACLs_ROLE_OWNER + `'
			FROM successor
			WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
				AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = successor.` + schema.SpaceGroupACLs_USER_EMAIL + `
		),
		removed AS (
			DELETE FROM ` + schema.SpaceGroupACLsTable + `
			WHERE ` + schema.FullSpaceGroupACLs_SPACE_GROUP_ID + ` = $1
				AND ` + schema.FullSpaceGroupACLs_USER_EMAIL + ` = $2
				AND (
					EXISTS (SELECT 1 FROM member WHERE ` + schema.SpaceGroupACLs_ROLE + ` <> '` + schema.SpaceGroupACLs_ROLE_OWNER + `')
					OR EXISTS (SELECT 1 FROM successor)
				)
		)
		SELECT
			EXISTS (SELECT 1 FROM member),
			EXISTS (SELECT 1 FROM member WHERE ` + schema.SpaceGroupACLs_ROLE + ` = '` + schema.SpaceGroupACLs_ROLE_OWNER + `'),
			EXISTS (SELECT 1 FROM successor)
	`
	var isMember, isOwner, hasSuccessor bool
	if err := s.pool.QueryRow(ctx, query, spaceGroupId, userEmail).Scan(&isMember, &isOwner, &hasSuccessor); err != nil {
		return err
	}
	if !isMember {
		return repository.ErrNotFound
	}
	if isOwner && !hasSuccessor {
		return repository.ErrConflict
	}

	return nil
}
