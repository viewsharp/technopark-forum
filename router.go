package main

import (
	"encoding/json"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/TexPark_DBMSs/handlers"
)

type HandleFunc func(ctx *fasthttp.RequestCtx) (json.Marshaler, int)

func GetHandler(handleFunc HandleFunc) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		request, statusCode := handleFunc(ctx)

		if request == nil {
			ctx.SetStatusCode(statusCode)
			return
		}

		body, err := request.MarshalJSON()
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBody([]byte(err.Error()))
			return
		}

		ctx.SetStatusCode(statusCode)
		ctx.SetBody(body)
	}
}

type Router struct {
	*fasthttprouter.Router
}

func (r *Router)POST(path string, handle HandleFunc)  {
	r.Handle("POST", path, GetHandler(handle))
}

func (r *Router)GET(path string, handle HandleFunc)  {
	r.Handle("GET", path, GetHandler(handle))
}

func NewRouter(sb *handlers.StorageBundle) Router {
	router := Router{fasthttprouter.New()}

	forumHandler := handlers.NewForumHandler(sb)
	router.POST("/api/forum/:slug", forumHandler.Create) // "forum/create"
	router.GET("/api/forum/:slug/details", forumHandler.Get)

	threadHandler := handlers.NewThreadHandler(sb)
	router.POST("/api/forum/:slug/create", threadHandler.Create)
	router.GET("/api/forum/:slug/threads", threadHandler.GetByForum)
	router.GET("/api/thread/:slug_or_id/details", threadHandler.Get)
	router.POST("/api/thread/:slug_or_id/details", threadHandler.Update)

	userHandler := handlers.NewUserHandler(sb)
	router.GET("/api/user/:nickname/profile", userHandler.Get)
	router.POST("/api/user/:nickname/profile", userHandler.Update)
	router.POST("/api/user/:nickname/create", userHandler.Create)
	router.GET("/api/forum/:slug/users", userHandler.GetByForum)

	postHandler := handlers.NewPostHandler(sb)
	router.POST("/api/thread/:slug_or_id/create", postHandler.Create)
	router.GET("/api/thread/:slug_or_id/posts", postHandler.GetByThread)

	voteHandler := handlers.NewVoteHandler(sb)
	router.POST("/api/thread/:slug_or_id/vote", voteHandler.Create)

	serviceHandler := handlers.NewServiceHandler(sb)
	router.GET("/api/service/status", serviceHandler.Status)
	router.POST("/api/service/clear", serviceHandler.Clear)


	return router
}
