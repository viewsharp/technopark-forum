package user

import "errors"

var (
	ErrUniqueViolation = errors.New("unique violation")
	ErrNotFound        = errors.New("not found")
	ErrNotFoundForum   = errors.New("not found forum")
)
