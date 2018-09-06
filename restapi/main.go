package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
)

var DSN = "host=127.0.0.1 port=5433 user=docker password=docker dbname=docker"

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

	router := NewRouter(&Handler{
		DB: db,
	})

	fmt.Printf("starting server at: %s\n", port)
	log.Fatal(fasthttp.ListenAndServe(":"+port, func(ctx *fasthttp.RequestCtx) {
		router.Handler(ctx)
		fmt.Printf("%d [%s] %s\n", ctx.Response.Header.StatusCode(), ctx.Method(), ctx.Path())
	}))
}
