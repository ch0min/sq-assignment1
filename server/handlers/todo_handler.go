package handlers

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/oTuff/sq-ola1/models"
)

// GetAllTodos godoc
// @Summary Get all todos
// @Description Get a list of all todos
// @Tags todos
// @Produce json
// @Success 200 {array} models.Todo
// @Router /api/todos [get]
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

// CreateTodo godoc
// @Summary Create a new todo
// @Description Create a new todo item and store it in the database
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body models.Todo true "Todo"
// @Success 201 {object} models.Todo
// @Router /api/todos [post]
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

// UpdateTodo godoc
// @Summary Update a todo item
// @Description Update the details of an existing todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body models.Todo true "Todo"
// @Success 200 {object} models.Todo
// @Router /api/todos/{id} [patch]
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

// ToggleTodoStatus godoc
// @Summary Toggle todo completion status
// @Description Toggle the "done" status of a specific todo item
// @Tags todos
// @Produce json
// @Param id path int true "Todo ID"
// @Success 204 "No Content"
// @Router /api/todos/{id}/done [patch]
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

// DeleteTodo godoc
// @Summary Delete a todo item
// @Description Delete a specific todo item by ID
// @Tags todos
// @Produce json
// @Param id path int true "Todo ID"
// @Success 204 "No Content"
// @Router /api/todos/{id} [delete]
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
