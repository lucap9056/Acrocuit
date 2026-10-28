package repository

import "errors"

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrConflict      = errors.New("storage: conflict")
	ErrSelfReference = errors.New("storage: switch cannot reference itself as upstream")
	ErrCycleDetected = errors.New("storage: upstream switch would create a cycle")
	ErrInvalidRole   = errors.New("storage: role must be viewer or editor")
	ErrInvalidMember = errors.New("storage: invalid member email")
	ErrNotSupported  = errors.New("storage: not supported by this backend")
	ErrAnonymous     = errors.New("storage: anonymous requests are not allowed")
)
