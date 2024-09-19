package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	fmt.Println("hello world")

	// Open text
	file, err := os.Open("./text.txt")
	if err != nil {
		log.Fatal(err)
	}
	// Handle file close
	defer file.Close()

	// Process text
	// https://www.kelche.co/blog/go/golang-bufio/
	// https://www.scaler.com/topics/golang/golang-read-file/
	var sb strings.Builder
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		sb.WriteString(scanner.Text())
	}
	fmt.Println(sb.String())
}

func readFile(fname string) (string, error) {
	file, err := os.Open("./db/test.txt")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	txt := ""

	// Loop in chunks to get everything in file
	// Scan filem
	for scanner.Scan() {
		// Can print per line here
		txt += scanner.Text()
	}

	// Check if error in scan
	return "", nil
}

func test() {
	// Create a Unix domain socket and listen incoming connections
	socket, err := net.Listen("unix", "/tmp/echo.sock")
	if err != nil {
		log.Fatal(err)
	}

	// cleanup the sockfile
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/echo.sock")
		os.Exit(1)
	}()

	for {
		// Accept an incoming connection
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a seperate goroutine
		go func(conn net.Conn) {
			defer conn.Close()
			buf := make([]byte, 4096)

			// Read data from the connection
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			// echo the data back to the connection
			// Respond back to the client
			_, err = conn.Write(buf[:n])
			// Log in server
			fmt.Println(buf[:n])
			fmt.Printf("Address(%s): %s\n", conn.LocalAddr(), buf[:n])
			if err != nil {
				log.Fatal(err)
			}
		}(conn)
	}

	// get all files get a list of names
	// get this file name
	// quit connection
	// force quit
}
