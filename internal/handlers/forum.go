package handlers

import (
	"github.com/valyala/fasthttp"
	forum2 "github.com/viewsharp/technopark-forum/internal/resources/forum"
)

type ForumHandler struct {
	sb *StorageBundle
}

func NewForumHandler(storageBundle *StorageBundle) *ForumHandler {
	return &ForumHandler{sb: storageBundle}
}

func (fh *ForumHandler) Create(ctx *fasthttp.RequestCtx) (interface{}, int) {
	if string(ctx.Path()) != "/api/forum/create" {
		return nil, fasthttp.StatusNotFound
	}

	var obj forum2.Forum
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	err = fh.sb.forum.Add(&obj)
	switch err {
	case nil:
		return obj, fasthttp.StatusCreated
	case forum2.ErrUniqueViolation:
		result, err := fh.sb.forum.BySlug(*obj.Slug)
		if err == nil {
			return result, fasthttp.StatusConflict
		}
	case forum2.ErrNotFoundUser:
		return Error{Message: "Can't find user with nickname: " + *obj.User}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (fh *ForumHandler) Get(ctx *fasthttp.RequestCtx) (interface{}, int) {
	slug := ctx.UserValue("slug").(string)

	result, err := fh.sb.forum.FullBySlug(slug)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case forum2.ErrNotFound:
		return Error{Message: "Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}
