package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/oTuff/sq-ola1/handlers"
)

func SetupTodoRoutes(app *fiber.App, db *sql.DB) {
	app.Get("/api/todos", handlers.GetAllTodos(db))
	app.Post("/api/todos", handlers.CreateTodo(db))
	app.Patch("/api/todos/:id", handlers.UpdateTodo(db))
	app.Patch("/api/todos/:id/done", handlers.ToggleTodoStatus(db))
	app.Delete("/api/todos/:id", handlers.DeleteTodo(db))
}

