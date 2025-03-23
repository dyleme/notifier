package apperr

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrUnexpectedType = errors.New("unexpected type")
	ErrEventPastType  = errors.New("event type can not be in the past")
)
