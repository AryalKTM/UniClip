package main

import (
    "bytes"
    "database/sql"
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
    "fyne.io/fyne/v2/widget"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Server Info Updater")

    // Load and set the icon image
    iconResource, err := loadIcon("../icon.png")
    if err != nil {
        fmt.Println("Error loading icon:", err)
        return
    }
    myWindow.SetIcon(iconResource)

    // Create text fields for IP address and port
    ipEntry := widget.NewEntry()
    ipEntry.SetPlaceHolder("Enter IP Address")
    
    portEntry := widget.NewEntry()
    portEntry.SetPlaceHolder("Enter Port")

    // Create a button to update the IP address and port in the database
    updateButton := widget.NewButton("Update Server Info", func() {
        ip := ipEntry.Text
        port := portEntry.Text

        if err := validateIP(ip); err != nil {
            showErrorDialog(myWindow, err.Error())
            return
        }

        if err := validatePort(port); err != nil {
            showErrorDialog(myWindow, err.Error())
            return
        }

        combinedInfo := combineIPAndPort(ip, port)
        err := updateServerInfo(combinedInfo)
        if err != nil {
            showErrorDialog(myWindow, err.Error())
        } else {
            showSuccessDialog(myWindow, "Server info updated successfully.")
        }
    })

    // Create a container to hold the entries and button
    content := container.NewVBox(
        widget.NewLabel("Enter new IP address:"),
        ipEntry,
        widget.NewLabel("Enter new Port:"),
        portEntry,
        updateButton,
    )

    myWindow.SetContent(content)

    // Resize the window to 800x600 pixels
    myWindow.Resize(fyne.NewSize(800, 600))

    myWindow.ShowAndRun()
}

// Function to load and return the icon as a fyne.Resource
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

// Function to combine IP address and port into a single string
func combineIPAndPort(ip, port string) string {
    return fmt.Sprintf("%s:%s", ip, port)
}

// Function to validate the IP address
func validateIP(ip string) error {
    // Validate if the IP address is valid
    if net.ParseIP(ip) == nil {
        return fmt.Errorf("invalid IP address")
    }
    return nil
}

// Function to validate the port number
func validatePort(port string) error {
    // Check if the port is a valid number
    p, err := strconv.Atoi(port)
    if err != nil {
        return fmt.Errorf("port must be a number")
    }
    if p < 1 || p > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
}

// Function to update the combined IP address and port in the SQLite database
func updateServerInfo(info string) error {
    // Open the SQLite database
    db, err := sql.Open("sqlite3", "../App/serverinfo.db")
    if err != nil {
        return fmt.Errorf("failed to open database: %v", err)
    }
    defer db.Close()

    // Update the combined IP address and port where id = 1
    _, err = db.Exec("UPDATE server_infos SET ip_address = ? WHERE id = 1", info)
    if err != nil {
        return fmt.Errorf("failed to update server info: %v", err)
    }

    return nil
}

// Function to show an error dialog
func showErrorDialog(window fyne.Window, message string) {
    dialog.NewError(fmt.Errorf(message), window).Show()
}

// Function to show a success dialog
func showSuccessDialog(window fyne.Window, message string) {
    dialog.NewInformation("Success", message, window).Show()
}