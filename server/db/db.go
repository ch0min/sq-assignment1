package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/oTuff/sq-ola1/models"
	"log"
)

func Connect() (*sql.DB, error) {
	connStr := "host=localhost port=5432 user=postgres password=test dbname=todo sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the database!")
	return db, nil
}

func GetAllTodos(db *sql.DB) ([]models.Todo, error) {
	todos := []models.Todo{}
	rows, err := db.Query("SELECT id, title, text, isCompleted, category, deadline FROM todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo models.Todo
		var category sql.NullString
		var deadline sql.NullTime

		err := rows.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done, &category, &deadline)
		if err != nil {
			return nil, err
		}

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

func CreateTodo(db *sql.DB, todo *models.Todo) (int, error) {
	var lastInsertId int
	query := `INSERT INTO todo (title, text, iscompleted, category, deadline)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline).Scan(&lastInsertId)
	return lastInsertId, err
}

func UpdateTodo(db *sql.DB, id int, todo *models.Todo) error {
	query := `UPDATE todo SET title=$1, text=$2, iscompleted=$3, category=$4, deadline=$5 WHERE id=$6`
	_, err := db.Exec(query, todo.Title, todo.Body, todo.Done, todo.Category, todo.Deadline, id)
	return err
}

func ToggleTodoStatus(db *sql.DB, id int) error {
	var currentStatus bool
	err := db.QueryRow("SELECT iscompleted FROM todo WHERE id=$1", id).Scan(&currentStatus)
	if err != nil {
		return err
	}

	newStatus := !currentStatus
	_, err = db.Exec("UPDATE todo SET iscompleted=$1 WHERE id=$2", newStatus, id)
	return err
}

func DeleteTodo(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM todo WHERE id = $1", id)
	return err
}
