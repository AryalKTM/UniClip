package main

import (
	"fmt"
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
		systray.SetTooltip("ClipSync - Universal Clipboard")

		mStart := systray.AddMenuItem("Start Server", "Start the Clipboard Server")
		mConnect := systray.AddMenuItem("Connect to Server", "Connect to Existing Server")
		systray.AddSeparator()
		mEncrypt := systray.AddMenuItemCheckbox((fmt.Sprintf("Secure Connection: %v", secure)), "Enable End-to-End Encryption", false)
		// mStartonLogin := systray.AddMenuItemCheckbox("Start on Login", "Start ClipSync on Login", false)
		mQuit := systray.AddMenuItem("Quit", "Quit ClipSync")

		go func() {
			for {
				select {
				case <-mStart.ClickedCh:
					go makeServer()
				case <-mConnect.ClickedCh:
					go showConnectDialog()
				// case <- mStartonLogin.ClickedCh:
				// 	mStartonLogin.Check()
				case <- mEncrypt.ClickedCh:
					secure = !secure
					fmt.Println(secure)
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
	ipEntry.SetPlaceHolder("192.168.1.1:8080")

	content := container.NewVBox(
		widget.NewLabel("Connect"),
		ipEntry,
		widget.NewButton("Connect", func() {
			go ConnectToServer(ipEntry.Text)
			window.Close()
			app.Quit()
		}),
	)
	window.SetContent(content)
	window.ShowAndRun()
}