package db

import (
	"database/sql"
	"fmt"
	"go_final_project/models"
	"log"
	"os"
	"path/filepath"
	"strconv"

	_ "modernc.org/sqlite" // Подключаем SQLite без CGO
)

func GetDatabasePath() string {
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}
		dbPath = filepath.Join(workingDir, "scheduler.db")
	}
	return dbPath
}

func SetupDatabase(dbFile string) error {
	_, err := os.Stat(dbFile)
	var install bool
	if err != nil && os.IsNotExist(err) {
		install = true
		log.Println("Database file not found, creating an empty database file.")

		file, err := os.Create(dbFile)
		if err != nil {
			return fmt.Errorf("failed to create database file: %v", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if install {
		err = createTable(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func createTable(db *sql.DB) error {
	log.Println("Creating table 'scheduler'...")
	query := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT CHECK(length(repeat) <= 128)
	);
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return err
	}
	log.Println("Table 'scheduler' created successfully.")
	return nil
}

func AddTask(db *sql.DB, date, title, comment, repeat string) (int64, error) {
	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`
	res, err := db.Exec(query, date, title, comment, repeat)
	if err != nil {
		log.Printf("Failed to insert task: %v", err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve last insert ID: %v", err)
		return 0, err
	}

	return id, nil
}

func GetTaskByID(db *sql.DB, id int) (*models.Task, error) {
	var task models.Task
	row := db.QueryRow(
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?",
		id,
	)

	var taskID int64
	err := row.Scan(&taskID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	task.ID = strconv.FormatInt(taskID, 10)
	return &task, nil
}

func UpdateTask(db *sql.DB, task models.Task) (int64, error) {
	query := `
		UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?
	`
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func DeleteTask(db *sql.DB, id int) (int64, error) {
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
