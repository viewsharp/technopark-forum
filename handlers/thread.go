package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/resources/thread"
	"github.com/viewsharp/TexPark_DBMSs/resources/user"
	"strconv"
)

type ThreadHandler struct {
	sb *StorageBundle
}

func NewThreadHandler(storageBundle *StorageBundle) *ThreadHandler {
	return &ThreadHandler{sb: storageBundle}
}

func (th *ThreadHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	slug:=ctx.UserValue("slug").(string)

	var obj thread.Thread
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}
	obj.Forum = &slug

	err = th.sb.thread.Add(&obj)
	switch err {
	case nil:
		return obj, fasthttp.StatusCreated
	case thread.ErrUniqueViolation:
		result, err := th.sb.thread.BySlug(*obj.Slug)
		if err == nil {
			return result, fasthttp.StatusConflict
		}
	case thread.ErrNotFoundUser:
		return Error{Message:"Can't find thread author by nickname: " + *obj.Author}, fasthttp.StatusNotFound
	case thread.ErrNotFoundForum:
		return Error{Message:"Can't find thread forum by slug: " + *obj.Forum}, fasthttp.StatusNotFound

	}

	return nil, fasthttp.StatusInternalServerError
}

func (th *ThreadHandler) GetByForum(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
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
	case thread.ErrNotFoundForum:
		return Error{Message:"Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (th *ThreadHandler) Get(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var result *thread.Thread
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, err := strconv.Atoi(slugOrId)
	if err == nil {
		result, err = th.sb.thread.ById(threadId)
	} else {
		result, err = th.sb.thread.BySlug(slugOrId)
	}
	if err == nil {
		return result, fasthttp.StatusOK
	}

	return nil, fasthttp.StatusInternalServerError
}

func (th *ThreadHandler) Update(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var obj thread.ThreadUpdate
	err := obj.UnmarshalJSON(ctx.PostBody())
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

	switch err {
	case nil:
		var result *thread.Thread
		if threadIdErr == nil {
			result, err = th.sb.thread.ById(threadId)
		} else {
			result, err = th.sb.thread.BySlug(slugOrId)
		}

		if err == nil {
			return result, fasthttp.StatusOK
		}
	case user.ErrUniqueViolation:
	case user.ErrNotFound:
	}

	return nil, fasthttp.StatusInternalServerError
}