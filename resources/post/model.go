package post

import "time"

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
type Posts []*Post

//type PostFull struct {
//	Author *User
//	Forum  *Forum
//	Post   *Post
//	Thread *Thread
//}

type PostUpdate struct {
	Message *string `json:"message,omitempty"`
}