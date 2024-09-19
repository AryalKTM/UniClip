package main

import (
	"bufio"
	"net"
)

// Handle a client as a server
func HandleClient(c net.Conn) {
	w := bufio.NewWriter(c)
	listOfClients = append(listOfClients, w)
	defer c.Close()
	go MonitorSentClips(bufio.NewReader(c))
	MonitorLocalClip(w)
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