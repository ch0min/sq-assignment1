// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	_ "github.com/lib/pq"
// 	"log"
// 	"time"
// )

// type Tododd struct {
// 	Id          *int
// 	Title       string
// 	Text        string
// 	IsCompleted bool
// 	Category    *string
// 	Deadline    *time.Time
// }

// func mainmain() {
// 	// Setup db connection
// 	connStr := "host=localhost port=5432 user=postgres password=test dbname=todo sslmode=disable"

// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	err = db.Ping()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Successfully connected to the database!")

// 	// testDate := time.Date(2024, 9, 10, 0, 0, 0, 0, time.UTC)
// 	// id := 2
// 	// category := "testcat"
// 	// newTodo := Todo{&id, "nytest", "test", false, &category, &testDate}
// 	// newTodo := Todo{nil, "newTodo", "test", false, nil, nil}

// 	// createTodo(db, newTodo)
// 	// getAllTodos(db)
// 	// deleteTodo(db, 1)
// 	// updateTodo(db, newTodo)

// }

// func createTodo(db *sql.DB, todo Todo) {
	// query := `
    //     INSERT INTO todo (title, text, category, deadline)
    //     VALUES ($1, $2, $3, $4)
    // `
// 	_, err := db.Exec(query, todo.Title, todo.Text, todo.Category, todo.Deadline)
// 	if err != nil {
// 		log.Fatal(err)
// 	} else {
// 		fmt.Println("Row inserted successfully!")
// 	}
// }

// func getAllTodos(db *sql.DB) {
// 	rows, err := db.Query("SELECT id, title, text, isCompleted, category, deadline FROM todo")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var id int
// 		var title string
// 		var text string
// 		var isCompleted bool
// 		var category string
// 		var deadline time.Time

// 		err = rows.Scan(&id, &title, &text, &isCompleted, &category, &deadline)
// 		if err != nil {
// 			log.Fatal(err)
// 		} else {
// 			// TODO: cannot handle null when printing
// 			fmt.Printf("ID: %d, title: %s, text: %s, isCompleted: %t, category: %s, deadline: %s\n", id, title, text, isCompleted, category, deadline)
// 		}
// 	}

// 	// Check for errors after iterating
// 	if err = rows.Err(); err != nil {
// 		log.Fatal(err)
// 	}
// }

// func deleteTodo(db *sql.DB, id int) {
// 	_, err := db.Exec("Delete FROM todo WHERE id=$1", id)
// 	if err != nil {
// 		log.Fatal(err)
// 	} else {
// 		fmt.Println("Deleted todo")
// 	}

// }

// func updateTodo(db *sql.DB, todo Todo) {
// 	_, err := db.Exec("UPDATE todo SET title=$2, text=$3, isCompleted=$4, category=$5, deadline=$6 WHERE id=$1", todo.Id, todo.Title, todo.Text, todo.IsCompleted, todo.Category, todo.Deadline)

// 	if err != nil {
// 		fmt.Println(err)
// 	} else {
// 		fmt.Println("Updated Todo")
// 	}
// }