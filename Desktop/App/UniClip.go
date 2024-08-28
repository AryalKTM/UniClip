package main

import (
	"bufio"
	"fmt"
	"runtime"
)

var (
	secondsBetweenChecksForClipChange = 1

	listOfClients  = make([]*bufio.Writer, 0)
	localClipboard string
	printDebugInfo = false
	cryptoStrength = 16384
	secure         = false
	password       []byte
	port           = 8091
)

// TODO: Add a way to reconnect (if computer goes to sleep)
func main() {
	fmt.Println("*****Welcome to ClipSync*****")
	switch runtime.GOOS {
	case "windows":
		startWindowsSystemTray()
	case "darwin":
		startWindowsSystemTray()
	default:
		runTerminalApp()
	}
}
