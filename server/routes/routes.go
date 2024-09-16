package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oTuff/sq-ola1/handlers"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/todos", handlers.GetAllTodos)
	api.Post("/todos", handlers.CreateTodoHandler)
	api.Patch("/todos/:id", handlers.UpdateTodoHandler)
	api.Patch("/todos/:id/done", handlers.ToggleTodoStatusHandler)
	api.Delete("/todos/:id", handlers.DeleteTodoHandler)
}
