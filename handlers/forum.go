package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/resources/forum"
)

type ForumHandler struct {
	sb *StorageBundle
}

func NewForumHandler(storageBundle *StorageBundle) *ForumHandler {
	return &ForumHandler{sb: storageBundle}
}

func (fh *ForumHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	if string(ctx.Path()) != "/forum/create" {
		return nil, fasthttp.StatusNotFound
	}

	var obj forum.Forum
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	err = fh.sb.forum.Add(&obj)
	switch err {
	case nil:
		return obj, fasthttp.StatusCreated
	case forum.ErrUniqueViolation:
		result, err := fh.sb.forum.BySlug(obj.Slug)
		if err == nil {
			return result, fasthttp.StatusConflict
		}
	case forum.ErrNotFoundUser:
		return Error{Message: "Can't find user with nickname: " + obj.User}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (fh *ForumHandler) Get(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	slug := ctx.UserValue("slug").(string)

	result, err := fh.sb.forum.FullBySlug(slug)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case forum.ErrNotFound:
		return Error{Message: "Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}
