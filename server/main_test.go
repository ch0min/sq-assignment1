package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	// "errors"

	"fmt"
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

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected the app to start and respond with 200")
}

func TestSetupAppAndDB_Success(t *testing.T) {
	app, db, err := setupAppAndDB()

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, db)

	if db != nil {
		db.Close()
	}
}

// Validation Business Logic
// func validateTodoInput(todo *Todo) error {
//     if todo.Title == "" {
//         return errors.New("task title must not be empty")
//     }
//     if len(todo.Body) < 10 {
//         return errors.New("task description must have at least 10 characters")
//     }
//     return nil
// }

func TestValidateTodoInput(t *testing.T) {
    // Case 1: Title is empty
    todo := &Todo{
        Title: "",
        Body:  "This is a valid description",
    }
    err := validateTodoInput(todo)
    assert.EqualError(t, err, "task title must not be empty", "Expected an error for empty title")

    // Case 2: Description is too short
    todo = &Todo{
        Title: "Valid Title",
        Body:  "Short",
    }
    err = validateTodoInput(todo)
    assert.EqualError(t, err, "task description must have at least 10 characters", "Expected an error for short description")

    // Case 3: Both title and description are valid
    todo = &Todo{
        Title: "Valid Title",
        Body:  "This is a valid description",
    }
    err = validateTodoInput(todo)
    assert.NoError(t, err, "Expected no error for valid title and description")
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

	err = updateTodo(db, todo.ID, &todo)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when calling updateTodo", err)
	}

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Integration Tests

func TestGetAllTodosIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppAndDB()
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

func TestGetTodoIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppAndDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Define the new todo to be created
	newTodo := Todo{
		Title: "Test Todo",
		Body:  "This is a test todo",
		Done:  false,
	}

	// Insert the todo into the database
	_, err = db.Exec("INSERT INTO todo (title, text, iscompleted) VALUES ($1, $2, $3)", newTodo.Title, newTodo.Body, newTodo.Done)
	assert.NoError(t, err)

	// Retrieve ID
	row := db.QueryRow("SELECT id FROM todo WHERE title = $1 AND text = $2", newTodo.Title, newTodo.Body)
	var todoID int
	err = row.Scan(&todoID)
	assert.NoError(t, err)

	// Create a test HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/todos/%d", todoID), nil)

	resp, err := app.Test(req, -1) // -1 disables the timeout
	assert.NoError(t, err)

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Response Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(respBody))

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected a 200 OK status code")

	var todo Todo
	err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&todo)
	assert.NoError(t, err)

	assert.Equal(t, newTodo.Title, todo.Title)
	assert.Equal(t, newTodo.Body, todo.Body)
}

func TestUpdateTodoIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppAndDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	initialTodo := Todo{
		Title:    "Initial Title",
		Body:     "Initial Body",
		Done:     false,
		Category: func() *string { s := "Initial Category"; return &s }(),
	}

	initialTodoJSON, err := json.Marshal(initialTodo)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(initialTodoJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Response Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(respBody))

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected a 201 Created status code")

	var createdTodo Todo
	err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&createdTodo)
	assert.NoError(t, err)
	createdTodoID := createdTodo.ID

	updatedTodo := Todo{
		Title:    "Updated Title",
		Body:     "Updated Body",
		Done:     true,
		Category: func() *string { s := "Updated Category"; return &s }(),
	}

	updatedTodoJSON, err := json.Marshal(updatedTodo)
	assert.NoError(t, err)

	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/todos/%d", createdTodoID), bytes.NewBuffer(updatedTodoJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err = app.Test(req, -1)
	assert.NoError(t, err)

	respBody, _ = io.ReadAll(resp.Body)
	t.Logf("Response Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(respBody))

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected a 200 OK status code")

	var todo Todo
	err = db.QueryRow("SELECT title, text, iscompleted, category FROM todo WHERE id=$1", createdTodoID).Scan(&todo.Title, &todo.Body, &todo.Done, &todo.Category)

	if err != nil {
		t.Errorf("Failed to query todo from database: %v", err)
	}

	assert.NoError(t, err)
	assert.Equal(t, updatedTodo.Title, todo.Title)
	assert.Equal(t, updatedTodo.Body, todo.Body)
	assert.Equal(t, updatedTodo.Done, todo.Done)
	assert.Equal(t, *updatedTodo.Category, *todo.Category)
}

func TestToggleTodoStatusIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppAndDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	initialTodo := Todo{
		Title:    "Toggle Test Title",
		Body:     "Toggle Test Body",
		Done:     false,
		Category: func() *string { s := "Toggle Test Category"; return &s }(),
	}

	initialTodoJSON, err := json.Marshal(initialTodo)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(initialTodoJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Response Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(respBody))

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected a 201 Created status code")

	var createdTodo Todo
	err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&createdTodo)
	assert.NoError(t, err)
	createdTodoID := createdTodo.ID

	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/todos/%d/done", createdTodoID), nil)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected a 204 No Content status code")

	var todo struct {
		IsCompleted bool `db:"iscompleted"`
	}
	err = db.QueryRow("SELECT iscompleted FROM todo WHERE id=$1", createdTodoID).Scan(&todo.IsCompleted)

	if err != nil {
		t.Errorf("Failed to query todo from database: %v", err)
	}

	assert.NoError(t, err)
	// The initial value of IsCompleted was false, so after toggling it should be true
	assert.True(t, todo.IsCompleted, "Todo status should be toggled to true")

	// Toggle again to test reverting
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/todos/%d/done", createdTodoID), nil)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected a 204 No Content status code")

	err = db.QueryRow("SELECT iscompleted FROM todo WHERE id=$1", createdTodoID).Scan(&todo.IsCompleted)
	assert.NoError(t, err)
	// After toggling again, it should revert to false
	assert.False(t, todo.IsCompleted, "Todo status should be toggled back to false")
}

func TestDeleteTodoIntegration(t *testing.T) {
	// Setup the Fiber app and the test database
	app, db, err := setupAppAndDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	initialTodo := Todo{
		Title:    "Delete Test Title",
		Body:     "Delete Test Body",
		Done:     false,
		Category: func() *string { s := "Delete Test Category"; return &s }(),
	}

	initialTodoJSON, err := json.Marshal(initialTodo)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(initialTodoJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Response Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(respBody))

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected a 201 Created status code")

	var createdTodo Todo
	err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&createdTodo)
	assert.NoError(t, err)
	createdTodoID := createdTodo.ID

	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/todos/%d", createdTodoID), nil)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected a 204 No Content status code")

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM todo WHERE id=$1)", createdTodoID).Scan(&exists)
	if err != nil {
		t.Errorf("Failed to query todo existence from database: %v", err)
	}

	assert.NoError(t, err)
	assert.False(t, exists, "Todo should have been deleted from the database")
}

func TestMainFunction(t *testing.T) {
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

// /* REMEMBER TO DO CLEAN UP NEXT TIME */
