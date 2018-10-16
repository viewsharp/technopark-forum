package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/resources/thread"
	"github.com/viewsharp/TexPark_DBMSs/resources/vote"
	"strconv"
)

type VoteHandler struct {
	sb *StorageBundle
}

func NewVoteHandler(storageBundle *StorageBundle) *VoteHandler {
	return &VoteHandler{sb: storageBundle}
}

func (vh *VoteHandler) Create(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var obj vote.Vote
	err := obj.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		return nil, fasthttp.StatusBadRequest
	}

	var result *thread.Thread
	slugOrId := ctx.UserValue("slug_or_id").(string)
	threadId, err := strconv.Atoi(slugOrId)

	if err == nil {
		err = vh.sb.vote.AddByThreadId(&obj, threadId)
		if err != nil {
			return nil, fasthttp.StatusInternalServerError
		}

		result, err = vh.sb.thread.ById(threadId)
	} else {
		err = vh.sb.vote.AddByThreadSlug(&obj, slugOrId)
		if err != nil {
			return nil, fasthttp.StatusInternalServerError
		}

		result, err = vh.sb.thread.BySlug(slugOrId)
	}

	switch err {
	case nil:
		return result, fasthttp.StatusOK
	}

	return nil, fasthttp.StatusInternalServerError
}