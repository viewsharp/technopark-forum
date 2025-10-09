package domain

import "errors"

// Common errors
var (
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
)

// User errors
var (
	ErrUserNotFoundForum = errors.New("not found forum")
)

// Forum errors
var (
	ErrForumNotFoundUser = errors.New("not found user")
)

// Thread errors
var (
	ErrThreadNotFoundForum = errors.New("not found forum")
	ErrThreadNotFoundUser  = errors.New("not found user")
)

// Post errors
var (
	ErrPostInvalidParent  = errors.New("invalid parent")
	ErrPostNotFoundThread = errors.New("not found thread")
)

type ErrPostNotFoundUser struct {
	Nickname string
}

func (e ErrPostNotFoundUser) Error() string {
	return "Not found user: " + e.Nickname
}

// Vote errors
var (
	ErrVoteNotFoundThread = errors.New("not found thread")
	ErrVoteNotFoundUser   = errors.New("not found user")
)
