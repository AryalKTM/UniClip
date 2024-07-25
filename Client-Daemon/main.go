package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/atotto/clipboard"
)

func main() {
	var lastClipboard string
	destinationFolder := "./copied_files"

	if err := os.MkdirAll(destinationFolder, 0755); err != nil {
		fmt.Println("Error creating destination folder:", err)
		return
	}

	for {
		clipboardContent, err := clipboard.ReadAll()
		if err != nil {
			fmt.Println("Error reading clipboard:", err)
		} else if clipboardContent != lastClipboard {
			fmt.Println("Clipboard content:", clipboardContent)
			lastClipboard = clipboardContent

			if isFilePath(clipboardContent) {
				// Copy the file to the destination folder
				err := copyFile(clipboardContent, destinationFolder)
				if err != nil {
					fmt.Println("Error copying file:", err)
				} else {
					fmt.Println("File copied successfully.")
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func isFilePath(s string) bool {
	regex := `^([a-zA-Z]:)?(\\[^<>:"/\\|?*]*)+\\?$`
	matched, err := regexp.MatchString(regex, s)
	if err != nil {
		fmt.Println("Error matching regex:", err)
		return false
	}
	return matched
}

func copyFile(sourcePath, destinationFolder string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceFileName := filepath.Base(sourcePath)

	destinationFilePath := filepath.Join(destinationFolder, sourceFileName)

	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
