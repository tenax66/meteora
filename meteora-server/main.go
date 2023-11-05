package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Id      string  `json:"id"`
	Content Content `json:"content"`
}

type Content struct {
	Created_at int64  `json:"timestamp"`
	Text       string `json:"text"`
}

var messages []Message
var mu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
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
			return
		}

		// store the received json message
		var message Message

		json.Unmarshal(p, &message)
		log.Println("Received message:", message)

		mu.Lock()
		messages = append(messages, message)
		mu.Unlock()

		// return messages stored on this server
		mu.Lock()
		jsonData, err := json.Marshal(messages)
		if err != nil {
			log.Println("Error while marshaling messages:", err)
		}

		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Println("Error while writing message:", err)
			mu.Unlock()
			return
		}
		mu.Unlock()
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocketConnection)

	port := ":8080"
	log.Println("WebSocket server started on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
