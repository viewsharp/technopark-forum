package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/viewsharp/technopark-forum/internal/resources/status"
)

type ServiceHandler struct {
	sb *StorageBundle
}

func NewServiceHandler(storageBundle *StorageBundle) *ServiceHandler {
	return &ServiceHandler{sb: storageBundle}
}

func (fh *ServiceHandler) Status(ctx *fasthttp.RequestCtx) (interface{}, int) {
	var result status.Status
	err := fh.sb.DB().QueryRow(ctx, `	SELECT (SELECT COUNT(*) FROM forums),
											(SELECT COUNT(*) FROM posts),
											(SELECT COUNT(*) FROM threads),
											(SELECT COUNT(*) FROM users);`,
	).Scan(&result.Forum, &result.Post, &result.Thread, &result.User)
	if err != nil {
		return Error{
			Message: err.Error(),
		}, fasthttp.StatusInternalServerError
	}
	return result, fasthttp.StatusOK
}

func (fh *ServiceHandler) Clear(ctx *fasthttp.RequestCtx) (interface{}, int) {
	_, err := fh.sb.DB().Exec(ctx, "TRUNCATE votes, posts, threads, forums, users, forum_user")
	if err != nil {
		return Error{
			Message: err.Error(),
		}, fasthttp.StatusInternalServerError
	}
	return nil, fasthttp.StatusOK
}
