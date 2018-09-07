package main

import (
	"encoding/json"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
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

func NewRouter(handler *Handler) Router {
	router := Router{fasthttprouter.New()}

	router.POST("/forum/:slug", handler.CreateForum) // "forum/create"
	router.POST("/forum/:slug/create", handler.CreateThread)
	router.GET("/forum/:slug/details", handler.GetForum)
	router.GET("/forum/:slug/threads", handler.GetForumThreads)
	router.GET("/forum/:slug/users", handler.GetForumUsers)
	router.GET("/post/:id/details", handler.GetPost)
	router.POST("/post/:id/details", handler.UpdatePost)
	router.POST("/service/clear", handler.ClearDB)
	router.GET("/service/status", handler.GetDBStatus)
	router.POST("/thread/:slug_or_id/create", handler.CreatePost)
	router.GET("/thread/:slug_or_id/details", handler.GetThread)
	router.POST("/thread/:slug_or_id/details", handler.UpdateThread)
	router.GET("/thread/:slug_or_id/posts", handler.GetThreadPosts)
	router.POST("/thread/:slug_or_id/vote", handler.CreateThreadVote)
	router.POST("/user/:nickname/create", handler.CreateUser)
	router.GET("/user/:nickname/profile", handler.GetUser)
	router.POST("/user/:nickname/profile", handler.UpdateUser)

	return router
}
