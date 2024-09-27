package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/viewsharp/technopark-forum/internal/handlers"
	"github.com/viewsharp/technopark-forum/internal/qlogger"
	"github.com/viewsharp/technopark-forum/internal/router"
)

var ServerAddr = os.Getenv("SERVER_ADDR")
var PostgresDSN = os.Getenv("POSTGRES_DSN")

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	db, err := sql.Open("postgres", PostgresDSN)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		log.Fatal(err)
	}

	//storageBundle := handlers.NewStorageBundle(db)
	storageBundle := handlers.NewStorageBundle(qlogger.NewQueryLogger(db))
	serverRouter := router.New(storageBundle)

	log.Printf("starting server at: %s\n", ServerAddr)
	log.Fatal(fasthttp.ListenAndServe(ServerAddr, func(ctx *fasthttp.RequestCtx) {
		t := time.Now()
		serverRouter.Handler(ctx)
		logger.Info(
			"handled",
			zap.Int("status", ctx.Response.Header.StatusCode()),
			zap.ByteString("method", ctx.Method()),
			zap.Duration("duration", time.Since(t)),
			zap.ByteString("uri", ctx.Request.Header.RequestURI()),
		)
	}))
}
