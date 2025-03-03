package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/neovim/go-client/nvim"
)

// handleConnection processes incoming requests to open files in Neovim
func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// strip drive
	noDrive := request[2:]

	parts := strings.Split(strings.TrimSpace(noDrive), ":")

	if len(parts) == 3 {
		// add the drive back
		filePath := request[:2] + parts[0]
		line := parts[1]
		column := parts[2]
		openFileInNeovim(filePath, line, column)
	} else {
		fmt.Printf("Invalid request format: %v\n%s\n\n", parts, request)
	}
}

func openFileInNeovim(filePath, line, column string) {
	// Connect to Neovim using the socket
	vim, err := nvim.Dial("127.0.0.1:4293")

	if err != nil {
		log.Fatalf("Failed to connect to Neovim: %v\n", err)
	}

	defer vim.Close()

	// Check if the buffer exists and is displayed in a window
	var bufNr int
	err = vim.Eval(fmt.Sprintf("bufnr('%s')", filePath), &bufNr)
	if err != nil {
		log.Fatalf("Failed to get buffer number: %v\n", err)
	}

	var isLoaded int
	err = vim.Eval(fmt.Sprintf("bufloaded(%d)", bufNr), &isLoaded)
	if err != nil {
		log.Fatalf("Failed to check if buffer is loaded: %v\n", err)
	}

	var winId int
	if bufNr != -1 && isLoaded == 1 {
		err = vim.Eval(fmt.Sprintf("bufwinnr(%d)", bufNr), &winId)
		if err != nil {
			log.Fatalf("Failed to check window: %v\n", err)
		}
	}

	if winId > 0 {
		// File is already open in a window, switch to that window
		err = vim.Command(fmt.Sprintf("%dwincmd w", winId))

		if err != nil {
			log.Fatalf("Failed to switch to window: %v\n", err)
		}
	} else {
		// Open the file in the current window
		err = vim.Command(fmt.Sprintf("edit %s", filePath))

		if err != nil {
			log.Fatalf("Failed to open file: %v\n", err)
		}
	}

	err = vim.Command(fmt.Sprintf("call cursor(%s, %s)", line, column))

	if err != nil {
		log.Fatalf("Failed to move cursor: %v\n", err)
	}
}

func main() {
	ip := "127.0.0.1:14905"
	listener, err := net.Listen("tcp", ip)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		defer conn.Close()

		go handleConnection(conn)
	}
}
