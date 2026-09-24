package models

import (
	"acrocuit/schema"
	"fmt"
	"strings"
)

const USER_ALL_COLUMNS = schema.FullUsers_ID + "," + schema.FullUsers_USERNAME

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

func ParseUserColumns(columnsStr string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return USER_ALL_COLUMNS, []uint8{0, 1, 2}, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.Users_ID, schema.FullUsers_ID:
			collector.Append(0, schema.FullUsers_ID)
		case schema.Users_USERNAME, schema.FullUsers_USERNAME:
			collector.Append(1, schema.FullUsers_USERNAME)
		default:
			return "", nil, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, nil
}

func (u *User) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &u.Id
		case 1:
			values[i] = &u.Username
		}
	}
	return values
}

const USER_SECRET_ALL_COLUMNS = schema.FullUserSecrets_USER_ID + "," + schema.FullUserSecrets_PASSWORD_SALT + "," + schema.FullUserSecrets_PASSWORD_HASH

type UserSecret struct {
	UserId       int    `json:"-"`
	PasswordSalt []byte `json:"-"`
	PasswordHash []byte `json:"-"`
}

func ParseUserSecretColumns(columnsStr string) (string, []uint8, error) {
	if strings.TrimSpace(columnsStr) == "" {
		return USER_SECRET_ALL_COLUMNS, []uint8{0, 1, 2}, nil
	}

	collector := newColumnCollector(columnsStr)
	defer collector.Release()

	scanner := newColumnScanner(columnsStr)

	for scanner.Next() {
		col := scanner.Text()
		switch col {
		case schema.UserSecrets_USER_ID, schema.FullUserSecrets_USER_ID:
			collector.Append(0, schema.FullUserSecrets_USER_ID)
		case schema.UserSecrets_PASSWORD_SALT, schema.FullUserSecrets_PASSWORD_SALT:
			collector.Append(1, schema.FullUserSecrets_PASSWORD_SALT)
		case schema.UserSecrets_PASSWORD_HASH, schema.FullUserSecrets_PASSWORD_HASH:
			collector.Append(2, schema.FullUserSecrets_PASSWORD_HASH)
		default:
			return "", nil, fmt.Errorf("invalid column: %s", col)
		}
	}

	keys, fields := collector.Result()
	return keys, fields, nil
}

func (s *UserSecret) Values(fields []uint8) []any {
	values := make([]any, len(fields))
	for i, f := range fields {
		switch f {
		case 0:
			values[i] = &s.UserId
		case 1:
			values[i] = &s.PasswordSalt
		case 2:
			values[i] = &s.PasswordHash
		}
	}
	return values
}
