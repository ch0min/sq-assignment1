package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupAppWithTestDB() (*fiber.App, *sql.DB, error) {
	// Use the same connection string as in `main.go`
	connStr := "host=localhost port=5432 user=postgres password=test dbname=todo sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, nil, err
	}

	// Setup Fiber app with the same routes
	app := fiber.New()
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

		lastInsertId, err := createTodo(db, todo)
		if err != nil {
			return c.Status(500).SendString("Failed to create todo")
		}

		todo.ID = lastInsertId
		return c.Status(201).JSON(todo)
	})

	return app, db, nil
}

func TestGetTodo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fixedTime := time.Date(2023, time.October, 10, 15, 30, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "title", "text", "isCompleted", "category", "deadline"}).
		AddRow(1, "Test Todo1", "This is a test todo", true, nil, nil).
		AddRow(2, "Test Todo2", "This is another test todo", false, "Work", fixedTime)

	mock.ExpectQuery("SELECT id, title, text, isCompleted, category, deadline FROM todo").WillReturnRows(rows)

	todo, err := getTodo(db, 1)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when calling getTodo", err)
	}

	expectedTodo := Todo{
		ID: 1, Title: "Test Todo1", Body: "This is a test todo", Done: true, Category: nil, Deadline: nil,
	}

	assert.Equal(t, expectedTodo, todo)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}

	todo, err = getTodo(db, 3)
	assert.EqualError(t, err, fmt.Sprintf("no todo found with id %d", 0))
}

func TestGetAllTodos(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fixedTime := time.Date(2023, time.October, 10, 15, 30, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "title", "text", "isCompleted", "category", "deadline"}).
		AddRow(1, "Test Todo1", "This is a test todo", true, nil, nil).
		AddRow(2, "Test Todo2", "This is another test todo", false, "Work", fixedTime)

	mock.ExpectQuery("SELECT id, title, text, isCompleted, category, deadline FROM todo").WillReturnRows(rows)

	todos, err := getAllTodos(db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when calling getAllTodos", err)
	}

	expectedTodos := []Todo{
		{ID: 1, Title: "Test Todo1", Body: "This is a test todo", Done: true, Category: nil, Deadline: nil},
		{ID: 2, Title: "Test Todo2", Body: "This is another test todo", Done: false, Category: func() *string { s := "Work"; return &s }(), Deadline: &fixedTime},
	}

	assert.Equal(t, expectedTodos, todos)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
func TestToggleTodoStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error '%s' when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT iscompleted FROM todo WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"iscompleted"}).AddRow(false))

	mock.ExpectExec("UPDATE todo SET iscompleted=\\$1 WHERE id=\\$2").
		WithArgs(true, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = toggleTodoStatus(db, 1)
	assert.NoError(t, err)

	mock.ExpectQuery("SELECT iscompleted FROM todo WHERE id=\\$1").
		WithArgs(2).
		WillReturnError(sql.ErrNoRows)

	err = toggleTodoStatus(db, 2)
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Integration Tests
func TestGetAllTodosIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppWithTestDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Seed test data into the test database
	_, err = db.Exec("INSERT INTO todo (title, text, iscompleted) VALUES ('Todo1', 'Body1', false), ('Todo2', 'Body2', true)")
	assert.NoError(t, err)

	// Create a test HTTP request to the /api/todos endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	resp, err := app.Test(req, -1) // -1 disables the timeout
	assert.NoError(t, err)

	// Check if the response status code is 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)

}
func TestCreateTodoIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppWithTestDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Define the new todo to be created
	newTodo := Todo{
		Title:    "New Integration Test Todo",
		Body:     "This is a test todo created in an integration test",
		Done:     false,
		Category: func() *string { s := "Testing"; return &s }(),
		Deadline: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}

	// Serialize the todo into JSON format
	newTodoJSON, err := json.Marshal(newTodo)
	assert.NoError(t, err)

	// Create a test HTTP request to the /api/todos endpoint
	req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(newTodoJSON))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, err := app.Test(req, -1) // -1 disables the timeout
	assert.NoError(t, err)

	// Print the response body and status code for debugging
	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Response Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(respBody))

	// Check if the response status code is 201 Created
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected a 201 Created status code")

	// Parse the response body into a Todo struct
	var createdTodo Todo
	err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&createdTodo)
	assert.NoError(t, err)

	// Verify that the response contains the correct title and body
	assert.Equal(t, newTodo.Title, createdTodo.Title)
	assert.Equal(t, newTodo.Body, createdTodo.Body)

	// Verify that the todo was inserted into the database
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM todo WHERE id=$1)", createdTodo.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "Todo should have been inserted into the database")
}

/* REMEMBER TO DO CLEAN UP NEXT TIME */
