package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/transactions_api/internal/handlers"
)

func RoutesInitialization(app *fiber.App) {
	api := app.Group("/")

	api.Post("/api/send", handlers.Send)
	api.Get("/api/transactions", handlers.GetLast)
	api.Get("/api/wallet/:address/balance", handlers.GetBalance)
	app.Get("/swagger/*", swagger.HandlerDefault)
}
