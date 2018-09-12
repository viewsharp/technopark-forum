package main

import "time"

type Error struct {
	Message *string `json:"message"`
}

type Forum struct {
	Posts   int32  `json:"posts,omitempty"`
	Slug    string `json:"slug"`
	Threads int32  `json:"threads,omitempty"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

type Post struct {
	Author   *string    `json:"author"`
	Created  *time.Time `json:"created,omitempty"`
	Forum    *string    `json:"forum,omitempty"`
	Id       *int32     `json:"id,omitempty"`
	IsEdited *bool      `json:"isEdited,omitempty"`
	Message  *string    `json:"message"`
	Parent   *int32     `json:"parent,omitempty"`
	Thread   *int32     `json:"thread,omitempty"`
}

//easyjson:json
type Posts []Post

type PostFull struct {
	Author *User
	Forum  *Forum
	Post   *Post
	Thread *Thread
}

type PostUpdate struct {
	Message *string `json:"message,omitempty"`
}

type Status struct {
	Forum  *int32 `json:"forum"`
	Post   *int32 `json:"post"`
	Thread *int32 `json:"thread"`
	User   *int32 `json:"user"`
}

type Thread struct {
	Author  *string    `json:"author"`
	Created *time.Time `json:"created,omitempty"`
	Forum   *string    `json:"forum,omitempty"`
	Id      *int32     `json:"id,omitempty"`
	Message *string    `json:"message"`
	Slug    *string    `json:"slug,omitempty"`
	Title   *string    `json:"title"`
	Votes   *int32     `json:"votes,omitempty"`
}

//easyjson:json
type Threads []*Thread

type ThreadUpdate struct {
	Message *string `json:"message,omitempty"`
	Title   *string `json:"title,omitempty"`
}

type User struct {
	About    *string `json:"about,omitempty"`
	Email    *string `json:"email"`
	FullName *string `json:"fullname"`
	Nickname *string `json:"nickname,omitempty"`
}

//easyjson:json
type Users []*User

type UserUpdate struct {
	About    *string `json:"about,omitempty"`
	Email    *string `json:"email,omitempty"`
	FullName *string `json:"fullname,omitempty"`
}

type Vote struct {
	Nickname *string `json:"nickname"`
	Voice    *int32  `json:"voice"`
}
