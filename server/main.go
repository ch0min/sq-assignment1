package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID       int        `json:"id"`
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Done     bool       `json:"done"`
	Category *string    `json:"category"`
	Deadline *time.Time `json:"deadline"`
}

// Validating Business Logic
func validateTodoInput(todo *Todo) error {
    if todo.Title == "" {
        return errors.New("task title must not be empty")
    }
    if len(todo.Body) < 10 {
        return errors.New("task description must have at least 10 characters")
    }
    return nil
}

func getTodo(db *sql.DB, id int) (Todo, error) {
	todo := Todo{}

	row := db.QueryRow("SELECT id, title, text, isCompleted, category, deadline FROM todo WHERE id = $1", id)

	var category sql.NullString
	var deadline sql.NullTime

	err := row.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done, &category, &deadline)
	if err != nil {
		// If no row is found, handle the error
		if err == sql.ErrNoRows {
			return todo, fmt.Errorf("no todo found with id %d", id)
		}
		return todo, err
	}

	// Handle nullable fields
	if category.Valid {
		todo.Category = &category.String
	} else {
		todo.Category = nil
	}

	if deadline.Valid {
		todo.Deadline = &deadline.Time
	} else {
		todo.Deadline = nil
	}

	return todo, nil
}

func getAllTodos(db *sql.DB) ([]Todo, error) {
	todos := []Todo{}

	rows, err := db.Query("SELECT id, title, text, isCompleted, category, deadline FROM todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		var category sql.NullString
		var deadline sql.NullTime

		err := rows.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done, &category, &deadline)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields (category and deadline)
		if category.Valid {
			todo.Category = &category.String
		} else {
			todo.Category = nil
		}

		if deadline.Valid {
			todo.Deadline = &deadline.Time
		} else {
			todo.Deadline = nil
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func createTodo(db *sql.DB, todo *Todo) (int, error) {
	var lastInsertId int
	query := `INSERT INTO todo (title, text, iscompleted, category, deadline)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline).Scan(&lastInsertId)
	return lastInsertId, err
}

func updateTodo(db *sql.DB, id int, todo *Todo) error {
	query := `UPDATE todo SET title=$1, text=$2, iscompleted=$3, category=$4, deadline=$5 WHERE id=$6`
	_, err := db.Exec(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline, id)
	return err
}

func toggleTodoStatus(db *sql.DB, id int) error {
	// Retrieve the current status of the task
	var currentStatus bool
	err := db.QueryRow("SELECT iscompleted FROM todo WHERE id=$1", id).Scan(&currentStatus)
	if err != nil {
		return err
	}

	// Toggle the status
	newStatus := !currentStatus

	// Update the status in the database
	_, err = db.Exec("UPDATE todo SET iscompleted=$1 WHERE id=$2", newStatus, id)
	return err
}

func deleteTodo(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM todo WHERE id=$1", id)
	return err
}

func setupAppAndDB() (*fiber.App, *sql.DB, error) {
	connStr := "host=localhost port=5432 user=postgres password=test dbname=todo sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, nil, err
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/api/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}


		// todo := Todo{}
		todo, err := getTodo(db, id) // Fixed by staticcheck - alternative for PMD
		if err != nil {
			return c.Status(400).SendString("no todo with that id")
		}

		return c.Status(200).JSON(todo)
	})

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		todos, err := getAllTodos(db)
		if err != nil {
			log.Fatal(err)
			return c.Status(500).SendString("Failed to retrieve todos")
		}

		return c.JSON(todos)
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := new(Todo)

		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		// Validate the todo before inserting it into the database
		if err := validateTodoInput(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}


		// Insert the todo into the database
		lastInsertId, err := createTodo(db, todo)
		if err != nil {
			return c.Status(500).SendString("Failed to create todo")
		}

		// Return the newly created todo
		todo.ID = lastInsertId
		return c.Status(201).JSON(todo)
	})

	app.Patch("/api/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}


		todo := new(Todo)
		if err := c.BodyParser(todo); err != nil {
			return c.Status(400).SendString("Invalid request body")
		}

			// Validate the todo before inserting it into the database
			if err := validateTodoInput(todo); err != nil {
				return c.Status(fiber.StatusBadRequest).SendString(err.Error())
			}

		err = updateTodo(db, id, todo)
		if err != nil {
			return c.Status(500).SendString("Failed to update task")
		}

		return c.Status(200).JSON(todo)
	})

	app.Patch("/api/todos/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}

		err = toggleTodoStatus(db, id)
		if err != nil {
			return c.Status(500).SendString("Failed to update task status")
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Delete("/api/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid todo ID")
		}

		err = deleteTodo(db, id)
		if err != nil {
			return c.Status(500).SendString("Failed to delete todo")
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	return app, db, nil
}

func main() {
	app, db, err := setupAppAndDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Fatal(app.Listen("localhost:4000"))
}
