package main

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
)

func main() {
	var lastClipboard string

	for {
		clipboardContent, err := clipboard.ReadAll()
		if err != nil {
			fmt.Println("Error reading clipboard:", err)
		} else if clipboardContent != lastClipboard {
			fmt.Println("Clipboard content:", clipboardContent)
			lastClipboard = clipboardContent
		}

		time.Sleep(1 * time.Second)
	}
}