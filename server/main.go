package main

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/tenax66/meteora/shared"
)

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// retrieves messages from the database with givem limit and offset.
func retrieveMessages(db *sql.DB, limit int, offset int) ([]byte, error) {
	// TODO: configurable limit
	messages, err := SelectMessages(db, limit, offset)
	if err != nil {
		log.Println("Error while selecting messages", err)
		return nil, err
	}
	jsonData, err := json.Marshal(messages)
	if err != nil {
		log.Println("Error while marshaling messages:", err)
		return nil, err
	}

	return jsonData, nil
}

// verifies the signature of the given message and returns the result as a boolean value.
func verifyMessage(message shared.Message) (bool, error) {
	// verify the attached signature
	ser, err := json.Marshal(message.Content)
	if err != nil {
		log.Println("Error while marshaling the content:", err)
		return false, err
	}

	pubkey, err := hex.DecodeString(message.Pubkey)
	if err != nil {
		log.Println("Cannot decode a public key")
		return false, err
	}

	sig, err := hex.DecodeString(message.Sig)
	if err != nil {
		log.Println("Cannot decode a signature")
		return false, err
	}

	return ed25519.Verify(pubkey, ser, sig), nil
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

		// verify the signature of the message
		if t, err := verifyMessage(message); err != nil {
			log.Println("Error while verifying message:", err)
		} else if t {
			log.Println("Signature verified")
		} else {
			log.Println("Signature does not verified:")
			break
		}

		InsertMessage(db, message)

		// return messages stored on this server
		jsonData, err := retrieveMessages(db, 10, 0)
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

	// parse query strings
	// FIXME: redundant unescaping her
	unescapedUrl, _ := url.QueryUnescape(r.URL.String())
	parsedUrl, _ := url.Parse(unescapedUrl)

	log.Println("query:", parsedUrl.Query())
	limitParam := parsedUrl.Query().Get("limit")
	log.Println("limit:", limitParam)
	offsetParam := parsedUrl.Query().Get("offset")
	log.Println("offset:", offsetParam)

	// set default values if parameters are invalid
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		log.Println("Error while parsing `limit` parameter:", err)
		limit = 10
	}

	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		log.Println("Error while parsing `offset` parameter:", err)
		offset = 0
	}

	log.Println("Client connected")
	jsonData, err := retrieveMessages(db, limit, offset)
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
	db, _ = CreateDB("./meteora.db")

	http.HandleFunc("/ws/send", handleSend)
	// XXX: slash after "fetch" is needed
	http.HandleFunc("/ws/fetch/", handleFetch)

	port := ":8080"
	log.Println("WebSocket server started on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error while starting server:", err)
	}
}
