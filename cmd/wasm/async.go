//go:build js

package main

import (
	"acrocuit/internal/models"
	"acrocuit/internal/storage/repository"
	"errors"
	"syscall/js"
)

const (
	statusBadRequest          = 400
	statusNotFound            = 404
	statusConflict            = 409
	statusInternalServerError = 500
	statusNotImplemented      = 501
)

func asyncFunc(fn func(args []js.Value) (any, error)) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any {
		executor := js.FuncOf(func(_ js.Value, settle []js.Value) any {
			resolve, reject := settle[0], settle[1]
			go func() {
				value, err := fn(args)
				if err != nil {
					reject.Invoke(storageError(err))
					return
				}

				result, err := toJS(value)
				if err != nil {
					reject.Invoke(storageError(err))
					return
				}
				resolve.Invoke(result)
			}()
			return nil
		})
		defer executor.Release()

		return js.Global().Get("Promise").New(executor)
	})
}

func storageError(err error) map[string]any {
	status := statusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, repository.ErrNotFound):
		status, message = statusNotFound, "not found"
	case errors.Is(err, repository.ErrConflict):
		status, message = statusConflict, "conflict"
	case errors.Is(err, repository.ErrSelfReference), errors.Is(err, repository.ErrCycleDetected), errors.Is(err, models.ErrInvalidColumn),
		errors.Is(err, repository.ErrInvalidRole), errors.Is(err, repository.ErrInvalidMember), errors.Is(err, errInvalidArgument):
		status, message = statusBadRequest, err.Error()
	case errors.Is(err, repository.ErrNotSupported):
		status, message = statusNotImplemented, err.Error()
	default:
		js.Global().Get("console").Call("error", err.Error())
	}

	return map[string]any{"status": status, "message": message}
}
