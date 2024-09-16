package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var (
	secondsBetweenChecksForClipChange = 1
	listOfClients                     = make([]*bufio.Writer, 0)
	localClipboard                    string
	printDebugInfo                    = false
	cryptoStrength                    = 16384
	secure                            = false
	password                          []byte
	port                              = 8091
)

// TODO: Add a way to reconnect (if computer goes to sleep)
func main() {
	db, err := initDB()
	if err != nil {
		handleError(err)
	}

	ipAddress, err := fetchIPAddressFromDB(db)
	if err != nil {
		handleError(err)
	}

	clipSyncApp := app.New()
	clipSyncWindow := clipSyncApp.NewWindow("ClipSync")

	iconResource, err := loadIcon("../icon.png")
	if err != nil {
		fmt.Println("Error Loading Icon:", err)
		return
	}
	clipSyncWindow.SetIcon(iconResource)

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter IP Address")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Enter Port")

	// Create a button to update the IP address and port in the database
	updateButton := widget.NewButton("Save Address", func() {
		ip := ipEntry.Text
		port := portEntry.Text

		if err := validateIP(ip); err != nil {
			showErrorDialog(clipSyncWindow, err.Error())
			return
		}

		if err := validatePort(port); err != nil {
			showErrorDialog(clipSyncWindow, err.Error())
			return
		}

		combinedInfo := combineIPAndPort(ip, port)
		err := updateServerInfo(combinedInfo)
		if err != nil {
			showErrorDialog(clipSyncWindow, err.Error())
		} else {
			showSuccessDialog(clipSyncWindow, "Server info updated successfully.")
		}
	})

	content := container.NewVBox(
		widget.NewLabel("IP Address:"),
		ipEntry,
		widget.NewLabel("Port:"),
		portEntry,
		updateButton,
	)

	clipSyncWindow.SetContent(content)
	clipSyncWindow.Resize(fyne.NewSize(400, 500))
	clipSyncWindow.SetCloseIntercept(func() {
		clipSyncWindow.Hide()
	})

	if deskEnv, ok := clipSyncApp.(desktop.App); ok {
		clipSyncMenu := fyne.NewMenu("ClipSync",
			fyne.NewMenuItem("Start Server", func() {
				fmt.Println("Starting a new clipboard")
				listenPortString := ":"
				if port > 0 {
					listenPortString = ":" + strconv.Itoa(port)
				}
				l, err := net.Listen("tcp4", listenPortString) //nolint // complains about binding to all interfaces
				if err != nil {
					handleError(err)
					return
				}
				defer l.Close()
				port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
				fmt.Println("Server Started at:", getOutboundIP().String()+":"+port)
				fmt.Println()
				for {
					c, err := l.Accept()
					if err != nil {
						handleError(err)
						return
					}
					fmt.Println("Connected to device at " + c.RemoteAddr().String())
					go HandleClient(c)
				}
			}),
			fyne.NewMenuItem("Connect to Server", func() {
				c, err := net.Dial("tcp4", ipAddress)
				if c == nil {
					handleError(err)
					fmt.Println("Could not connect to", ipAddress)
					return
				}
				if err != nil {
					handleError(err)
					return
				}
				defer func() { _ = c.Close() }()
				fmt.Println("Connected to the clipboard")
				go MonitorSentClips(bufio.NewReader(c))
				MonitorLocalClip(bufio.NewWriter(c))
			}),
			fyne.NewMenuItem("Settings", func() {
				clipSyncWindow.Show()
			}),
		)
		deskEnv.SetSystemTrayMenu(clipSyncMenu)
	}

	clipSyncWindow.ShowAndRun()
}

func loadIcon(filePath string) (fyne.Resource, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG image: %v", err)
	}

	return fyne.NewStaticResource(filePath, buf.Bytes()), nil
}

func combineIPAndPort(ip, port string) string {
	return fmt.Sprintf("%s:%s", ip, port)
}

func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address")
	}
	return nil
}

func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

func updateServerInfo(info string) error {
	db, err := sql.Open("sqlite3", "../App/serverinfo.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE server_infos SET ip_address = ? WHERE id = 1", info)
	if err != nil {
		return fmt.Errorf("failed to update server info: %v", err)
	}

	return nil
}

func showErrorDialog(window fyne.Window, message string) {
	dialog.NewError(errors.New(message), window).Show()
}

func showSuccessDialog(window fyne.Window, message string) {
	fullMessage := fmt.Sprintf("%s. \nRestart ClipSync to Apply Changes.", message)
	dialog.NewInformation("Success", fullMessage, window).Show()
}
