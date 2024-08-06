package main

import (
	"fmt"
	"os/exec"

	"fyne.io/systray"
)

func startMacOSMenuBar() {
	onReady := func() {
		systray.SetIcon(getIcon("icon.png"))
		systray.SetTooltip("Universal Clipboard")

		mStart := systray.AddMenuItem("Start Server", "Start Clipboard Server")
		mConnect := systray.AddMenuItem("Connect to Server", "Connect to Existing Clipboard Server")
		mStartOnLogin := systray.AddMenuItem("Start on Login", "Start Application on Login")
		mQuit := systray.AddMenuItem("Quit", "Quit Application")

		go func() {
			for {
				select {
				case <-mStart.ClickedCh:
					go makeServer()
				case <- mConnect.ClickedCh:
					go ConnectToServer("aryal-macbook:`port`")
				case <- mStartOnLogin.ClickedCh:
					go addToMacOSLoginItems()
				case <- mQuit.ClickedCh:
					systray.Quit()
				}
			}
		}()
	}

	onExit := func() {
		//Exits
	}

	systray.Run(onReady, onExit)
}



func addToMacOSLoginItems() {
	//TODO: Add PATH to Installed Application
	script := `tell application "System Events" to make login item at end with properties {path:"/path/to/your/app", hidden:false}`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to add to login items:", err)
	} else {
		fmt.Println("Successfully added to login items")
	}
}