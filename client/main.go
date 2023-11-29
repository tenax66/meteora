package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/gorilla/websocket"
	"github.com/tenax66/meteora/shared"
)

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
	meteoraApp := app.NewWithID("meteora")
	myWindow := meteoraApp.NewWindow("WebSocket Client")
	myWindow.Resize(fyne.NewSize(800, 600))

	// entries
	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Enter Host Address")
	privateKeyEntry := widget.NewEntry()
	privateKeyEntry.SetPlaceHolder("Enter Private Key Path")
	publicKeyEntry := widget.NewEntry()
	publicKeyEntry.SetPlaceHolder("Enter Public Key Path")
	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Type your message here...")

	// get values from preferences

	loadPreferences(meteoraApp, []PreferencesPair{
		{Entry: addressEntry, Key: "address"},
		{Entry: privateKeyEntry, Key: "privateKey"},
		{Entry: publicKeyEntry, Key: "publicKey"},
	})

	var messages []shared.Message

	messageList := widget.NewList(
		func() int {
			// Return the number of messages
			return len(messages)
		},
		func() fyne.CanvasObject {
			// Return a template for each item in the list
			return container.NewVBox(
				widget.NewLabel("Timestamp:"),
				widget.NewLabel("Content:"),
				// Add more labels or widgets as needed
			)
		},
		func(i widget.ListItemID, item fyne.CanvasObject) {
			// Populate the template with data for each item in the list
			timestampLabel := item.(*fyne.Container).Objects[0].(*widget.Label)
			contentLabel := item.(*fyne.Container).Objects[1].(*widget.Label)

			// Use messages[i] to populate the labels with actual data
			timestampLabel.SetText(fmt.Sprintf("Timestamp: %v", messages[i].Content.Created_at))
			contentLabel.SetText(fmt.Sprintf("Content: %v", messages[i].Content.Text))
			// Update other labels or widgets as needed
		})

	messageList.Resize(fyne.NewSize(700, 400))

	sendButton := widget.NewButton("Send", func() {
		addr := addressEntry.Text
		privateKeyPath := privateKeyEntry.Text
		publicKeyPath := publicKeyEntry.Text
		messageText := messageEntry.Text

		// save preferences
		savePreferences(meteoraApp, []PreferencesPair{
			{Entry: addressEntry, Key: "address"},
			{Entry: privateKeyEntry, Key: "privateKey"},
			{Entry: publicKeyEntry, Key: "publicKey"},
		})

		k, p, err := shared.ReadKeys(privateKeyPath, publicKeyPath)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		message := createMessage(messageText, k, p)
		jsonData, err := json.Marshal(message)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		serverAddr := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
		conn, _, err := websocket.DefaultDialer.Dial(serverAddr.String(), nil)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		defer conn.Close()

		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		log.Println("Message sent to server:", message)

		// Wait for a server response
		_, response, err := conn.ReadMessage()
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		messages = parseResponse(response)

		messageList.Refresh()

	})

	reloadButton := widget.NewButton("Reload", func() {
		// mock
		log.Println("Reload button clicked")
	})

	buttonBox := container.NewHBox(sendButton, reloadButton)

	top := container.NewVBox(
		addressEntry,
		privateKeyEntry,
		publicKeyEntry,
		messageEntry,
		buttonBox,
	)

	content := container.NewBorder(top, nil, nil, nil, messageList)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()

}
