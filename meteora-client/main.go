package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func parseResponse(response []byte) []Message {
	var messages []Message
	if err := json.Unmarshal(response, &messages); err != nil {
		log.Println("Error unmarshalling message", err)
	}

	return messages
}

func main() {
	flag.Parse()

	serverAddr := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr.String(), nil)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	text := "Hello, WebSocket Server!"

	// create a message id by SHA-256 of contents
	hash := sha256.New()
	hash.Write([]byte(text))
	hashInBytes := hash.Sum(nil)
	id := hex.EncodeToString(hashInBytes)

	message := Message{
		Id:   id,
		Text: text,
	}

	// encoding to json
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}

	log.Println("Message sent to server:", message)

	// wait for a server response
	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error reading response:", err)
		return
	}

	fmt.Println(parseResponse(response))

}
