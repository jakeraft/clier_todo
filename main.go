package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	db, err := openDB("todo.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: todo add <title>")
			os.Exit(1)
		}
		title := strings.Join(os.Args[2:], " ")
		if err := addTodo(db, title); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		showDone := len(os.Args) > 2 && os.Args[2] == "--done"
		if err := listTodos(db, showDone); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "done":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: todo done <id>")
			os.Exit(1)
		}
		if err := doneTodo(db, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: todo delete <id>")
			os.Exit(1)
		}
		if err := deleteTodo(db, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("todo - a simple task manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo add <title>    Add a new todo")
	fmt.Println("  todo list [--done]  List all todos")
	fmt.Println("  todo done <id>      Mark a todo as done")
	fmt.Println("  todo delete <id>    Delete a todo")
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		done INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func addTodo(db *sql.DB, title string) error {
	result, err := db.Exec("INSERT INTO todos (title) VALUES (?)", title)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	fmt.Printf("Added todo #%d: %s\n", id, title)
	return nil
}

func listTodos(db *sql.DB, showDone bool) error {
	var rows *sql.Rows
	var err error
	if showDone {
		rows, err = db.Query("SELECT id, title, done FROM todos WHERE done = 1")
	} else {
		rows, err = db.Query("SELECT id, title, done FROM todos WHERE done = 0")
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var title string
		var done int
		if err := rows.Scan(&id, &title, &done); err != nil {
			return err
		}
		check := " "
		if done == 1 {
			check = "x"
		}
		fmt.Printf("  [%s] #%d: %s\n", check, id, title)
		count++
	}
	if count == 0 {
		if showDone {
			fmt.Println("No completed todos.")
		} else {
			fmt.Println("No todos yet. Add one with: todo add <title>")
		}
	}
	return nil
}

func doneTodo(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE todos SET done = 1 WHERE id = ?", id)
	if err != nil {
		return err
	}
	fmt.Printf("Marked todo #%s as done\n", id)
	return nil
}

func deleteTodo(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted todo #%s\n", id)
	return nil
}
