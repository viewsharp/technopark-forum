package forum

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrNotFoundUser        = errors.New("not found user")
	ErrUniqueViolation = errors.New("unique violation")
	ErrUnknown         = errors.New("unknown error")
)
