package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

func reader(r io.Reader) error {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return err
		}
		println("Client got:", string(buf[0:n]))
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 3 {
		log.Fatal("Please provide a socket to connect to and a filepath relative or absolute")
	}

	sock := os.Args[1]
	filepath := os.Args[2]
	fmt.Println(filepath)

	// Test if file exts from client perspective
	if pkg.FileExists(filepath) {
		println("File exists, ", filepath)
	} else {
		println("File does not exist in, ", filepath)
	}

	// establish connection
	conn, err := net.Dial("unix", sock)
	if err != nil {
		log.Fatal(err)
		log.Fatalf("Failed to connect to the socket: %s", err)
	}
	defer conn.Close()

	fmt.Println("Connected to the Unix socket.")

	// // Setup teardown
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func(con net.Conn) {
	// 	fmt.Println("Closing all connections...")
	// 	conn.Close()
	// }(conn)

	// Write message to the server
	outboundMsg := []byte(filepath)
	_, err = conn.Write(outboundMsg)
	if err != nil {
		log.Fatalf("Failed to write to the socket: %s", err)
	}

	// Wait for the response

	// Read inbound message from the server
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from socket: %s", err)
	}

	inboundMsg := string(buf[:n])
	fmt.Printf("Server Response: %s\n", inboundMsg)
	fmt.Println("shutting down...")
}
