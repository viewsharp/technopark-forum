package main

type Error struct {
	Message string `json:"message"`
}

type Forum struct {
	Posts   int    `json:"posts,omitempty"`
	Slug    string `json:"slug"`
	Threads int    `json:"threads,omitempty"`
	Title   int    `json:"title"`
	User    string `json:"user"`
}

type Post struct {
	Author   string `json:"author"`
	Created  string `json:"created,omitempty"`
	Forum    string `json:"forum,omitempty"`
	Id       int    `json:"id,omitempty"`
	IsEdited bool   `json:"isEdited,omitempty"`
	Message  string `json:"message"`
	Parent   int    `json:"parent,omitempty"`
	Thread   int    `json:"thread,omitempty"`
}

type PostFull struct {
	Author User
	Forum Forum
	Post Post
	Thread Thread
}

type PostUpdate struct {
	Message string `json:"message,omitempty"`
}

type Status struct {
	Forum  int `json:"forum"`
	Post   int `json:"post"`
	Thread int `json:"thread"`
	User   int `json:"user"`
}

type Thread struct {
	Author  string `json:"author"`
	Created string `json:"created,omitempty"`
	Forum   string `json:"forum,omitempty"`
	Id      int    `json:"id,omitempty"`
	Message string `json:"message"`
	Slug    string `json:"slug,omitempty"`
	Title   string `json:"title"`
	Votes   int    `json:"votes,omitempty"`
}

type ThreadUpdate struct {
	Message string `json:"message,omitempty"`
	Title   string `json:"title,omitempty"`
}

type User struct {
	About    string `json:"about,omitempty"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	NickName string `json:"nickname,omitempty"`
}

type UserUpdate struct {
	About    string `json:"about,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullname,omitempty"`
}

type Vote struct {
	Nickname string `json:"nickname"`
	Voice int `json:"voice"`
}