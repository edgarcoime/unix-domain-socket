package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"unicode"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	BYTE_BUFFER = pkg.BYTE_BUFFER
)

type ClientConnectionError struct {
	CC    *ClientConnection
	Error error
}

func NewCCError(cc *ClientConnection, err error) *ClientConnectionError {
	return &ClientConnectionError{
		CC:    cc,
		Error: err,
	}
}

type ClientConnection struct {
	ID   int64
	Conn net.Conn
}

func NewClientConnection(conn net.Conn) *ClientConnection {
	cc := &ClientConnection{
		ID:   pkg.GenerateUniqueID(),
		Conn: conn,
	}

	return cc
}

func (cc *ClientConnection) WriteToClient(s string) error {
	convertedMsg := []byte(s)
	_, err := cc.Conn.Write(convertedMsg)
	if err != nil {
		msg := "ClientConnection.ProcessRequest: Could not write message back to Client."
		return pkg.HandleErrorFormat(msg, err)
	}
	return nil
}

func (cc *ClientConnection) ProcessRequest(leaving chan *ClientConnection, errors chan *ClientConnectionError) {
	// Defer leaving in case return early
	defer func() {
		leaving <- cc
	}()

	reader := bufio.NewReader(cc.Conn)
	// writer := bufio.NewWriter(cc.Conn)

	var sb strings.Builder

	// Process metadata first
	header, err := reader.ReadString('\n')
	if err != nil {
		errors <- NewCCError(cc, pkg.HandleErrorFormat("ClientConnection.ProcessRequest: Could not read packet header for the file", err))
		return
	}

	for {
		// Read until new line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected", cc.Conn.RemoteAddr())
				break
			}
			errors <- NewCCError(cc, pkg.HandleErrorFormat("ClientConnection.ProcessRequest: Error reading from client", err))
			break
		}

		// Trim newline characters and append to the message
		trimmedLine := line[:len(line)-1] // Remove the newline character
		sb.WriteString(trimmedLine + "\n")
	}

	fmt.Println("Header: ", header)
	fmt.Println(sb.String())
	count := 0
	fullMsg := sb.String()
	for _, c := range fullMsg {
		if unicode.IsLetter(c) {
			count++
			fmt.Println(string(c), count)
		}
	}

	fmt.Println(count)
	// // Add new line for seperator
	// resp := fmt.Sprintf("%d\n", count)
	// _, err = writer.WriteString(resp)
	// if err != nil {
	// 	errors <- NewCCError(cc, pkg.HandleErrorFormat("ClientConnection.ProcessRequest: Error writing to client", err))
	// }

	fmt.Println("ClientConnection processing request concluding...")
}

func (cc *ClientConnection) Close() {
	cc.Conn.Close()
	fmt.Println("Closing client connection")
}
