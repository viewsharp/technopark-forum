package handlers

import (
	"github.com/valyala/fasthttp"
	user2 "github.com/viewsharp/technopark-forum/internal/resources/user"
	"strconv"
)

type UserHandler struct {
	sb *StorageBundle
}

func NewUserHandler(storageBundle *StorageBundle) *UserHandler {
	return &UserHandler{sb: storageBundle}
}

func (uh *UserHandler) Create(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var obj user2.User
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	nickname := ctx.UserValue("nickname").(string)
	obj.Nickname = &nickname

	err = uh.sb.user.Add(&obj)
	switch err {
	case nil:
		return obj, fasthttp.StatusCreated
	case user2.ErrUniqueViolation:
		var result user2.Users

		// TODO: Make union

		userByEmail, err := uh.sb.user.ByEmail(*obj.Email)
		if err == nil {
			result = append(result, userByEmail)
		}

		userByNickname, err := uh.sb.user.ByNickname(*obj.Nickname)
		if err == nil {
			if userByEmail == nil {
				result = append(result, userByNickname)
			} else if *userByNickname.Nickname != *userByEmail.Nickname {
				result = append(result, userByNickname)
			}
		}

		return result, fasthttp.StatusConflict
	default:
		return nil, fasthttp.StatusInternalServerError
	}

}

func (uh *UserHandler) Get(ctx *fasthttp.RequestCtx) (interface{}, int) {
	nickname := ctx.UserValue("nickname").(string)

	result, err := uh.sb.user.ByNickname(nickname)

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	case user2.ErrNotFound:
		return Error{
			Message: "Can't find user by nickname: " + nickname,
		}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}

func (uh *UserHandler) Update(ctx *fasthttp.RequestCtx) (interface{}, int) {
	nickname := ctx.UserValue("nickname").(string)

	var obj user2.UserUpdate
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	err = uh.sb.user.UpdateByNickname(nickname, &obj)

	switch err {
	case nil:
		return user2.User{
			About:    obj.About,
			Email:    obj.Email,
			FullName: obj.FullName,
			Nickname: &nickname,
		}, fasthttp.StatusOK
	case user2.ErrUniqueViolation:
		return Error{
			Message: "This email is already registered by user: " + *obj.Email,
		}, fasthttp.StatusConflict
	case user2.ErrNotFound:
		return Error{
			Message: "Can't find user by nickname: " + nickname,
		}, fasthttp.StatusNotFound
	}
	return nil, fasthttp.StatusInternalServerError
}

func (uh *UserHandler) GetByForum(ctx *fasthttp.RequestCtx) (interface{}, int) {
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
	case user2.ErrNotFoundForum:
		return Error{Message: "Can't find forum by slug: " + slug}, fasthttp.StatusNotFound
	}

	return nil, fasthttp.StatusInternalServerError
}
