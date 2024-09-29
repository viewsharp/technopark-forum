package handlers

import (
	"fmt"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/viewsharp/technopark-forum/internal/usecase/thread"
	vote2 "github.com/viewsharp/technopark-forum/internal/usecase/vote"
)

type VoteHandler struct {
	sb *UsecaseSet
}

func NewVoteHandler(storageBundle *UsecaseSet) *VoteHandler {
	return &VoteHandler{sb: storageBundle}
}

func (vh *VoteHandler) Create(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var obj vote2.Vote
	err := json.Unmarshal(ctx.PostBody(), &obj)
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var result *thread.Thread
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, err := strconv.Atoi(slugOrId)

	if err == nil {
		err = vh.sb.vote.AddByThreadId(ctx, &obj, threadId)
		switch err {
		case nil:
			result, err = vh.sb.thread.ById(ctx, threadId)
		case vote2.ErrNotFoundThread:
			return Error{
				Message: fmt.Sprintf("Can't find thread by id: %d", threadId),
			}, fasthttp.StatusNotFound
		case vote2.ErrNotFoundUser:
			return Error{
				Message: "Can't find user by nickname: " + *obj.Nickname,
			}, fasthttp.StatusNotFound

		default:
			return nil, fasthttp.StatusInternalServerError
		}
	} else {
		err = vh.sb.vote.AddByThreadSlug(ctx, &obj, slugOrId)
		switch err {
		case nil:
			result, err = vh.sb.thread.BySlug(ctx, slugOrId)
		case vote2.ErrNotFoundThread:
			return Error{
				Message: "Can't find user by nickname: " + *obj.Nickname,
			}, fasthttp.StatusNotFound
		case vote2.ErrNotFoundUser:
			return Error{
				Message: fmt.Sprintf("Can't find thread by id: %d", threadId),
			}, fasthttp.StatusNotFound

		default:
			return nil, fasthttp.StatusInternalServerError
		}
	}

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	}

	return nil, fasthttp.StatusInternalServerError
}
