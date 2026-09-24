package storage

import (
	"acrocuit/internal/models"
	"acrocuit/schema"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Users interface {
	CreateUser(ctx context.Context, username string, passwordSalt, passwordHash []byte) (int, error)
	GetUserSecret(ctx context.Context, username string) (*models.UserSecret, error)
}

type pgUsers struct {
	pool *pgxpool.Pool
}

func (u *pgUsers) CreateUser(ctx context.Context, username string, passwordSalt, passwordHash []byte) (int, error) {
	query := `
		WITH new_user AS (
			INSERT INTO ` + schema.UsersTable + ` (` + schema.Users_USERNAME + `) VALUES ($1) RETURNING ` + schema.Users_ID + `
		)
		INSERT INTO ` + schema.UserSecretsTable + ` (` + schema.UserSecrets_USER_ID + `, ` + schema.UserSecrets_PASSWORD_SALT + `, ` + schema.UserSecrets_PASSWORD_HASH + `)
		SELECT ` + schema.Users_ID + `, $2, $3 FROM new_user
		RETURNING ` + schema.UserSecrets_USER_ID + `
	`
	var userId int
	err := u.pool.QueryRow(ctx, query, username, passwordSalt, passwordHash).Scan(&userId)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrConflict
	}
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (u *pgUsers) GetUser(ctx context.Context, userId int) (*models.User, error) {
	keys, fields, err := models.ParseUserColumns("")
	if err != nil {
		return nil, err
	}

	query := `
		SELECT ` + keys + `
		FROM ` + schema.UsersTable + `
		WHERE ` + schema.FullUsers_ID + `=$1
	`

	user := &models.User{}
	if err := u.pool.QueryRow(ctx, query, userId).Scan(user.Values(fields)...); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *pgUsers) GetUserSecret(ctx context.Context, username string) (*models.UserSecret, error) {
	keys, fields, err := models.ParseUserSecretColumns("")
	if err != nil {
		return nil, err
	}
	query := `
		SELECT ` + keys + `
		FROM ` + schema.UsersTable + `
		JOIN ` + schema.UserSecretsTable + ` ON ` + schema.JoinUserSecrets_Users + `
		WHERE ` + schema.FullUsers_USERNAME + `=$1
	`

	secret := &models.UserSecret{}
	if err := u.pool.QueryRow(ctx, query, username).Scan(secret.Values(fields)...); err != nil {
		return nil, err
	}

	return secret, nil
}
