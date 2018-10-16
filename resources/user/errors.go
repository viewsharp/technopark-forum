package user

import "errors"

var (
	ErrUniqueViolation = errors.New("unique violation")
	ErrUnknown = errors.New("unknown error")
	ErrNotFound = errors.New("not found")
	ErrNotFoundForum = errors.New("not found forum")
)