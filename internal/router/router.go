package router

import (
	"github.com/buaazp/fasthttprouter"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	"github.com/viewsharp/technopark-forum/internal/handlers"
)

var json = jsoniter.ConfigFastest

type HandleFunc func(ctx *fasthttp.RequestCtx) (interface{}, int)

func GetHandler(handleFunc HandleFunc) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		responseData, statusCode := handleFunc(ctx)

		if responseData == nil {
			ctx.SetStatusCode(statusCode)
			return
		}

		body, err := json.Marshal(&responseData)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBody([]byte(err.Error()))
			return
		}

		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(statusCode)
		ctx.SetBody(body)
	}
}

type Router struct {
	*fasthttprouter.Router
}

func (r *Router) POST(path string, handle HandleFunc) {
	r.Handle("POST", path, GetHandler(handle))
}

func (r *Router) GET(path string, handle HandleFunc) {
	r.Handle("GET", path, GetHandler(handle))
}

func New(sb *handlers.StorageBundle) Router {
	router := Router{fasthttprouter.New()}

	router.Handle("GET", "/api", func(ctx *fasthttp.RequestCtx) {
		ctx.SetBody([]byte("[]"))
	})

	forumHandler := handlers.NewForumHandler(sb)
	router.POST("/api/forum/:slug", forumHandler.Create) // "/api/forum/create"
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
	router.GET("/api/post/:id/details", postHandler.Get)
	router.POST("/api/post/:id/details", postHandler.Update)

	voteHandler := handlers.NewVoteHandler(sb)
	router.POST("/api/thread/:slug_or_id/vote", voteHandler.Create)

	serviceHandler := handlers.NewServiceHandler(sb)
	router.GET("/api/service/status", serviceHandler.Status)
	router.POST("/api/service/clear", serviceHandler.Clear)

	return router
}
