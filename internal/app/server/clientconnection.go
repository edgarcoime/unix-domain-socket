package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

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
	var sb strings.Builder

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

	fmt.Println(sb.String())

	// Read Client message stream
	// - First packet is meta data colon seperated
	// - The rest will be txt contained in the file
	// Create a reader to read client text
	// Loop through all client requests until completed
	// - Use string builder to be more efficient
	// - build string and do calculations on it
	// Respond back to the user and and connection

	// fmt.Println("ClientConnection processing request...")
	// buf := make([]byte, BYTE_BUFFER)
	//
	// // NEED n so that null bytes are ommitted when converting to string
	// n, err := cc.Conn.Read(buf)
	// if err != nil {
	// 	cc.WriteToClient("Could not read incoming message.")
	// 	msg := "ClientConnection.ProcessRequest: Could not read client message through buffer."
	// 	errors <- NewCCError(cc, pkg.HandleErrorFormat(msg, err))
	// 	return
	// }
	//
	// // Only slices can be converted to string not []byte
	// msg := string(buf[:n])
	//
	// // Echo message back to user
	// err = cc.WriteToClient(msg)
	// if err != nil {
	// 	cc.WriteToClient("Could not write back to respond.")
	// 	errors <- NewCCError(cc, err)
	// 	return
	// }

	fmt.Println("ClientConnection processing request concluding...")
}

func (cc *ClientConnection) Close() {
	cc.Conn.Close()
	fmt.Println("Closing client connection")
}
