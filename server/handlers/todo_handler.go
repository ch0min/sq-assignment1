package handlers

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/oTuff/sq-ola1/db"
	"github.com/oTuff/sq-ola1/models"
	"log"
)

func GetAllTodos(c *fiber.Ctx) error {
	database := c.Locals("db").(*sql.DB) // Access the database from context
	todos, err := db.GetAllTodos(database)
	if err != nil {
		log.Fatal(err)
		return c.Status(500).SendString("Failed to retrieve todos")
	}
	return c.JSON(todos)
}

func CreateTodoHandler(c *fiber.Ctx) error {
	database := c.Locals("db").(*sql.DB)
	todo := new(models.Todo)
	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	lastInsertId, err := db.CreateTodo(database, todo)
	if err != nil {
		return c.Status(500).SendString("Failed to create todo")
	}
	todo.ID = lastInsertId
	return c.Status(201).JSON(todo)
}

func UpdateTodoHandler(c *fiber.Ctx) error {
	database := c.Locals("db").(*sql.DB)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	todo := new(models.Todo)
	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).SendString("Invalid request body")
	}

	err = db.UpdateTodo(database, id, todo)
	if err != nil {
		return c.Status(500).SendString("Failed to update task")
	}

	return c.Status(200).JSON(todo)
}

func ToggleTodoStatusHandler(c *fiber.Ctx) error {
	database := c.Locals("db").(*sql.DB)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	err = db.ToggleTodoStatus(database, id)
	if err != nil {
		return c.Status(500).SendString("Failed to update task status")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func DeleteTodoHandler(c *fiber.Ctx) error {
	database := c.Locals("db").(*sql.DB)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid todo ID")
	}

	err = db.DeleteTodo(database, id)
	if err != nil {
		return c.Status(500).SendString("Failed to delete todo")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
