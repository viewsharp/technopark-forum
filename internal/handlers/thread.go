package handlers

import (
	"fmt"
	"github.com/valyala/fasthttp"
	thread2 "github.com/viewsharp/technopark-forum/internal/resources/thread"
	"strconv"
)

type ThreadHandler struct {
	sb *StorageBundle
}

func NewThreadHandler(storageBundle *StorageBundle) *ThreadHandler {
	return &ThreadHandler{sb: storageBundle}
}

func (th *ThreadHandler) Create(ctx *fasthttp.RequestCtx) (interface{}, int) {
	slug := ctx.UserValue("slug").(string)

	var obj thread2.Thread
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}
	obj.Forum = &slug

	err = th.sb.thread.Add(&obj)
	switch err {
	case nil:
		return obj, fasthttp.StatusCreated
	case thread2.ErrUniqueViolation:
		result, err := th.sb.thread.BySlug(*obj.Slug)
		if err == nil {
			return result, fasthttp.StatusConflict
		}
	case thread2.ErrNotFoundUser:
		return Error{Message: "Can't find thread author by nickname: " + *obj.Author}, fasthttp.StatusNotFound
	case thread2.ErrNotFoundForum:
		return Error{Message: "Can't find thread forum by slug: " + *obj.Forum}, fasthttp.StatusNotFound

	}

	return nil, fasthttp.StatusInternalServerError
}

func (th *ThreadHandler) GetByForum(ctx *fasthttp.RequestCtx) (interface{}, int) {
	slug := ctx.UserValue("slug").(string)

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

	since := string(ctx.QueryArgs().Peek("since"))

	result, err := th.sb.thread.ByForumSlug(slug, desc, since, limit)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case thread2.ErrNotFoundForum:
		return Error{Message: "Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (th *ThreadHandler) Get(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var result *thread2.Thread
	var err error
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, threadIdParseErr := strconv.Atoi(slugOrId)
	if threadIdParseErr == nil {
		result, err = th.sb.thread.ById(threadId)
	} else {
		result, err = th.sb.thread.BySlug(slugOrId)
	}

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case thread2.ErrNotFound:
		if threadIdParseErr == nil {
			return Error{
				Message: fmt.Sprintf("Can't find thread by id: %d", threadId),
			}, fasthttp.StatusNotFound
		} else {
			return Error{
				Message: "Can't find thread by slug: " + slugOrId,
			}, fasthttp.StatusNotFound
		}
	}

	return nil, fasthttp.StatusInternalServerError
}

func (th *ThreadHandler) Update(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var obj thread2.ThreadUpdate
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, threadIdErr := strconv.Atoi(slugOrId)
	if threadIdErr == nil {
		err = th.sb.thread.UpdateById(threadId, &obj)
	} else {
		err = th.sb.thread.UpdateBySlug(slugOrId, &obj)
	}

	if err == nil {
		var result *thread2.Thread
		if threadIdErr == nil {
			result, err = th.sb.thread.ById(threadId)
		} else {
			result, err = th.sb.thread.BySlug(slugOrId)
		}

		switch err {
		case nil:
			return result, fasthttp.StatusOK
		case thread2.ErrNotFound:
			if threadIdErr == nil {
				return Error{
					Message: fmt.Sprintf("Can't find thread by id: %d", threadId),
				}, fasthttp.StatusNotFound
			} else {
				return Error{
					Message: "Can't find thread by slug: " + slugOrId,
				}, fasthttp.StatusNotFound
			}
		}
	}
	return nil, fasthttp.StatusInternalServerError
}
