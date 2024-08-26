package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/png"
	"log"
	"net"
	"os"
	"strconv"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	_ "github.com/mattn/go-sqlite3"
)

// ShowGUI starts the server info updater GUI.
func renderSettingsMenu() {
	go func() {
		w := app.NewWindow(app.Title("Server Info Updater"), app.Size(unit.Dp(800), unit.Dp(600)))
		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	th := material.NewTheme()
	var ops op.Ops

	ipEntry := new(widget.Editor)
	portEntry := new(widget.Editor)
	updateButton := new(widget.Clickable)
	var errorText string

	for e := range w.Events() {
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			if updateButton.Clicked() {
				ip := ipEntry.Text()
				port := portEntry.Text()

				if err := validateIP(ip); err != nil {
					errorText = "Error: " + err.Error()
				} else if err := validatePort(port); err != nil {
					errorText = "Error: " + err.Error()
				} else {
					combinedInfo := combineIPAndPort(ip, port)
					err := updateServerInfo(combinedInfo)
					if err != nil {
						errorText = "Error: " + err.Error()
					} else {
						errorText = "Server info updated successfully."
					}
				}
			}

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEvenly,
			}.Layout(gtx,
				layout.Rigid(material.H6(th, "Enter new IP address:").Layout),
				layout.Rigid(material.Editor(th, ipEntry, "Enter IP Address").Layout),
				layout.Rigid(material.H6(th, "Enter new Port:").Layout),
				layout.Rigid(material.Editor(th, portEntry, "Enter Port").Layout),
				layout.Rigid(material.Button(th, updateButton, "Update Server Info").Layout),
				layout.Rigid(material.Label(th, unit.Sp(16), errorText).Layout),
			)

			e.Frame(gtx.Ops)
		}
	}
}

// Function to combine IP address and port into a single string
func combineIPAndPort(ip, port string) string {
	return fmt.Sprintf("%s:%s", ip, port)
}

// Function to validate the IP address
func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address")
	}
	return nil
}

// Function to validate the port number
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

// Function to update the combined IP address and port in the SQLite database
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

// Function to load and return the icon as a fyne.Resource
func loadIcon(filePath string) ([]byte, error) {
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

	return buf.Bytes(), nil
}
