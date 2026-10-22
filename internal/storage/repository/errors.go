package repository

import "errors"

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrConflict      = errors.New("storage: conflict")
	ErrSelfReference = errors.New("storage: switch cannot reference itself as upstream")
	ErrCycleDetected = errors.New("storage: upstream switch would create a cycle")
)
