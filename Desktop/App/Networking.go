package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

func makeServer() {
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
	fmt.Println("Run", "`clipsync", getOutboundIP().String()+":"+port+"`", "to join this clipboard")
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

func getOutboundIP() net.IP {
	// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go/37382208#37382208
	conn, err := net.Dial("udp", "8.8.8.8:80") // address can be anything. Doesn't even have to exist
	if err != nil {
		handleError(err)
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
