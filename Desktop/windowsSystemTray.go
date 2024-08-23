package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/systray"
)

func startWindowsSystemTray() {
	onReady := func() {
		systray.SetIcon(getIcon("icon.png"))
		systray.SetTitle("ClipSync")
		systray.SetTooltip("ClipSync - Universal Clipboard")

		mStart := systray.AddMenuItem("Start Server", "Start the Clipboard Server")
		systray.AddSeparator()
		mConnect := systray.AddMenuItem("Connect to Server", "Connect to Existing Server")
		// mStartonLogin := systray.AddMenuItemCheckbox("Start on Login", "Start ClipSync on Login", false)
		mQuit := systray.AddMenuItem("Quit", "Quit ClipSync")

		go func() {
			for {
				select {
				case <-mStart.ClickedCh:
					go makeServer()
				case <-mConnect.ClickedCh:
					go ConnectToServer("localhost:8091")
				// case <- mStartonLogin.ClickedCh:
				// 	mStartonLogin.Check()
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

func showConnectDialog() {
	app := app.New()
	window := app.NewWindow("Connect to Server")

	ipEntry := widget.NewEntry()
	ipEntry.setPlaceHolder("192.168.1.1:8080")

	content := container.NewVBox(
		widget.NewLabel("Connect"),
		ipEntry,
		widget.NewButton("Connect", func() {
			go ConnectToServer(ipEntry.Text)
			window.close()
			app.Quit()
		}),
	)
	window.setContent(content)
	window.ShowAndRun()
}
