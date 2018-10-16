package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/resources/user"
	"strconv"
)

type UserHandler struct {
	sb *StorageBundle
}

func NewUserHandler(storageBundle *StorageBundle) *UserHandler {
	return &UserHandler{sb: storageBundle}
}

func (uh *UserHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var obj user.User
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	nickname := ctx.UserValue("nickname").(string)
	obj.Nickname = &nickname

	err = uh.sb.user.Add(&obj)
	switch err {
	case nil:
		return obj, fasthttp.StatusCreated
	case user.ErrUniqueViolation:
		var result user.Users
		userByEmail, _ := uh.sb.user.ByEmail(*obj.Email)
		// TODO: handle error
		result = append(result, userByEmail)

		userByUsername, _ := uh.sb.user.ByNickname(*obj.Nickname)
		// TODO: handle error
		result = append(result, userByUsername)

		return result, fasthttp.StatusConflict
	default:
		return nil, fasthttp.StatusInternalServerError
	}

}

func (uh *UserHandler) Get(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	nickname := ctx.UserValue("nickname").(string)

	result, err := uh.sb.user.ByNickname(nickname)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case user.ErrNotFound:
		return Error{
			Message: "Can't find user by nickname: " + nickname,
		}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (uh *UserHandler) Update(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	nickname := ctx.UserValue("nickname").(string)

	var obj user.UserUpdate
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	err = uh.sb.user.UpdateByNickname(nickname, &obj)

	switch err {
	case nil:
		return user.User{
			About:    obj.About,
			Email:    obj.Email,
			FullName: obj.FullName,
			Nickname: &nickname,
		}, fasthttp.StatusOK
	case user.ErrUniqueViolation:
		return Error{
			Message: "This email is already registered by user: " + *obj.Email,
		}, fasthttp.StatusConflict
	case user.ErrNotFound:
		return Error{
			Message: "Can't find user by nickname: " + nickname,
		}, fasthttp.StatusNotFound
	}
	return nil, fasthttp.StatusInternalServerError
}

func (uh *UserHandler) GetByForum(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
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

	result, err := uh.sb.user.ByForumSlug(slug, desc, since, limit)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case user.ErrNotFoundForum:
		return Error{Message: "Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}
