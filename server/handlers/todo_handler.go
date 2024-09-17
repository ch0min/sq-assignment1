package handlers

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/oTuff/sq-ola1/models"
)

func GetAllTodos(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		todos := []models.Todo{}

		rows, err := db.Query("SELECT id, title, text, isCompleted, category, deadline FROM todo")
		if err != nil {
			log.Fatal(err)
			return c.Status(500).SendString("Failed to retrieve todos")
		}
		defer rows.Close()

		for rows.Next() {
			var todo models.Todo
			var category sql.NullString
			var deadline sql.NullTime

			err := rows.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done, &category, &deadline)
			if err != nil {
				return c.Status(500).SendString("Failed to scan todo")
			}

			if category.Valid {
				todo.Category = &category.String
			}
			if deadline.Valid {
				todo.Deadline = &deadline.Time
			}

			todos = append(todos, todo)
		}

		return c.JSON(todos)
	}
}

func CreateTodo(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		todo := new(models.Todo)
		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		var lastInsertId int
		query := `INSERT INTO todo (title, text, iscompleted, category, deadline)
				  VALUES ($1, $2, $3, $4, $5) RETURNING id`
		err := db.QueryRow(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline).Scan(&lastInsertId)
		if err != nil {
			return c.Status(500).SendString("Failed to create todo")
		}

		todo.ID = lastInsertId
		return c.Status(201).JSON(todo)
	}
}

func UpdateTodo(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		todo := new(models.Todo)
		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		query := `UPDATE todo SET title=$1, text=$2, iscompleted=$3, category=$4, deadline=$5 WHERE id=$6`
		_, err = db.Exec(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline, id)
		if err != nil {
			return c.Status(500).SendString("Failed to update todo")
		}

		return c.Status(200).JSON(todo)
	}
}

func ToggleTodoStatus(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}

		var currentStatus bool
		err = db.QueryRow("SELECT iscompleted FROM todo WHERE id=$1", id).Scan(&currentStatus)
		if err != nil {
			return c.Status(500).SendString("Failed to retrieve todo status")
		}

		newStatus := !currentStatus
		_, err = db.Exec("UPDATE todo SET iscompleted=$1 WHERE id=$2", newStatus, id)
		if err != nil {
			return c.Status(500).SendString("Failed to update task status")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func DeleteTodo(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		_, err = db.Exec("DELETE FROM todo WHERE id = $1", id)
		if err != nil {
			return c.Status(500).SendString("Failed to delete todo")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
