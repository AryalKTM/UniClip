package main

import (
	"log"
	"os/exec"
	
	"fyne.io/systray"
	"golang.org/x/sys/windows/registry"
)

func startWindowsSystemTray() {
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
				case <-mConnect.ClickedCh:
					go ConnectToServer("192.168.12.76:51054")
				case <-mStartOnLogin.ClickedCh:
					go addToWindowsStartupTask()
				case <-mQuit.ClickedCh:
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

func addToWindowsStartupTask() {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		log.Println("Failed to Open Registry Key:", err)
		return
	}
	defer key.Close()

	exePath, err := exec.LookPath("UniClip.exe")
	if err != nil {
		log.Println("Failed to Find Application Path:", err)
		return
	}

	err = key.SetStringValue("UniClip", exePath)
	if err != nil {
		log.Println("Failed to Add Application to Startup:", err)
	} else {
		log.Println("Successfully Added to Startup")
	}
}
