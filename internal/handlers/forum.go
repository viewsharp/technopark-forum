package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/goccy/go-json"

	forumUC "github.com/viewsharp/technopark-forum/internal/usecase/forum"
)

type ForumHandler struct {
	sb *UsecaseSet
}

func NewForumHandler(storageBundle *UsecaseSet) *ForumHandler {
	return &ForumHandler{sb: storageBundle}
}

func (fh *ForumHandler) Create(ctx *fasthttp.RequestCtx) (interface{}, int) {
	if string(ctx.Path()) != "/api/forum/create" {
		return nil, fasthttp.StatusNotFound
	}

	var forum forumUC.Forum
	err := json.Unmarshal(ctx.PostBody(), &forum)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	createdForum, err := fh.sb.forum.Add(ctx, forum)
	if err != nil {
		switch err {
		case forumUC.ErrUniqueViolation:
			result, err := fh.sb.forum.BySlug(ctx, *forum.Slug)
			if err == nil {
				return result, fasthttp.StatusConflict
			}
		case forumUC.ErrNotFoundUser:
			return Error{Message: "Can't find user with nickname: " + *forum.User}, fasthttp.StatusNotFound
		}
		return Error{Message: err.Error()}, fasthttp.StatusInternalServerError
	}

	return createdForum, fasthttp.StatusCreated
}

func (fh *ForumHandler) Get(ctx *fasthttp.RequestCtx) (interface{}, int) {
	slug := ctx.UserValue("slug").(string)

	result, err := fh.sb.forum.FullBySlug(ctx, slug)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case forumUC.ErrNotFound:
		return Error{Message: "Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return Error{Message: err.Error()}, fasthttp.StatusInternalServerError
}
