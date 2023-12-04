package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tenax66/meteora/shared"
)

const DB = "sqlite3"

// Create database if it does not exist
func CreateDB(dbPath string) (*sql.DB, error) {
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
func InsertMessage(db *sql.DB, message shared.Message) error {
	insertQuery := "INSERT INTO messages (id, created_at, text, pubkey, sig) VALUES (?, ?, ?, ?, ?)"

	_, err := db.Exec(insertQuery, message.Id, message.Content.Created_at, message.Content.Text, message.Pubkey, message.Sig)

	if err != nil {
		log.Println("Error while inserting", err)
		return err
	}

	return nil
}

func SelectMessageById(id string, db *sql.DB) (shared.Message, error) {
	query := "SELECT id, created_at, text, pubkey, sig FROM messages WHERE id = ?"

	row := db.QueryRow(query, id)

	var message shared.Message
	err := row.Scan(&message.Id, &message.Content.Created_at, &message.Content.Text, &message.Pubkey, &message.Sig)
	if err != nil {
		log.Println("Error while scanning a row", err)
	}

	return message, nil
}

func SelectMessagesWithLimit(db *sql.DB, limit int) ([]shared.Message, error) {
	query := "SELECT id, created_at, text, pubkey, sig FROM messages ORDER BY created_at DESC LIMIT ?"
	rows, err := db.Query(query, limit)

	if err != nil {
		log.Println("Error while selecting", err)
		return nil, err
	}

	// scan messages
	var messages []shared.Message
	for rows.Next() {
		var message shared.Message
		err := rows.Scan(&message.Id, &message.Content.Created_at, &message.Content.Text, &message.Pubkey, &message.Sig)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		messages = append(messages, message)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return messages, nil
}

// Deprecated: use SelectMessagesWithLimit instead.
func SelectAllMessages(db *sql.DB) ([]shared.Message, error) {
	query := "SELECT id, created_at, text, pubkey, sig FROM messages"
	rows, err := db.Query(query)

	if err != nil {
		log.Println(err)
	}

	// scan all messages
	var messages []shared.Message
	for rows.Next() {
		var message shared.Message
		err := rows.Scan(&message.Id, &message.Content.Created_at, &message.Content.Text, &message.Pubkey, &message.Sig)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		messages = append(messages, message)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return messages, nil
}
