package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	json "github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/viewsharp/technopark-forum/internal/api"
	"github.com/viewsharp/technopark-forum/internal/controller"
	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/repository"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	serverAddr := os.Getenv("SERVER_ADDR")
	postgresDSN := os.Getenv("POSTGRES_DSN")

	dbpool, err := pgxpool.New(context.Background(), postgresDSN)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %s", err)
	}
	defer dbpool.Close()

	querier := db.New(dbpool)
	repositorySet := repository.NewRepositorySet(dbpool, querier)
	server := controller.NewServer(repositorySet)

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Add logger middleware
	//app.Use(logger.New(logger.Config{
	//	Format: "${time} ${status} - ${latency} ${method} ${path}\n",
	//}))

	// Register API routes with /api prefix
	apiGroup := app.Group("/api")
	api.RegisterHandlers(apiGroup, server)

	log.Printf("starting server at: %s\n", serverAddr)
	log.Fatal(app.Listen(serverAddr))
}
