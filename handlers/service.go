package handlers

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/resources/status"
)

type ServiceHandler struct {
	sb *StorageBundle
}

func NewServiceHandler(storageBundle *StorageBundle) *ServiceHandler {
	return &ServiceHandler{sb: storageBundle}
}

func (fh *ServiceHandler) Status(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	var result status.Status
	err := fh.sb.DB().QueryRow(`	SELECT (SELECT count(*) FROM forums),
											(SELECT count(*) FROM posts),
											(SELECT count(*) FROM threads),
											(SELECT count(*) FROM users);`,
	).Scan(&result.Forum, &result.Post, &result.Thread, &result.User)

	if err == nil {
		return result, fasthttp.StatusOK
	}

	return nil, fasthttp.StatusInternalServerError
}

func (fh *ServiceHandler) Clear(ctx *fasthttp.RequestCtx) (json.Marshaler, int) {
	//_, err := fh.sb.DB().Exec("TRUNCATE votes, posts, threads, forums, users")

	//if err == nil {
	//	return nil, fasthttp.StatusOK
	//}
	return nil, fasthttp.StatusInternalServerError
}
