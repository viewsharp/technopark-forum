package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/handlers"
	"github.com/viewsharp/technopark-forum/internal/router"
)

var ServerAddr = os.Getenv("SERVER_ADDR")
var PostgresDSN = os.Getenv("POSTGRES_DSN")

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dbpool, err := pgxpool.New(context.Background(), PostgresDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	querier := db.New(dbpool)

	usecaseSet := handlers.NewUsecaseSet(dbpool, querier)
	serverRouter := router.New(usecaseSet)

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
