package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const DB = "sqlite3"

// Create database if it does not exist
func createDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open(DB, dbPath)
	if err != nil {
		log.Println("Error while opening a database", err)
		return nil, err
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY, 
			created_at INTEGER,
			text TEXT, 
			pubkey TEXT,
			sig TEXT
		)`)
	if err != nil {
		log.Println("Failed to create a table:", err)
		return nil, err
	}

	return db, nil
}
