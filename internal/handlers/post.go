package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	post2 "github.com/viewsharp/technopark-forum/internal/resources/post"
	"github.com/viewsharp/technopark-forum/internal/resources/thread"
	"strconv"
	"strings"
)

type PostHandler struct {
	sb *StorageBundle
}

func NewPostHandler(storageBundle *StorageBundle) *PostHandler {
	return &PostHandler{sb: storageBundle}
}

func (ph *PostHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	posts := make(post2.Posts, 0, 1)
	err := posts.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, threadIdParseErr := strconv.Atoi(slugOrId)

	if len(posts) == 0 {
		if threadIdParseErr == nil {
			_, err = ph.sb.thread.ById(threadId)
		} else {
			_, err = ph.sb.thread.BySlug(slugOrId)
		}

		switch err {
		case nil:
			return posts, fasthttp.StatusCreated
		case thread.ErrNotFound:
			if threadIdParseErr == nil {
				return Error{
					Message: fmt.Sprintf("Can't find post thread by id: %d", threadId),
				}, fasthttp.StatusNotFound
			} else {
				return Error{
					Message: "Can't find post thread by slug: " + slugOrId,
				}, fasthttp.StatusNotFound
			}
		}

		return nil, fasthttp.StatusInternalServerError
	}

	if threadIdParseErr == nil {
		err = ph.sb.post.AddByThreadId(&posts, threadId)
	} else {
		err = ph.sb.post.AddByThreadSlug(&posts, slugOrId)
	}

	switch err {
	case nil:
		return posts, fasthttp.StatusCreated
	case post2.ErrInvalidParent:
		return Error{Message: "Parent post was created in another thread"}, fasthttp.StatusConflict
	case post2.ErrNotFoundUser:
		return Error{
			Message: "Can't find post author by nickname: " + err.(post2.ErrNotFoundUserClass).GetNickname(),
		}, fasthttp.StatusNotFound
	case post2.ErrNotFoundThread:
		if threadIdParseErr == nil {
			return Error{
				Message: fmt.Sprintf("Can't find post thread by id: %d", threadId),
			}, fasthttp.StatusNotFound
		} else {
			return Error{
				Message: "Can't find post thread by slug: " + slugOrId,
			}, fasthttp.StatusNotFound
		}
	}

	return nil, fasthttp.StatusInternalServerError
}

func (ph *PostHandler) Get(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	idString := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var result *post2.PostFull

	related := ctx.QueryArgs().Peek("related")

	if related == nil {
		result, err = ph.sb.post.ById(postId, nil)
	} else {
		result, err = ph.sb.post.ById(postId, strings.Split(string(related), ","))
	}

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case post2.ErrNotFound:
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
	var posts post2.Posts
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

	switch err {
	case nil:
		return posts, fasthttp.StatusOK
	case post2.ErrNotFoundThread:
		if threadIdParseErr == nil {
			return Error{
				Message: fmt.Sprintf("Can't find thread by slug: %d", threadId),
			}, fasthttp.StatusNotFound
		} else {
			return Error{
				Message: "Can't find thread by slug: " + slugOrId,
			}, fasthttp.StatusNotFound
		}
	}

	return nil, fasthttp.StatusInternalServerError
}

func (ph *PostHandler) Update(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var obj post2.PostUpdate
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	idString := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	result, err := ph.sb.post.ById(postId, nil)
	switch err {
	case nil:
		var err error = nil
		if obj.Message != nil {
			if *result.Post.Message != *obj.Message {
				err = ph.sb.post.UpdateById(postId, obj)
				result.Post.IsEdited = new(bool)
				*result.Post.IsEdited = true
				result.Post.Message = obj.Message
			}
		}

		if err == nil {
			return result.Post, fasthttp.StatusOK
		}
	case post2.ErrNotFound:
		return Error{
			Message: fmt.Sprintf("Can't find post with id: %d", postId),
		}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}
