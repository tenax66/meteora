package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tenax66/meteora/shared"
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

// Insert the given message into the database.
func insertMessage(db *sql.DB, message shared.Message) error {
	insertQuery := "INSERT INTO messages (id, created_at, text, pubkey, sig) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(insertQuery, message.Id, message.Content.Created_at, message.Content.Text, message.Pubkey, message.Sig)
	if err != nil {
		log.Println("Error while inserting", err)
		return err
	}
	return nil
}
