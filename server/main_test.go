package main

import (
	"bytes"
	"database/sql"
	"encoding/json"

	// "fmt"
	"io"
	"log"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAppStartup(t *testing.T) {
	app, db, err := setupAppAndDB()
	assert.NoError(t, err)
	defer db.Close()

	// Perform a simple request to see if the app starts correctly
	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected the app to start and respond with 200")
}

func TestSetupAppAndDB_Success(t *testing.T) {
	// Call setupAppAndDB
	app, db, err := setupAppAndDB()

	// Ensure no errors are returned
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, db)

	// Cleanup: Close the database connection if necessary
	if db != nil {
		db.Close()
	}
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

	// _, err = getTodo(db, 3)
	// assert.EqualError(t, err, fmt.Sprintf("no todo found with id %d", 0))
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

func TestUpdateTodo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Define the todo item to be updated
	todo := Todo{
		ID:       1,
		Title:    "Updated Title",
		Body:     "Updated Body",
		Done:     true,
		Category: func() *string { s := "Updated Category"; return &s }(),
		Deadline: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
	}

	// Expect the update query to be executed with the correct parameters
	mock.ExpectExec("UPDATE todo SET title=\\$1, text=\\$2, iscompleted=\\$3, category=\\$4, deadline=\\$5 WHERE id=\\$6").
		WithArgs(todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline, todo.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the function to update the todo
	err = updateTodo(db, todo.ID, &todo)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when calling updateTodo", err)
	}

	// Check that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestDeleteTodo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	todoID := 1

	mock.ExpectExec("DELETE FROM todo WHERE id=\\$1").
		WithArgs(todoID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = deleteTodo(db, todoID)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when calling deleteTodo", err)
	}

	// Check that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Integration Tests

func TestGetAllTodosIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppAndDB()
	// setupAppWithTestDB()
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
	app, db, err := setupAppAndDB()
	// setupAppWithTestDB()
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

func TestMainFunction(t *testing.T) {
	// Run the main function in a separate goroutine to avoid blocking
	go func() {
		main()
	}()

	// Give the server a second to start
	time.Sleep(1 * time.Second)

	// Make a request to ensure the server started
	resp, err := http.Get("http://localhost:4000/api/todos")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

/* REMEMBER TO DO CLEAN UP NEXT TIME */
