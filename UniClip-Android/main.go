package main

import (
	"bufio"
	"bytes"
	// "compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"golang.org/x/term"
	"io"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/crypto/scrypt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	secondsBetweenChecksForClipChange = 1
	listOfClients  = make([]*bufio.Writer, 0)
	localClipboard string
	printDebugInfo = false
	cryptoStrength = 16384
	secure         = false
	password       []byte
	port           = 0
	clip           fyne.Clipboard
)

// TODO: Add a way to reconnect (if computer goes to sleep)
func main() {
	a := app.New()
	w := a.NewWindow("Uniclip")
	clip = w.Clipboard()

	w.SetContent(container.NewVBox(
		widget.NewLabel("Uniclip Clipboard"),
		widget.NewButton("Start Clipboard", func() {
			go makeServer()
		}),
		widget.NewButton("Connect to Clipboard", func() {
			address := widget.NewEntry()
			dialog.ShowForm("Enter address to connect:",
				"Connect", "Cancel",
				[]*widget.FormItem{
					widget.NewFormItem("Address", address),
				},
				func(b bool) {
					if b {
						go ConnectToServer(address.Text)
					}
				},
				w)
		}),
	))

	w.Resize(fyne.NewSize(300, 200))
	w.ShowAndRun()
}

func makeServer() {
	fmt.Println("Starting a new clipboard")
	listenPortString := ":"
	if port > 0 {
		listenPortString = ":" + strconv.Itoa(port)
	}
	l, err := net.Listen("tcp4", listenPortString)
	if err != nil {
		handleError(err)
		return
	}
	defer l.Close()
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	fmt.Println("Run", "`uniclip", getOutboundIP().String()+":"+port+"`", "to join this clipboard")
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
}

func HandleClient(c net.Conn) {
	w := bufio.NewWriter(c)
	listOfClients = append(listOfClients, w)
	defer c.Close()
	go MonitorSentClips(bufio.NewReader(c))
	MonitorLocalClip(w)
}

func ConnectToServer(address string) {
	c, err := net.Dial("tcp4", address)
	if c == nil {
		handleError(err)
		fmt.Println("Could not connect to", address)
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
}

func MonitorLocalClip(w *bufio.Writer) {
	for {
		localClipboard = getLocalClip()
		fmt.Println("Local clipboard content:", localClipboard) // Debug statement
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

func MonitorSentClips(r *bufio.Reader) {
	var foreignClipboard string
	var foreignClipboardBytes []byte
	for {
		err := gob.NewDecoder(r).Decode(&foreignClipboardBytes)
		if err != nil {
			if err == io.EOF {
				return
			}
			handleError(err)
			continue
		}

		if secure {
			foreignClipboardBytes, err = decrypt(password, foreignClipboardBytes)
			if err != nil {
				handleError(err)
				continue
			}
		}

		foreignClipboard = string(foreignClipboardBytes)
		fmt.Println("Received clipboard content:", foreignClipboard) // Debug statement
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
	}
}

func sendClipboard(w *bufio.Writer, clipboard string) error {
	var clipboardBytes []byte
	var err error
	clipboardBytes = []byte(clipboard)
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
	return w.Flush()
}

func encrypt(key, data []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, err
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	ciphertext = append(ciphertext, salt...)
	return ciphertext, nil
}

func decrypt(key, data []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]
	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, err
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}
	key, err := scrypt.Key(password, salt, cryptoStrength, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}
	return key, salt, nil
}

func getLocalClip() string {
	return clip.Content()
}

func setLocalClip(s string) {
	clip.SetContent(s)
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		handleError(err)
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func debug(a ...interface{}) {
	if printDebugInfo {
		fmt.Println(a...)
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func PromptPassword(confirmationRequired bool) {
	fmt.Print("Enter a password to secure your clipboard: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		handleError(err)
		os.Exit(1)
	}
	fmt.Println()
	if confirmationRequired {
		fmt.Print("Enter the same password again: ")
		confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			handleError(err)
			os.Exit(1)
		}
		fmt.Println()
		if !bytes.Equal(bytePassword, confirmPassword) {
			fmt.Println("Passwords do not match!")
			os.Exit(1)
		}
	}
	password = bytePassword
}

// func getRemoteClip(r *bufio.Reader) (string, error) {
// 	var b bytes.Buffer
// 	r.ReadLine()
// 	io.Copy(&b, r)
// 	d := flate.NewReader(&b)
// 	defer d.Close()
// 	out, err := io.ReadAll(d)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(out), nil
// }
