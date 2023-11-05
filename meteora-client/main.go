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
	"time"

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

var addr = flag.String("addr", "localhost:8080", "http service address")

func parseResponse(response []byte) []Message {
	var messages []Message
	if err := json.Unmarshal(response, &messages); err != nil {
		log.Println("Error unmarshalling message", err)
	}

	return messages
}

func createMessage(text string) *Message {
	content := Content{
		Created_at: time.Now().Unix(),
		Text:       text,
	}

	// create a message id by SHA-256 of contents
	hash := sha256.New()

	s, err := json.Marshal(content)
	if err != nil {
		log.Println("Error marshalling content", err)
	}

	log.Println("Serialized content:", string(s))
	hash.Write(s)
	hashInBytes := hash.Sum(nil)
	id := hex.EncodeToString(hashInBytes)

	message := Message{
		Id:      id,
		Content: content,
	}

	return &message
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

	message := createMessage("Hello, WebSocket Server!")

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
