package vote

import "errors"

var (
	ErrNotFoundThread  = errors.New("not found thread")
	ErrNotFoundUser    = errors.New("not found user")
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
	ErrUnknown         = errors.New("unknown error")
)
