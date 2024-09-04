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
	Deadline    *time.Time `json:"deadline"`
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
	
			if deadline.Valid {
				todo.Deadline = &deadline.Time
			} else {
				todo.Deadline = nil
			}
	
			todos = append(todos, todo)
		}
	
		return c.JSON(todos)	
	})


	app.Post("api/todos", func(c *fiber.Ctx) error {
		todo := new(Todo)


		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		// Insert the todo into the database
		var lastInsertId int
		query := `INSERT INTO todo (title, text, isCompleted, category, deadline) 
				  VALUES ($1, $2, $3, $4, $5) RETURNING id`
		err := db.QueryRow(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline).Scan(&lastInsertId)
		if err != nil {
			return c.Status(500).SendString("Failed to create todo")
		}

		// Return the newly created todo
		todo.ID = lastInsertId
		return c.Status(201).JSON(todo)
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


	app.Patch("api/todos/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(401).SendString("Invalid id")
		}


		_, err = db.Exec("UPDATE todo SET isCompleted = TRUE WHERE id=$1", id)
		if err != nil {
			return c.Status(500).SendString("Failed to update done")
		}

	
		return c.SendStatus(fiber.StatusNoContent)

	})




	log.Fatal(app.Listen(":4000"))
}