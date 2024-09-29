package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"

	post2 "github.com/viewsharp/technopark-forum/internal/usecase/post"
	"github.com/viewsharp/technopark-forum/internal/usecase/thread"
)

type PostHandler struct {
	sb *UsecaseSet
}

func NewPostHandler(storageBundle *UsecaseSet) *PostHandler {
	return &PostHandler{sb: storageBundle}
}

func (ph *PostHandler) Create(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var posts []post2.Post
	err := json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, threadIdParseErr := strconv.Atoi(slugOrId)

	if len(posts) == 0 {
		if threadIdParseErr == nil {
			_, err = ph.sb.thread.ById(ctx, threadId)
		} else {
			_, err = ph.sb.thread.BySlug(ctx, slugOrId)
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

		return Error{Message: err.Error()}, fasthttp.StatusInternalServerError
	}

	if threadIdParseErr == nil {
		err = ph.sb.post.AddByThreadId(ctx, posts, int32(threadId))
	} else {
		err = ph.sb.post.AddByThreadSlug(ctx, posts, slugOrId)
	}

	if err != nil {
		if errors.Is(err, post2.ErrInvalidParent) {
			return Error{Message: "Parent post was created in another thread"}, fasthttp.StatusConflict
		}
		if errors.Is(err, post2.ErrNotFoundThread) {
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

		var errNotFoundUser post2.ErrNotFoundUser
		if errors.As(err, &errNotFoundUser) {
			return Error{
				Message: "Can't find post author by nickname: " + errNotFoundUser.Nickname,
			}, fasthttp.StatusNotFound
		}

		return Error{Message: err.Error()}, fasthttp.StatusInternalServerError
	}

	return posts, fasthttp.StatusCreated
}

func (ph *PostHandler) Get(ctx *fasthttp.RequestCtx) (interface{}, int) {
	idString := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var result *post2.PostFull

	related := ctx.QueryArgs().Peek("related")

	if related == nil {
		result, err = ph.sb.post.ById(ctx, postId, nil)
	} else {
		result, err = ph.sb.post.ById(ctx, postId, strings.Split(string(related), ","))
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

func (ph *PostHandler) GetByThread(ctx *fasthttp.RequestCtx) (interface{}, int) {
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
	var posts []post2.Post
	switch string(ctx.QueryArgs().Peek("sort")) {
	case "tree":
		if threadIdParseErr == nil {
			posts, err = ph.sb.post.TreeByThreadId(ctx, threadId, limit, desc, since)
		} else {
			posts, err = ph.sb.post.TreeByThreadSlug(ctx, slugOrId, limit, desc, since)
		}
	case "parent_tree":
		if threadIdParseErr == nil {
			posts, err = ph.sb.post.ParentTreeByThreadId(ctx, threadId, limit, desc, since)
		} else {
			posts, err = ph.sb.post.ParentTreeByThreadSlug(ctx, slugOrId, limit, desc, since)
		}
	default:
		if threadIdParseErr == nil {
			posts, err = ph.sb.post.FlatByThreadId(ctx, threadId, limit, desc, since)
		} else {
			posts, err = ph.sb.post.FlatByThreadSlug(ctx, slugOrId, limit, desc, since)
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

func (ph *PostHandler) Update(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var obj post2.PostUpdate
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	idString := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	result, err := ph.sb.post.ById(ctx, postId, nil)
	switch err {
	case nil:
		var err error = nil
		if obj.Message != nil {
			if *result.Post.Message != *obj.Message {
				err = ph.sb.post.UpdateById(ctx, postId, obj)
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
