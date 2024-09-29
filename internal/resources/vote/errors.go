package vote

import "errors"

var (
	ErrNotFoundThread = errors.New("not found thread")
	ErrNotFoundUser   = errors.New("not found user")
)
