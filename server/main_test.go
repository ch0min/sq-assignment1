package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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
