package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tenax66/meteora/shared"
)

var m = flag.String("m", "", "The message to send")
var addr = flag.String("addr", "localhost:8080", "The http service address")
var key = flag.String("key", "", "The path of the private key")
var pubkey = flag.String("pubkey", "", "The path of the public key")

func parseResponse(response []byte) []shared.Message {
	var messages []shared.Message
	if err := json.Unmarshal(response, &messages); err != nil {
		log.Println("Error while unmarshalling message", err)
	}

	return messages
}

// create a message from the given text. Generate and append a message ID and a signature.
func createMessage(text string, key ed25519.PrivateKey, pubkey ed25519.PublicKey) *shared.Message {
	content := shared.Content{
		Created_at: time.Now().Unix(),
		Text:       text,
	}

	ser, err := json.Marshal(content)
	if err != nil {
		log.Println("Error while marshalling content", err)
	}
	log.Println("Serialized content:", string(ser))

	// create a message id by SHA-256 of contents
	hash := sha256.New()
	hash.Write(ser)
	hashInBytes := hash.Sum(nil)
	id := hex.EncodeToString(hashInBytes)

	message := shared.Message{
		Id:      id,
		Content: content,
		Pubkey:  hex.EncodeToString(pubkey),
		Sig:     hex.EncodeToString(ed25519.Sign(key, ser)),
	}

	return &message
}

func main() {
	flag.Parse()

	serverAddr := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr.String(), nil)
	if err != nil {
		log.Fatal("Error while connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	k, p, err := shared.ReadKeys(*key, *pubkey)
	if err != nil {
		log.Println("Error while reading keys:", err)
		return
	}
	message := createMessage(*m, k, p)

	// encoding to json
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Println("Error while encoding JSON:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Println("Error while sending message:", err)
		return
	}

	log.Println("Message sent to server:", message)

	// wait for a server response
	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error while reading response:", err)
		return
	}

	messages := parseResponse(response)

	for i, m := range messages {
		log.Println("Message", i, ": ", m.Content.Text)
	}

}
