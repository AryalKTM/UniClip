package main

import (
	"fmt"
	"runtime"
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