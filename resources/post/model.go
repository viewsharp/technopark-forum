package post

import (
	"github.com/viewsharp/TexPark_DBMSs/resources/forum"
	"github.com/viewsharp/TexPark_DBMSs/resources/thread"
	"github.com/viewsharp/TexPark_DBMSs/resources/user"
	"time"
)

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

type PostFull struct {
	Author *user.User     `json:"author,omitempty"`
	Forum  *forum.Forum   `json:"forum,omitempty"`
	Post   *Post          `json:"post,omitempty"`
	Thread *thread.Thread `json:"thread,omitempty"`
}

type PostUpdate struct {
	Message *string `json:"message,omitempty"`
}
