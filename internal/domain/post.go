package domain

import "time"

type Post struct {
	Message  string
	Author   string
	Created  *time.Time
	Forum    *string
	Id       *int64
	IsEdited *bool
	Parent   *int64
	Thread   *int32
}

type PostFull struct {
	Author *User
	Forum  *Forum
	Post   *Post
	Thread *Thread
}

type PostUpdate struct {
	Message *string
}
