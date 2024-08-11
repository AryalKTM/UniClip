package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
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
It will also provide an address with which you can connect to the same clipboard with another device.`
	listOfClients  = make([]*bufio.Writer, 0)
	localClipboard string
	printDebugInfo = false
	version        = "dev"
	cryptoStrength = 16384
	secure         = false
	password       []byte
	port           = 0
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
		os.Args = removeElemFromSlice(os.Args, i) // delete the port argument and run again
		port = requestedPort
		main()
		return
	}
	if hasOption, _ := argsHaveOption("version", "v"); hasOption {
		fmt.Println(version)
		return
	}
	if len(os.Args) == 2 { // has exactly one argument
		ConnectToServer(os.Args[1])
		return
	}
	makeServer()
}



func argsHaveOption(long string, short string) (hasOption bool, foundAt int) {
	for i, arg := range os.Args {
		if arg == "--"+long || arg == "-"+short {
			return true, i
		}
	}
	return false, 0
}

