package main

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
)

func main() {
	var lastClipboard string

	for {
		// Read clipboard content
		clipboardContent, err := clipboard.ReadAll()
		if err != nil {
			fmt.Println("Error reading clipboard:", err)
		} else if clipboardContent != lastClipboard {
			// Print clipboard content if it's changed
			fmt.Println("Clipboard content:", clipboardContent)
			lastClipboard = clipboardContent
		}

		// Sleep for a short interval before checking again
		time.Sleep(1 * time.Second)
	}
}
