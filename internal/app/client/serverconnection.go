package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	CONN_TYPE = pkg.SERVER_TYPE
)

type ServerConnection struct {
	Conn     *net.TCPConn
	Filepath string
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	wg       sync.WaitGroup
}

func NewServerConnection(c *net.TCPConn, fp string) *ServerConnection {
	reader := bufio.NewReader(c)
	writer := bufio.NewWriter(c)

	sc := &ServerConnection{
		Conn:     c,
		Filepath: fp,
		Reader:   reader,
		Writer:   writer,
		// Don't need to add wg
	}

	// REMEMBER IF THERE IS AN OS EXIT YOU MUST SET THIS UP
	os_c := make(chan os.Signal, 1)
	signal.Notify(os_c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-os_c
		fmt.Println("Sig call shutdown")
		fmt.Println("Os signal: ", s)
		sc.Close()
		os.Exit(1)
	}()

	return sc
}

func (sc *ServerConnection) Close() {
	fmt.Println("Shutting down client...")
	sc.Conn.Close()
}

func (sc *ServerConnection) Start() {
	defer sc.Close()

	sc.wg.Add(2)
	go sc.sendFile()
	go sc.receiveMsg()
	sc.wg.Wait()
}

func (sc *ServerConnection) sendFile() {
	defer sc.wg.Done() // Mark work as done
	// Open file
	file, err := os.Open(sc.Filepath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	fileReader := bufio.NewReader(file)

	// Create first packet
	_, err = sc.Writer.WriteString(sc.Filepath + "\n")
	if err != nil {
		log.Fatalf("Could not send header for the packets to the server\n")
	}

	// Loop through chunks of the file and send chunks to the server
	for {
		line, err := fileReader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				if len(line) > 0 {
					// send last line if doesn't work with new line
					_, writeErr := sc.Writer.WriteString(line)
					if writeErr != nil {
						log.Printf("Error sending last line: %v\n", writeErr)
						break
					}
				}
				break
			}
			log.Fatalf("Error reading the file: %s\n", err)
		}

		// Send the line to the server
		_, err = sc.Writer.WriteString(line)
		if err != nil {
			log.Fatalf("Error sending data: %v\n", err)
		}

		// Flush the buffered data to ensure it's sent immediately
		err = sc.Writer.Flush()
		if err != nil {
			log.Fatalf("Error flushing data: %v\n", err)
		}
	}

	// Signals to the server that there is no more incoming data
	sc.Conn.CloseWrite()
}

func (sc *ServerConnection) receiveMsg() {
	defer sc.wg.Done()

	for {
		message, err := sc.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Server closed the connection.")
			} else {
				log.Printf("Error reading from server: %v", err)
			}

			os.Exit(0)
		}

		// Received final msg from server
		fmt.Printf("Message received from server: %s\n", message)
		break
	}
}
