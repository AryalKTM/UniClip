// +build darwin

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
	port           = 8085
)

func main() {
	fmt.Println(runtime.GOOS)
	startApplication()
}

func startApplication() {
	switch runtime.GOOS {
	case "darwin":
		startMacOSMenuBar()
	case "windows":
		startWindowsSystemTray()
	}
}