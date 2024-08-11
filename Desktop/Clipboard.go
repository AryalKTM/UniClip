package main

import (
	"bufio"
	"time"
	"io"
	"encoding/gob"
	"path/filepath"
	"fmt"
	"runtime"
	"os"
	"errors"
	"strings"
	"io/ioutil"
	"os/exec"
)

func MonitorLocalClip(w *bufio.Writer) {
	for {
		localClipboard = getLocalClip()
		//debug("clipboard changed so sending it. localClipboard =", localClipboard)
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

// monitors for clipboards sent through r
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
			err := gob.NewDecoder(r).Decode(&foreignClipboardBytes)
			if err != nil {
				if err == io.EOF {
					return // no need to monitor: disconnected
				}
				handleError(err)
				continue // continue getting next message
			}

			// decrypt if needed
			if secure {
				foreignClipboardBytes, err = decrypt(password, foreignClipboardBytes)
				if err != nil {
					handleError(err)
					continue
				}
			}

			foreignClipboard = string(foreignClipboardBytes)
			//TO Prevent Empty Clipboard, Don't know why that happens.
			if foreignClipboard == "" {
				continue
			}
			setLocalClip(foreignClipboard)
			localClipboard = foreignClipboard
			debug("rcvd:", foreignClipboard)
			for i := range listOfClients {
				if listOfClients[i] != nil {
					err = sendClipboard(listOfClients[i], foreignClipboard)
					if err != nil {
						listOfClients[i] = nil
						fmt.Println("Error when trying to send the clipboard to a device. Will not contact that device again.")
					}
				}
			}
		case 0x01:
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

			if runtime.GOOS == "windows" {
				downloadPath = filepath.Join(`C:\\Users\aryal\Downloads`, fileName)
				fmt.Println("File Created at:" + downloadPath)
			} else if runtime.GOOS == "darwin" {
				downloadPath = filepath.Join("/Users/aryal/Downloads", fileName)
			}

			err = os.WriteFile(downloadPath, fileContent, 0644)
			if err != nil {
				fmt.Println("Error while Writing File")
				handleError(err)
				continue
			}
			fmt.Print("File Created Successfully\n", fileName)
		default:
			handleError(errors.New("unknown stream type"))
			continue
		}

		//foreignClipboard = decompress(foreignClipboardBytes)
	}
}

// sendClipboard compresses and then if secure is enabled, encrypts data
func sendClipboard(w *bufio.Writer, clipboard string) error {
	var clipboardBytes []byte
	var err error
	var streamType byte
	trimmedPath := strings.Trim(clipboard, `"`)
	fmt.Println("Received:" + clipboard)
	if isValidFilePath(trimmedPath) {
		file, err := os.OpenFile(trimmedPath, os.O_RDONLY, 0755)
		if err != nil {
			return err
		}
		defer file.Close()
		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		streamType = 0x01
		err = w.WriteByte(streamType)
		if err != nil {
			return err
		}

		fileName := []byte(trimmedPath)
		err = gob.NewEncoder(w).Encode(fileName)
		if err != nil {
			return err
		}
		clipboardBytes = fileData
	} else {
		streamType = 0x00
		err = w.WriteByte(streamType)
		if err != nil {
			return err
		}
		clipboardBytes = []byte(clipboard)
	}
	//clipboardBytes = compress(clipboard)
	//fmt.Printf("cmpr: %x\ndcmp: %x\nstr: %s\n\ncmpr better by %d\n", clipboardBytes, []byte(clipboard), clipboard, len(clipboardBytes)-len(clipboard))
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
	debug("sent:", clipboard)
	//if secure {
	//	debug("--secure is enabled, so actually sent as:", hex.EncodeToString(clipboardBytes))
	//}
	return w.Flush()
}

func getLocalClip() string {
	str := runGetClipCommand()
	//for ; str == ""; str = runGetClipCommand() { // wait until it's not empty
	//	time.Sleep(time.Millisecond * 100)
	//}
	return str
}

func setLocalClip(s string) {
	var copyCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		copyCmd = exec.Command("pbcopy")
	case "windows":
		copyCmd = exec.Command("clip")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			copyCmd = exec.Command("xclip", "-in", "-selection", "clipboard")
		} else if _, err = exec.LookPath("xsel"); err == nil {
			copyCmd = exec.Command("xsel", "--input", "--clipboard")
		} else if _, err = exec.LookPath("wl-copy"); err == nil {
			copyCmd = exec.Command("wl-copy")
		} else if _, err = exec.LookPath("termux-clipboard-set"); err == nil {
			copyCmd = exec.Command("termux-clipboard-set")
		} else {
			handleError(errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux:API installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues"))
			os.Exit(2)
		}
	}
	in, err := copyCmd.StdinPipe()
	if err != nil {
		handleError(err)
		return
	}
	if err = copyCmd.Start(); err != nil {
		handleError(err)
		return
	}
	if _, err = in.Write([]byte(s)); err != nil {
		handleError(err)
		return
	}
	if err = in.Close(); err != nil {
		handleError(err)
		return
	}
	if err = copyCmd.Wait(); err != nil {
		handleError(err)
		return
	}
}

func runGetClipCommand() string {
	var out []byte
	var err error
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows": //nolint // complains about literal string "windows" being used multiple times
		cmd = exec.Command("powershell.exe", "-command", "Get-Clipboard")
	default:
		if _, err = exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-out", "-selection", "clipboard")
		} else if _, err = exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--output", "--clipboard")
		} else if _, err = exec.LookPath("wl-paste"); err == nil {
			cmd = exec.Command("wl-paste", "--no-newline")
		} else if _, err = exec.LookPath("termux-clipboard-get"); err == nil {
			cmd = exec.Command("termux-clipboard-get")
		} else {
			handleError(errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues"))
			os.Exit(2)
		}
	}
	if out, err = cmd.Output(); err != nil {
		handleError(err)
		return "An error occurred wile getting the local clipboard"
	}
	if runtime.GOOS == "windows" {
		return strings.TrimSuffix(string(out), "\r\n") // powershell's get-clipboard adds a windows newline to the end for some reason
	}
	return string(out)
}