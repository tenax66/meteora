package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var messages []string
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

		// store the received message
		message := string(p)
		log.Println("Received message:", message)

		mu.Lock()
		messages = append(messages, message)
		mu.Unlock()

		// return messages stored on this server
		mu.Lock()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(strings.Join(messages, ","))); err != nil {
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
