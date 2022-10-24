package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/technopark-forum/internal/resources/thread"
	vote2 "github.com/viewsharp/technopark-forum/internal/resources/vote"
	"strconv"
)

type VoteHandler struct {
	sb *StorageBundle
}

func NewVoteHandler(storageBundle *StorageBundle) *VoteHandler {
	return &VoteHandler{sb: storageBundle}
}

func (vh *VoteHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var obj vote2.Vote
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var result *thread.Thread
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, err := strconv.Atoi(slugOrId)

	if err == nil {
		err = vh.sb.vote.AddByThreadId(&obj, threadId)
		switch err {
		case nil:
			result, err = vh.sb.thread.ById(threadId)
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
		err = vh.sb.vote.AddByThreadSlug(&obj, slugOrId)
		switch err {
		case nil:
			result, err = vh.sb.thread.BySlug(slugOrId)
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
