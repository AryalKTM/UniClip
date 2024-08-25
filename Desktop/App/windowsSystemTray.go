package main

import (
	"log"
	"os"

	"fyne.io/systray"
)

func startWindowsSystemTray() {
	db, err := initDB()
	if err != nil {
		handleError(err)
	}

	ipAddress, err := fetchIPAddressFromDB(db)
	if err != nil {
		handleError(err)
	}

	onReady := func() {
		systray.SetIcon(getIcon("icon.png"))
		systray.SetTooltip("ClipSync - Universal Clipboard")

		mStart := systray.AddMenuItem("Start Server", "Start the Clipboard Server")

		mConnect := systray.AddMenuItem("Connect to Server", "Connect to Saved Server")

		systray.AddSeparator()

		mQuit := systray.AddMenuItem("Quit", "Quit ClipSync")

		go func() {
			for {
				select {
				case <-mStart.ClickedCh:
					go makeServer()
				case <-mConnect.ClickedCh:
					go ConnectToServer(ipAddress)
				case <-mQuit.ClickedCh:
					systray.Quit()
				}
			}
		}()
	}

	onExit := func() {

	}
	systray.Run(onReady, onExit)
}

func getIcon(iconPath string) []byte {
	icon, err := os.ReadFile(iconPath)
	if err != nil {
		log.Fatal(err)
	}
	return icon
}
