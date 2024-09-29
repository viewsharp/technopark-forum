package post

import "errors"

type ErrNotFoundUser struct {
	Nickname string
}

func (e ErrNotFoundUser) Error() string {
	return "Not found user: " + e.Nickname
}

var (
	ErrInvalidParent  = errors.New("invalid parent")
	ErrNotFoundThread = errors.New("not found thread")
	ErrNotFound       = errors.New("not found")
)
