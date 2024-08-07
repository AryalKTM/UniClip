package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

const (
	uploadDir                         = "./Download"
	secondsBetweenChecksForClipChange = 1
	serverPort                        = ":8080"
	tcpPort                           = ":8081"
)

// Function to handle file downloads
func downloadFile(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "File path is missing", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Unable to send file", http.StatusInternalServerError)
		return
	}
}

// Function to handle file uploads
func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	os.MkdirAll(uploadDir, os.ModePerm)
	dest, err := os.Create(filepath.Join(uploadDir, handler.Filename))
	if err != nil {
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}

// Function to upload a file to a server
func uploadToServer(serverURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	resp, err := http.Post(fmt.Sprintf("%s/upload", serverURL), writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(respBody))
	return nil
}

// Function to send clipboard text to the connected device
func sendClipboard(conn net.Conn, text string) error {
	writer := bufio.NewWriter(conn)
	_, err := writer.WriteString(text + "\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	return err
}

// Function to monitor local clipboard and send changes
func MonitorLocalClip(conn net.Conn, serverURL string) {
	var localClipboard string
	for {
		newClipboard, err := clipboard.ReadAll()
		if err != nil {
			handleError(err)
			continue
		}
		if newClipboard != localClipboard {
			localClipboard = newClipboard
			isFilePath := strings.HasPrefix(localClipboard, `"`) && strings.HasSuffix(localClipboard, `"`)
			if isFilePath {
				trimmedPath := strings.Trim(localClipboard, `"`)
				err := sendClipboard(conn, localClipboard)
				if err != nil {
					handleError(err)
					continue
				}
				err = uploadToServer(serverURL, trimmedPath)
				if err != nil {
					handleError(err)
					continue
				}
			} else {
				err := sendClipboard(conn, localClipboard)
				if err != nil {
					handleError(err)
					continue
				}
			}
		}
		time.Sleep(time.Second * time.Duration(secondsBetweenChecksForClipChange))
	}
}

// Function to handle received clipboard content
func handleReceivedClipboard(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		receivedClipboard, err := reader.ReadString('\n')
		if err != nil {
			handleError(err)
			continue
		}
		receivedClipboard = strings.TrimSpace(receivedClipboard)
		isFilePath := strings.HasPrefix(receivedClipboard, `"`) && strings.HasSuffix(receivedClipboard, `"`)
		if isFilePath {
			trimmedPath := strings.Trim(receivedClipboard, `"`)
			fmt.Println("Received file path:", trimmedPath)
			// Handle file transfer here
		} else {
			err := clipboard.WriteAll(receivedClipboard)
			if err != nil {
				handleError(err)
				continue
			}
			fmt.Println("Received clipboard text:", receivedClipboard)
		}
	}
}

func handleError(err error) {
	fmt.Println("Error:", err)
}

func main() {
	http.HandleFunc("/download", downloadFile)
	http.HandleFunc("/upload", uploadFile)

	go func() {
		fmt.Println("Server started at http://localhost:8080")
		err := http.ListenAndServe(serverPort, nil)
		if err != nil {
			handleError(err)
		}
	}()

	listener, err := net.Listen("tcp", tcpPort)
	if err != nil {
		handleError(fmt.Errorf("error starting TCP server: %w", err))
		return
	}
	defer listener.Close()
	fmt.Println("TCP server started on port 8081")

	connections := make([]net.Conn, 0)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				handleError(fmt.Errorf("error accepting connection: %w", err))
				continue
			}
			connections = append(connections, conn)
			go handleReceivedClipboard(conn)
		}
	}()

	fmt.Print("Enter the IP of the server to connect (or press Enter to run as server only): ")
	var serverIP string
	fmt.Scanln(&serverIP)

	if serverIP != "" {
		conn, err := net.Dial("tcp", serverIP+tcpPort)
		if err != nil {
			handleError(fmt.Errorf("error connecting to server: %w", err))
			return
		}
		connections = append(connections, conn)
		go handleReceivedClipboard(conn)
		serverURL := fmt.Sprintf("http://%s%s", serverIP, serverPort)
		go MonitorLocalClip(conn, serverURL)
	}

	// Monitor local clipboard changes and send them to all connected devices
	for _, conn := range connections {
		go MonitorLocalClip(conn, "http://localhost"+serverPort)
	}

	select {} // Keep the main function running
}
