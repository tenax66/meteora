package main

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tenax66/meteora/meteora-server/database"
	"github.com/tenax66/meteora/shared"
)

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func retrieveMessages(db *sql.DB) ([]byte, error) {
	// TODO: configurable limit
	messages, err := database.SelectMessagesWithLimit(db, 10)
	if err != nil {
		log.Println("Error while selecting all messages", err)
		return nil, err
	}
	jsonData, err := json.Marshal(messages)
	if err != nil {
		log.Println("Error while marshaling messages:", err)
		return nil, err
	}

	return jsonData, nil
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			// Ignore errors caused by the client closing the connection without sending a close message.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Error while reading message:", err)
			}
			break
		}

		// store the received json message
		var message shared.Message

		json.Unmarshal(p, &message)
		log.Println("Received message:", message)

		// verify the attached signature
		ser, err := json.Marshal(message.Content)
		if err != nil {
			log.Println("Error while marshaling the content:", err)
			break
		}

		pubkey, err := hex.DecodeString(message.Pubkey)
		if err != nil {
			log.Println("Cannot decode a public key")
			break
		}

		sig, err := hex.DecodeString(message.Sig)
		if err != nil {
			log.Println("Cannot decode a signature")
			break
		}

		if ed25519.Verify(pubkey, ser, sig) {
			log.Println("Signature verified")
		} else {
			log.Println("Signature verification failed")
			break
		}

		database.InsertMessage(db, message)

		// return messages stored on this server
		jsonData, err := retrieveMessages(db)
		if err != nil {
			log.Println("Error while retrieving messages from database:", err)
			break
		}

		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Println("Error while writing message:", err)
			break
		}
	}
}

func handleFetch(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")
	jsonData, err := retrieveMessages(db)
	if err != nil {
		log.Println("Error while retrieving messages from database:", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Println("Error while writing message:", err)
		return
	}
}

func main() {
	db, _ = database.CreateDB("./meteora.db")

	http.HandleFunc("/ws/send", handleSend)
	http.HandleFunc("/ws/fetch", handleFetch)

	port := ":8080"
	log.Println("WebSocket server started on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error while starting server:", err)
	}
}
