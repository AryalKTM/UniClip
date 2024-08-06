package main

import (
	"bufio"

	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"strings"
	"time"
) 

var (
	secondsBetweenChecksForClipChange = 1
	listOfClients                     = make([]*bufio.Writer, 0)
	localClipboard                    string
	printDebugInfo                    = false
)

func getIcon(iconPath string) []byte {
	icon, err := os.ReadFile(iconPath)
	if err != nil {
		log.Fatal(err)
	}
	return icon
}

// monitors for changes to the local clipboard and writes them to w
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
	for {
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
		// hacky way to prevent empty clipboard TODO: find out why empty cb happens
		if foreignClipboard == "" {
			continue
		}
		//foreignClipboard = decompress(foreignClipboardBytes)
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
	}
}

// sendClipboard compresses and then if secure is enabled, encrypts data
func sendClipboard(w *bufio.Writer, clipboard string) error {
	var clipboardBytes []byte
	var err error
	clipboardBytes = []byte(clipboard)
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

// Thanks to https://bruinsslot.jp/post/golang-crypto/ for crypto logic

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

func debug(a ...interface{}) {
	if printDebugInfo {
		fmt.Println("verbose:", a)
	}
}
