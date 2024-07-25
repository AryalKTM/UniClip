package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/atotto/clipboard"
)

type ClipboardResponse struct {
	PayloadData string `json:"payloadData"`
}

func main() {
	// URL to send the GET request
	fmt.Println("Starting...")
	url := "http://100.83.93.64:3000/clipboard/latest"

	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %v", resp.StatusCode)
	}

	// Decode JSON response
	var clipboardResponse ClipboardResponse
	err = json.NewDecoder(resp.Body).Decode(&clipboardResponse)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	// Get the payloadData from the response
	payloadData := clipboardResponse.PayloadData
	fmt.Println("Received payloadData:", payloadData)

	// Set the payloadData to clipboard
	err = clipboard.WriteAll(payloadData)
	if err != nil {
		log.Fatalf("Error setting clipboard content: %v", err)
	}

	fmt.Println("Successfully set clipboard content:", payloadData)
}
