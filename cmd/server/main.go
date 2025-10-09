package main

import (
	"context"
	"fmt"
	"log"
	"os"

	json "github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/controller"
	"github.com/viewsharp/technopark-forum/internal/db"
)

var ServerAddr = os.Getenv("SERVER_ADDR")
var PostgresDSN = os.Getenv("POSTGRES_DSN")

func main() {
	dbpool, err := pgxpool.New(context.Background(), PostgresDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	querier := db.New(dbpool)
	usecaseSet := controller.NewUsecaseSet(dbpool, querier)
	server := controller.NewServer(usecaseSet)

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Add logger middleware
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${latency} ${method} ${path}\n",
	}))

	// Register API routes with /api prefix
	apiGroup := app.Group("/api")
	api.RegisterHandlers(apiGroup, server)

	log.Printf("starting server at: %s\n", ServerAddr)
	log.Fatal(app.Listen(ServerAddr))
}
