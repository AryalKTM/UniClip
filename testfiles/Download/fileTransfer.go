package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	HOST = "localhost"
	PORT = "9001"
	TYPE = "tcp"
)

type MetaData struct {
	name     string
	fileSize uint32
	reps     uint32
}

func prepareMetadata(file *os.File) MetaData {

	fileInfo, err := file.Stat()

	if err != nil {
		log.Fatal(err)
	}

	size := fileInfo.Size()

	header := MetaData{
		name:     file.Name(),
		fileSize: uint32(size),
		reps:     uint32(size / 1014) + 1,
	}

	return header
}

func sendFile(path string, conn *net.TCPConn) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0755)

	if err != nil {
		log.Fatal(err)
	}

	header := prepareMetadata(file)

	dataBuffer := make([]byte, 1014)

	// Start (all 1s) - 1 byte, reps - 4 bytes, lengthofname - 4 bytes, name - `lengthofname` bytes, End (all 0s) - 1 byte;
	headerBuffer := []byte{1}

	// Start (all 0s) - 1 byte, Segment number - 4 bytes, lengthofdata - 4 bytes, Data - `lengthofdata` bytes, End (all 1s) - 1 byte
	segmentBuffer := []byte{0}

	// Temporary buffer for uint32
	temp := make([]byte, 4)

	// Temporary buffer for responses received
	received := make([]byte, 100);

	for i := 0; i < int(header.reps); i++ {
		n, _ := file.ReadAt(dataBuffer, int64(i*1014))

		if i == 0 {
			binary.BigEndian.PutUint32(temp, header.reps)
			headerBuffer = append(headerBuffer, temp...)

			binary.BigEndian.PutUint32(temp, uint32(len(header.name)))
			headerBuffer = append(headerBuffer, temp...)

			headerBuffer = append(headerBuffer, []byte(header.name)...)
			headerBuffer = append(headerBuffer, 0)

			_, err := conn.Write(headerBuffer)

			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.Read(received)

			if err != nil {
				log.Fatal(err)
			}

			println(string(received))
		}

		binary.BigEndian.PutUint32(temp, uint32(i))
		segmentBuffer = append(segmentBuffer, temp...);

		binary.BigEndian.PutUint32(temp, uint32(n))
		segmentBuffer = append(segmentBuffer, temp...)

		segmentBuffer = append(segmentBuffer, dataBuffer...)
		segmentBuffer = append(segmentBuffer, 1)

		_, err = conn.Write(segmentBuffer);

		if err != nil {
			log.Fatal(err);
		}

		_, err = conn.Read(received);
		fmt.Println(string(received));

		if err != nil {
			log.Fatal(err)
		}

		// Reset segment buffer
		segmentBuffer = []byte{0};
	}

}

func main() {
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	if err != nil {
		log.Fatal(err)
	}

	sendFile("Me.png", conn)

	received := make([]byte, 1024)

	_, err = conn.Read(received)

	if err != nil {
		log.Fatal(err)
	}

	println(string(received))
}