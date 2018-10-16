package post

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
	ErrUnknown         = errors.New("unknown error")
)