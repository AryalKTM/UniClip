package main

import (
	"io"
	"net"
	
	"os"
	"path/filepath"
	"fmt"
)

func sendFile(filePath string, conn net.Conn) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	// Send file name size and file name
	fileNameSize := int64(len(fileName))
	err = binary.Write(conn, binary.LittleEndian, fileNameSize)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(fileName))
	if err != nil {
		return err
	}

	// Send file size
	err = binary.Write(conn, binary.LittleEndian, fileSize)
	if err != nil {
		return err
	}

	// Send file content
	_, err = io.Copy(conn, file)
	if err != nil {
		return err
	}

	return nil
}

func receiveFile(filePath string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", fileTransferPort))
	if err != nil {
		return err
	}
	defer ln.Close()

	conn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	var fileNameSize int64
	err = binary.Read(conn, binary.LittleEndian, &fileNameSize)
	if err != nil {
		return err
	}

	fileName := make([]byte, fileNameSize)
	_, err = conn.Read(fileName)
	if err != nil {
		return err
	}

	var fileSize int64
	err = binary.Read(conn, binary.LittleEndian, &fileSize)
	if err != nil {
		return err
	}

	downloadPath := filepath.Join(downloadFolder, string(fileName))
	file, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.CopyN(file, conn, fileSize)
	if err != nil {
		return err
	}

	fmt.Printf("Received file: %s\n", downloadPath)
	return nil
}

// Main.go
//package main

var version = "0.3.2"

// func main() {
// 	flag.Usage = func() {
// 		fmt.Printf("Usage of %s:\n", os.Args[0])
// 		fmt.Printf("  -connect <host>:<port>\n")
// 		fmt.Printf("  -port <port>\n")
// 		fmt.Printf("  -file <path>\n")
// 		fmt.Printf("  -v\n")
// 		fmt.Printf("  -version\n")
// 		flag.PrintDefaults()
// 	}
// 	var (
// 		hostToConnectTo = flag.String("connect", "", "Connect to a clipboard at this address")
// 		portToConnectTo = flag.String("port", "", "Port to connect to or host the clipboard")
// 		fileToTransfer  = flag.String("file", "", "Path to the file to transfer")
// 		verbose         = flag.Bool("v", false, "Print verbose (debug) information")
// 		printVersion    = flag.Bool("version", false, "Print the version and exit")
// 	)
// 	flag.Parse()

// 	printDebugInfo = *verbose
// 	if *printVersion {
// 		fmt.Println("UniClip version", version)
// 		os.Exit(0)
// 	}

// 	if *hostToConnectTo != "" && *portToConnectTo != "" {
// 		connectToServer(*hostToConnectTo, *portToConnectTo)
// 	} else if *fileToTransfer != "" {
// 		transferFile(*fileToTransfer)
// 	} else {
// 		makeServer()
// 	}
// }