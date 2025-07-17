package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	_ "github.com/transactions_api/docs"
	"github.com/transactions_api/internal/config"
	"github.com/transactions_api/internal/storage/postgres"
	"github.com/transactions_api/routes"
)

func main() {
	cfg := config.MustLoad()

	app := fiber.New(fiber.Config{
		Prefork: true,
	})

	db := postgres.ConnectDB(cfg)
	postgres.FillDatabase()
	defer db.Close(context.Background())

	routes.RoutesInitialization(app)
	log.Fatal(app.Listen(cfg.Server.Port))
}
