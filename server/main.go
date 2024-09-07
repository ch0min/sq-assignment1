package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"log"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Done  bool   `json:"done"`
	Category    *string `json:"category"`
	Deadline    *string `json:"deadline"`
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

func main() {
		// Setup db connection
		connStr := "host=localhost port=5432 user=postgres password=test dbname=todo sslmode=disable"

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Successfully connected to the database!")

		// App
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// todos := []Todo{}

	// app.Get("/healthcheck", func(c *fiber.Ctx) error {
	// 	return c.SendString("OK")
	// })




	app.Get("/api/todos", func(c *fiber.Ctx) error {
		if err != nil {
			return c.Status(500).SendString("Failed to retrieve todos")
		}
	
		todos := []Todo{}
	
		rows, err := db.Query("SELECT id, title, text, isCompleted, category, deadline FROM todo")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
	
		for rows.Next() {
			var todo Todo
			var category sql.NullString
			var deadline sql.NullTime
	
			err := rows.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done, &category, &deadline)
			if err != nil {
				log.Fatal(err)
			}
	
			// Handle nullable fields (category and deadline)
			if category.Valid {
				todo.Category = &category.String
			} else {
				todo.Category = nil
			}
	
			// Convert deadline to a string in the format "YYYY-MM-DD"
			if deadline.Valid {
				formattedDeadline := deadline.Time.Format("2006-01-02")
				todo.Deadline = &formattedDeadline
			} else {
				todo.Deadline = nil
			}
	
			todos = append(todos, todo)
		}
	
		return c.JSON(todos)
	})


	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := new(Todo)
	
		// Parse the request body into the todo struct
		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}
	
		// Convert the deadline string to time.Time, if provided
		var deadline *time.Time
		if todo.Deadline != nil && *todo.Deadline != "" {
			// Convert the deadline from string to time.Time
			parsedDeadline, err := time.Parse("2006-01-02", *todo.Deadline)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid deadline format, expected YYYY-MM-DD")
			}
			deadline = &parsedDeadline
		}
	
		// Insert the todo into the database
		var lastInsertId int
		query := `INSERT INTO todo (title, text, iscompleted, category, deadline) 
				  VALUES ($1, $2, $3, $4, $5) RETURNING id`
		err := db.QueryRow(query, todo.Title, todo.Body, todo.Done, todo.Category, deadline).Scan(&lastInsertId)
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
	
		// Convert the deadline string to time.Time, if provided
		var deadline *time.Time
		if todo.Deadline != nil && *todo.Deadline != "" {
			// Convert the deadline from string to time.Time
			parsedDeadline, err := time.Parse("2006-01-02", *todo.Deadline)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid deadline format, expected YYYY-MM-DD")
			}
			deadline = &parsedDeadline
		}
	
		query := `UPDATE todo SET title=$1, text=$2, iscompleted=$3, category=$4, deadline=$5 WHERE id=$6`
		_, err = db.Exec(query, todo.Title, todo.Body, todo.Done, todo.Category, deadline, id)
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
	
		// Retrieve the current status of the task
		var currentStatus bool
		err = db.QueryRow("SELECT iscompleted FROM todo WHERE id=$1", id).Scan(&currentStatus)
		if err != nil {
			return c.Status(500).SendString("Failed to retrieve task")
		}
	
		// Toggle the status
		newStatus := !currentStatus
	
		// Update the status in the database
		_, err = db.Exec("UPDATE todo SET iscompleted=$1 WHERE id=$2", newStatus, id)
		if err != nil {
			return c.Status(500).SendString("Failed to update task status")
		}
	
		return c.SendStatus(fiber.StatusNoContent)
	})





	app.Delete("api/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid todo ID")
		}

		_, err = db.Exec("Delete FROM todo WHERE id = $1", id)
		if err != nil {
			return c.Status(500).SendString("Failed to delete todo")
		}

		return c.SendStatus(fiber.StatusNoContent)
	})







	log.Fatal(app.Listen(":4000"))
}