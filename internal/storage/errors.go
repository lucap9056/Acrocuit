package storage

import "errors"

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrConflict      = errors.New("storage: conflict")
	ErrSelfReference = errors.New("storage: breaker cannot reference itself as upstream")
	ErrCycleDetected = errors.New("storage: upstream breaker would create a cycle")
)
