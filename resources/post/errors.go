package post

import "errors"

type ErrNotFoundUserClass struct {
	s        string
	nickname string
}

func (e *ErrNotFoundUserClass) setNickname(nickname string) {
	e.nickname = nickname
}

func (e ErrNotFoundUserClass) GetNickname() string {
	return e.nickname
}

func (e ErrNotFoundUserClass) Error() string {
	return e.s
}

var (
	ErrInvalidParent   = errors.New("invalid parent")
	ErrNotFoundThread    = ErrNotFoundUserClass{s: "not found thread"}
	ErrNotFoundUser    = ErrNotFoundUserClass{s: "not found user"}
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
	ErrUnknown         = errors.New("unknown error")
)
