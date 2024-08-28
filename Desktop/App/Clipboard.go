package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// MonitorLocalClip monitors the local clipboard for changes
func MonitorLocalClip(w *bufio.Writer) {
	for {
		localClipboard = getLocalClip()
		err := sendClipboard(w, localClipboard)
		if err != nil {
			handleError(err)
			return
		}
		for localClipboard == getLocalClip() {
			time.Sleep(time.Second * time.Duration(secondsBetweenChecksForClipChange))
		}
	}
}

// MonitorSentClips monitors the received clips or files from the server
func MonitorSentClips(r *bufio.Reader) {
	var foreignClipboard string
	var foreignClipboardBytes []byte
	var downloadPath string = ""

	for {
		streamType, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return
			}
			handleError(err)
			continue
		}

		switch streamType {
		case 0x00:
			// Clipboard text received
			err := gob.NewDecoder(r).Decode(&foreignClipboardBytes)
			if err != nil {
				if err == io.EOF {
					return
				}
				handleError(err)
				continue
			}

			// Decrypt if needed
			if secure {
				foreignClipboardBytes, err = decrypt(password, foreignClipboardBytes)
				if err != nil {
					handleError(err)
					continue
				}
			}

			foreignClipboard = string(foreignClipboardBytes)
			if foreignClipboard == "" {
				continue
			}
			setLocalClip(foreignClipboard)
			localClipboard = foreignClipboard
			debug("Received:", foreignClipboard)
			for i := range listOfClients {
				if listOfClients[i] != nil {
					err = sendClipboard(listOfClients[i], foreignClipboard)
					if err != nil {
						listOfClients[i] = nil
						fmt.Println("Error sending clipboard to a device. Will not contact that device again.")
					}
				}
			}

		case 0x01:
			// File received
			var fileNameBytes []byte
			err := gob.NewDecoder(r).Decode(&fileNameBytes)
			if err != nil {
				if err == io.EOF {
					return
				}
				handleError(err)
				continue
			}
			fileName := filepath.Base(string(fileNameBytes))

			var fileContent []byte
			err = gob.NewDecoder(r).Decode(&fileContent)
			if err != nil {
				if err == io.EOF {
					return
				}
				handleError(err)
				continue
			}

			if secure {
				fileContent, err = decrypt(password, fileContent)
				if err != nil {
					handleError(err)
					continue
				}
			}

			// Determine the download path based on the OS
			switch runtime.GOOS {
			case "windows":
				downloadPath = filepath.Join(os.Getenv("USERPROFILE"), "Downloads", fileName)
			case "darwin":
				downloadPath = filepath.Join(os.Getenv("HOME"), "Downloads", fileName)
			default:
				downloadPath = filepath.Join(os.Getenv("HOME"), "Downloads", fileName)
			}

			err = os.WriteFile(downloadPath, fileContent, 0644)
			if err != nil {
				fmt.Println("Error writing file")
				handleError(err)
				continue
			}
			fmt.Printf("File created successfully: %s\n", downloadPath)

		default:
			handleError(errors.New("unknown stream type"))
			continue
		}
	}
}

// sendClipboard compresses, encrypts, and sends clipboard or file content
func sendClipboard(w *bufio.Writer, clipboard string) error {
	var clipboardBytes []byte
	var err error
	var streamType byte

	trimmedPath := strings.Trim(clipboard, `"`)
	if isValidFilePath(trimmedPath) {
		// Send file
		file, err := os.Open(trimmedPath)
		if err != nil {
			return err
		}
		defer file.Close()

		fileData, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		streamType = 0x01
		err = w.WriteByte(streamType)
		if err != nil {
			return err
		}

		fileName := filepath.Base(trimmedPath)
		err = gob.NewEncoder(w).Encode([]byte(fileName))
		if err != nil {
			return err
		}

		clipboardBytes = fileData
	} else {
		// Send clipboard text
		streamType = 0x00
		err = w.WriteByte(streamType)
		if err != nil {
			return err
		}

		clipboardBytes = []byte(clipboard)
	}

	if secure {
		clipboardBytes, err = encrypt(password, clipboardBytes)
		if err != nil {
			return err
		}
	}

	err = gob.NewEncoder(w).Encode(clipboardBytes)
	if err != nil {
		return err
	}
	return w.Flush()
}

// getLocalClip gets the current content of the local clipboard
func getLocalClip() string {
	return runGetClipCommand()
}

// setLocalClip sets the clipboard content to the provided string
func setLocalClip(s string) {
	var copyCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		copyCmd = exec.Command("pbcopy")
	case "windows":
		copyCmd = exec.Command("clip")
	default:
		copyCmd = getLinuxCopyCmd()
		if copyCmd == nil {
			handleError(errors.New("unable to find suitable clipboard command"))
			os.Exit(2)
		}
	}

	in, err := copyCmd.StdinPipe()
	if err != nil {
		handleError(err)
		return
	}
	defer in.Close()

	if err := copyCmd.Start(); err != nil {
		handleError(err)
		return
	}

	if _, err := in.Write([]byte(s)); err != nil {
		handleError(err)
		return
	}

	if err := copyCmd.Wait(); err != nil {
		handleError(err)
		return
	}
}

// getLinuxCopyCmd returns the appropriate command to set the clipboard in Linux
func getLinuxCopyCmd() *exec.Cmd {
	if _, err := exec.LookPath("xclip"); err == nil {
		return exec.Command("xclip", "-in", "-selection", "clipboard")
	} else if _, err := exec.LookPath("xsel"); err == nil {
		return exec.Command("xsel", "--input", "--clipboard")
	} else if _, err := exec.LookPath("wl-copy"); err == nil {
		return exec.Command("wl-copy")
	} else if _, err := exec.LookPath("termux-clipboard-set"); err == nil {
		return exec.Command("termux-clipboard-set")
	}
	return nil
}

// runGetClipCommand runs the command to get the current clipboard content
func runGetClipCommand() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows":
		cmd = exec.Command("powershell.exe", "-command", "Get-Clipboard")
	default:
		cmd = getLinuxPasteCmd()
		if cmd == nil {
			handleError(errors.New("unable to find suitable clipboard command"))
			os.Exit(2)
		}
	}

	out, err := cmd.Output()
	if err != nil {
		handleError(err)
		return "Error getting local clipboard"
	}

	if runtime.GOOS == "windows" {
		return strings.TrimSuffix(string(out), "\r\n")
	}
	return string(out)
}

// getLinuxPasteCmd returns the appropriate command to get the clipboard content in Linux
func getLinuxPasteCmd() *exec.Cmd {
	if _, err := exec.LookPath("xclip"); err == nil {
		return exec.Command("xclip", "-out", "-selection", "clipboard")
	} else if _, err := exec.LookPath("xsel"); err == nil {
		return exec.Command("xsel", "--output", "--clipboard")
	} else if _, err := exec.LookPath("wl-paste"); err == nil {
		return exec.Command("wl-paste", "--no-newline")
	} else if _, err := exec.LookPath("termux-clipboard-get"); err == nil {
		return exec.Command("termux-clipboard-get")
	}
	return nil
}