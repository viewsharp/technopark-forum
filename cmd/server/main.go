package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"github.com/viewsharp/technopark-forum/internal/handlers"
	"github.com/viewsharp/technopark-forum/internal/router"
)

var DSN = "host=127.0.0.1 port=5432 user=postgres password=postgres dbname=tp-forum sslmode=disable"

func main() {
	if len(os.Args) != 2 {
		fmt.Println("The program accepts one argument <port>")
		return
	}
	port := os.Args[1]

	db, err := sql.Open("postgres", DSN)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	serverRouter := router.New(handlers.NewStorageBundle(db))

	fmt.Printf("starting server at: %s\n", port)
	//fasthttp.ListenAndServe(":"+port, serverRouter.Handler)
	log.Fatal(fasthttp.ListenAndServe(":"+port, func(ctx *fasthttp.RequestCtx) {
		t := time.Now()
		serverRouter.Handler(ctx)
		fmt.Printf("%d\t[%s]\t{%v}\t%s\n", ctx.Response.Header.StatusCode(), ctx.Method(), time.Since(t), ctx.URI())
	}))
}
