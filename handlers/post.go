package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/resources/post"
	"github.com/viewsharp/TexPark_DBMSs/resources/user"
	"strconv"
)

type PostHandler struct {
	sb *StorageBundle
}

func NewPostHandler(storageBundle *StorageBundle) *PostHandler {
	return &PostHandler{sb: storageBundle}
}

func (ph *PostHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	posts := make(post.Posts, 0, 1)
	err := posts.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	if len(posts) == 0 {
		return posts, fasthttp.StatusCreated
	}

	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, err := strconv.Atoi(slugOrId)
	if err == nil {
		err = ph.sb.post.AddByThreadId(&posts, threadId)
	} else {
		err = ph.sb.post.AddByThreadSlug(&posts, slugOrId)
	}

	switch err {
	case nil:
		return posts, fasthttp.StatusCreated
	}

	return nil, fasthttp.StatusInternalServerError
}

func (ph *PostHandler) Get(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	idString := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	result, err := ph.sb.post.ById(postId)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case user.ErrNotFound:
		return Error{
			Message: "Can't find user by nickname: ",
		}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (ph *PostHandler) GetByThread(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, threadIdParseErr := strconv.Atoi(slugOrId)

	limit := 1000
	limitParam := ctx.QueryArgs().Peek("limit")
	if limitParam != nil {
		var err error
		limit, err = strconv.Atoi(string(limitParam))
		if err != nil {
			return nil, fasthttp.StatusBadRequest
		}
	}

	desc := false
	descParam := ctx.QueryArgs().Peek("desc")
	if descParam != nil {
		desc = string(descParam) == "true"
	}

	since := 0
	sinceParam := ctx.QueryArgs().Peek("since")
	if sinceParam != nil {
		var err error
		since, err = strconv.Atoi(string(sinceParam))
		if err != nil {
			return nil, fasthttp.StatusBadRequest
		}
	}


	var err error
	var posts post.Posts
	switch string(ctx.QueryArgs().Peek("sort")) {
	case "tree":
		if threadIdParseErr == nil {
			posts, err = ph.sb.post.TreeByThreadId(threadId, limit, desc, since)
		} else {
			posts, err = ph.sb.post.TreeByThreadSlug(slugOrId, limit, desc, since)
		}
	case "parent_tree":
		if threadIdParseErr == nil {
			posts, err = ph.sb.post.ParentTreeByThreadId(threadId, limit, desc, since)
		} else {
			posts, err = ph.sb.post.ParentTreeByThreadSlug(slugOrId, limit, desc, since)
		}
	default:
		if threadIdParseErr == nil {
			posts, err = ph.sb.post.FlatByThreadId(threadId, limit, desc, since)
		} else {
			posts, err = ph.sb.post.FlatByThreadSlug(slugOrId, limit, desc, since)
		}
	}

	if err == nil {
		return posts, fasthttp.StatusOK
	}

	return nil, fasthttp.StatusInternalServerError
}
