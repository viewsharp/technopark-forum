package thread

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrNotFoundForum   = errors.New("not found forum")
	ErrNotFoundUser    = errors.New("not found user")
	ErrUniqueViolation = errors.New("unique violation")
)
