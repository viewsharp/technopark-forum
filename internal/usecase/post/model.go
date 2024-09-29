package post

import (
	"time"

	"github.com/viewsharp/technopark-forum/internal/usecase/forum"
	"github.com/viewsharp/technopark-forum/internal/usecase/thread"
	"github.com/viewsharp/technopark-forum/internal/usecase/user"
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

type PostFull struct {
	Author *user.User     `json:"author,omitempty"`
	Forum  *forum.Forum   `json:"forum,omitempty"`
	Post   *Post          `json:"post,omitempty"`
	Thread *thread.Thread `json:"thread,omitempty"`
}

type PostUpdate struct {
	Message *string `json:"message,omitempty"`
}
