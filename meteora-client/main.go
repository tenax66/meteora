package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	serverAddr := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr.String(), nil)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	message := "Hello, WebSocket Server!"

	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}

	fmt.Println("Message sent to server:", message)

	// wait for a server response
	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error reading response:", err)
		return
	}

	fmt.Println("Response from server:", string(response))
}
