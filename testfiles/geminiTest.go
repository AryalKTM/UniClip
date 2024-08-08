package main

import (
    "fmt"
    "io"
    "net"
    "os"
    "time"

    "github.com/atotto/clipboard"
    // "github.com/skratchdot/open-golang/open"
)

const (
    port          = "8080" // Choose a suitable port
    pollInterval  = 1 * time.Second
    network       = "tcp"
)

func handleClipboardChanges(conn net.Conn) {
    for {
        text, err := clipboard.ReadAll()
        if err != nil {
            fmt.Println("Error reading clipboard:", err)
            continue
        }

        // If text is a file path
        if _, err := os.Stat(text); err == nil {
            fmt.Println("Sending file:", text)

            // Send file data to connected PC
            file, _ := os.Open(text)
            defer file.Close()

            _, err = io.Copy(conn, file)
            if err != nil {
                fmt.Println("Error sending file:", err)
            }
        } else {
            fmt.Println("Sharing clipboard text:", text)
            // Send text to connected PC
            _, err = conn.Write([]byte(text))
            if err != nil {
                fmt.Println("Error sending text:", err)
            }
        }

        time.Sleep(pollInterval)
    }
}

func handleIncomingFile(conn net.Conn) {
    // Read and save incoming files to a temporary directory
}

func main() {
    // Establish connection (either as server or client)
    listener, err := net.Listen(network, ":"+port)
    if err != nil {
        // If listening fails, try connecting as a client
        conn, err := net.Dial(network, ":"+port)
        if err != nil {
            fmt.Println("Error connecting as client:", err)
            return
        }
        defer conn.Close()
        go handleClipboardChanges(conn)
        handleIncomingFile(conn)

    } else {
        defer listener.Close()
        fmt.Println("Waiting for connections...")

        for {
            conn, err := listener.Accept()
            if err != nil {
                fmt.Println("Error accepting connection:", err)
                continue
            }
            fmt.Println("Connection established:", conn.RemoteAddr())
            defer conn.Close()

            go handleClipboardChanges(conn)
            go handleIncomingFile(conn)
        }
    }
}
