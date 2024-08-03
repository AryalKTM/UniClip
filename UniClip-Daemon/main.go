package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"encoding/base64"

	"golang.org/x/crypto/scrypt"
)

var (
	secondsBetweenChecksForClipChange = 1
	helpMsg                           = `Uniclip - Universal Clipboard
With Uniclip, you can copy from one device and paste on another.

Usage: uniclip [--secure/-s] [--debug/-d] [ <address> | --help/-h ]
Examples:
   uniclip                                   # start a new clipboard
   uniclip -p 6666                           # start a new clipboard on a set port number
   uniclip 192.168.86.24:53701               # join the clipboard at 192.168.86.24:53701
   uniclip -d                                # start a new clipboard with debug output
   uniclip -d --secure 192.168.86.24:53701   # join the clipboard with debug output and enable encryption
Running just ` + "`uniclip`" + ` will start a new clipboard.
It will also provide an address with which you can connect to the same clipboard with another device.
Refer to https://github.com/quackduck/uniclip for more information`
	listOfClients  = make([]*bufio.Writer, 0)
	localClipboard string
	printDebugInfo = false
	version        = "dev"
	cryptoStrength = 16384
	secure         = false
	password       []byte
	port           = 8050
)

// TODO: Add a way to reconnect (if computer goes to sleep)
func main() {
	if len(os.Args) > 4 {
		handleError(errors.New("too many arguments"))
		fmt.Println(helpMsg)
		return
	}
	if hasOption, _ := argsHaveOption("help", "h"); hasOption {
		fmt.Println(helpMsg)
		return
	}
	if hasOption, i := argsHaveOption("debug", "d"); hasOption {
		printDebugInfo = true
		os.Args = removeElemFromSlice(os.Args, i) // delete the debug option and run again
		main()
		return
	}
	// --secure encrypts your data
	if hasOption, i := argsHaveOption("secure", "s"); hasOption {
		secure = true
		os.Args = removeElemFromSlice(os.Args, i) // delete the secure option and run again
		fmt.Print("Password for --secure: ")
		password, _ = terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		main()
		return
	}
	if hasOption, i := argsHaveOption("port", "p"); hasOption {
		os.Args = removeElemFromSlice(os.Args, i) // delete the port option
		if port > 0 {
			fmt.Fprintln(os.Stderr, "Only one port number allowed")
			os.Exit(1)
		}
		if len(os.Args) < i+1 {
			fmt.Fprintln(os.Stderr, "Missing port number")
			os.Exit(1)
		}
		requestedPort, err := strconv.Atoi(os.Args[i])
		if err != nil || requestedPort < 1 || requestedPort > 65534 {
			fmt.Fprintln(os.Stderr, "Invalid port number")
			os.Exit(1)
		}
		os.Args = removeElemFromSlice(os.Args, i)
		port = requestedPort
		main()
		return
	}
	if hasOption, _ := argsHaveOption("version", "v"); hasOption {
		fmt.Println(version)
		return
	}
	if len(os.Args) == 2 {
		ConnectToServer(os.Args[1])
		return
	}
	makeServer()
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

// Handle a client as a server
func HandleClient(c net.Conn) {
	w := bufio.NewWriter(c)
	listOfClients = append(listOfClients, w)
	defer c.Close()
	go MonitorSentClips(bufio.NewReader(c))
	MonitorLocalClip(w)
}

// Connect to the server (which starts a new clipboard)
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
	var isImage bool
	var foreignClipboardBytes []byte

	for {
		// Receive metadata first
		err := gob.NewDecoder(r).Decode(&isImage)
        if err != nil {
            if err == io.EOF {
                return // no need to monitor: disconnected
            }
            handleError(err)
            continue // continue getting next message
        }

		// Receive actual clipboard data
		err = gob.NewDecoder(r).Decode(&foreignClipboardBytes)
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

		foreignClipboard := string(foreignClipboardBytes)

		if isImage {
			// Convert base64 string to image data (platform-dependent)
			// For example, on Linux, you might use `xclip` or `xsel`
			err = setLocalClipAsImage(foreignClipboard)
			if err != nil {
				handleError(err)
				continue
			}
		} else {
			setLocalClip(foreignClipboard)
		}

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

func setLocalClipAsImage(b64Data string) error {
	imageData, err := base64.StdEncoding.DecodeString(b64Data[strings.IndexByte(b64Data, ',')+1:])
	if err != nil {
		return err
	}

	var copyCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		copyCmd = exec.Command("osascript", "-e", `set the clipboard to (read (POSIX file "/dev/stdin") as JPEG picture)`)
	case "windows":
		copyCmd = exec.Command("powershell.exe", "-command", "$img = [System.Drawing.Image]::FromStream([System.IO.MemoryStream]::new([Convert]::FromBase64String((Get-Clipboard))))", "Set-Clipboard -Image $img")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			copyCmd = exec.Command("xclip", "-selection", "clipboard", "-t", "image/png")
		} else if _, err = exec.LookPath("xsel"); err == nil {
			copyCmd = exec.Command("xsel", "--input", "--clipboard")
		} else if _, err = exec.LookPath("wl-copy"); err == nil {
			copyCmd = exec.Command("wl-copy", "--type", "image/png")
		} else if _, err = exec.LookPath("termux-clipboard-set"); err == nil {
			copyCmd = exec.Command("termux-clipboard-set")
		} else {
			return errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux:API installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues")
		}
	}

	in, err := copyCmd.StdinPipe()
	if err != nil {
		return err
	}
	if err = copyCmd.Start(); err != nil {
		return err
	}
	if _, err = in.Write(imageData); err != nil {
		return err
	}
	if err = in.Close(); err != nil {
		return err
	}
	return copyCmd.Wait()
}


// sendClipboard compresses and then if secure is enabled, encrypts data
func sendClipboard(w *bufio.Writer, clipboard string) error {
	var clipboardBytes []byte
	var err error
	isImage := strings.HasPrefix(clipboard, "data:image/")

	clipboardBytes = []byte(clipboard)
	if secure {
		clipboardBytes, err = encrypt(password, clipboardBytes)
		if err != nil {
			return err
		}
	}

	// Send metadata first
	err = gob.NewEncoder(w).Encode(isImage)
	if err != nil {
		return err
	}

	// Send actual clipboard data
	err = gob.NewEncoder(w).Encode(clipboardBytes)
	if err != nil {
		return err
	}

	debug("sent:", clipboard)
	return w.Flush()
}

// Thanks to https://bruinsslot.jp/post/golang-crypto/ for crypto logic
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

func compress(str string) []byte {
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, -1)
	_, _ = zw.Write([]byte(str))
	_ = zw.Close()
	return buf.Bytes()
}

func decompress(b []byte) string {
	var buf bytes.Buffer
	_, _ = buf.Write(b)
	zr := flate.NewReader(&buf)
	decompressed, err := ioutil.ReadAll(zr)
	if err != nil {
		handleError(err)
		return "Issues while decompressing clipboard"
	}
	_ = zr.Close()
	return string(decompressed)
}

func runGetClipCommand() string {
	var out []byte
	var err error
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows":
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

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80") // address can be anything. Doesn't even have to exist
	if err != nil {
		handleError(err)
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func handleError(err error) {
	if err == io.EOF {
		fmt.Println("Disconnected")
	} else {
		fmt.Fprintln(os.Stderr, "error: ["+err.Error()+"]")
	}
}

func debug(a ...interface{}) {
	if printDebugInfo {
		fmt.Println("verbose:", a)
	}
}

func argsHaveOption(long string, short string) (hasOption bool, foundAt int) {
	for i, arg := range os.Args {
		if arg == "--"+long || arg == "-"+short {
			return true, i
		}
	}
	return false, 0
}

// keep order
func removeElemFromSlice(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}