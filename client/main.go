package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/gorilla/websocket"
	"github.com/tenax66/meteora/shared"
)

var messages []shared.Message
var currentPage = 1

const MESSAGES_PER_PAGE = 5

func parseResponse(response []byte) []shared.Message {
	var m []shared.Message
	if err := json.Unmarshal(response, &m); err != nil {
		log.Println("Error while unmarshalling message", err)
	}

	return m
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

func updateMessageList(addr string, window fyne.Window, limit int, offset int) ([]byte, error) {
	u := "/ws/fetch"
	params := url.Values{}

	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))

	serverAddr := url.URL{Scheme: "ws", Host: addr, Path: u + "/?" + params.Encode()}
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr.String(), nil)
	if err != nil {
		dialog.ShowError(err, window)
		return nil, err
	}
	defer conn.Close()

	// Wait for a server response
	_, response, err := conn.ReadMessage()
	if err != nil {
		dialog.ShowError(err, window)
		return nil, err
	}

	return response, nil
}

// An entry point for meteora client.
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

	messageList := widget.NewList(
		func() int {
			// Return the number of messages
			return len(messages)
		},
		func() fyne.CanvasObject {
			// Return a template for each item in the list
			return container.NewVBox(
				widget.NewLabel("Timestamp:"),
				widget.NewLabel("Pubkey:"),
				widget.NewLabel("Content:"),
				// Add more labels or widgets as needed
			)
		},
		func(i widget.ListItemID, item fyne.CanvasObject) {
			// Populate the template with data for each item in the list
			timestampLabel := item.(*fyne.Container).Objects[0].(*widget.Label)
			pubkeyLabel := item.(*fyne.Container).Objects[1].(*widget.Label)
			contentLabel := item.(*fyne.Container).Objects[2].(*widget.Label)

			// Use messages[i] to populate the labels with actual data
			timestampLabel.SetText(fmt.Sprintf("Timestamp: %v", time.Unix(messages[i].Content.Created_at, 0)))
			pubkeyLabel.SetText(fmt.Sprintf("Pubkey: %v", messages[i].Pubkey))
			contentLabel.SetText(fmt.Sprintf("Content: %v", messages[i].Content.Text))
			// Update other labels or widgets as needed
		})

	messageList.Resize(fyne.NewSize(700, 400))

	// label for paging
	currentPageLabel := widget.NewLabel(fmt.Sprintf("Page: %d", currentPage))

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

		serverAddr := url.URL{Scheme: "ws", Host: addr, Path: "/ws/send"}
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

		// reset page count
		currentPage = 1
		currentPageLabel.SetText(fmt.Sprintf("Page: %d", currentPage))

		messages = parseResponse(response)
		messageList.Refresh()

	})

	reloadButton := widget.NewButton("Reload", func() {
		response, err := updateMessageList(addressEntry.Text, myWindow, MESSAGES_PER_PAGE, (currentPage-1)*MESSAGES_PER_PAGE)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		messages = parseResponse(response)
		messageList.Refresh()
	})

	// ページング用のナビゲーションボタンを追加
	prevButton := widget.NewButton("Prev Page", func() {
		if currentPage > 1 {
			currentPage--
			response, err := updateMessageList(addressEntry.Text, myWindow, MESSAGES_PER_PAGE, (currentPage-1)*MESSAGES_PER_PAGE)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			messages = parseResponse(response)
			messageList.Refresh()

			currentPageLabel.SetText(fmt.Sprintf("Page: %d", currentPage))
		}
	})

	nextButton := widget.NewButton("Next Page", func() {
		// TODO: add an exit condition for paging
		currentPage++
		response, err := updateMessageList(addressEntry.Text, myWindow, MESSAGES_PER_PAGE, (currentPage-1)*MESSAGES_PER_PAGE)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		messages = parseResponse(response)
		messageList.Refresh()

		currentPageLabel.SetText(fmt.Sprintf("Page: %d", currentPage))
	})

	buttonBox := container.NewHBox(sendButton, reloadButton, prevButton, currentPageLabel, nextButton)

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
