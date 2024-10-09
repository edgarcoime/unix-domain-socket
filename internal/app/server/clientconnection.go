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
	ID         int64
	Conn       net.Conn
	incoming   chan string
	outgoing   chan string
	disconnect chan bool
	server     *NetworkSocketServer
}

func NewClientConnection(conn net.Conn, s *NetworkSocketServer) *ClientConnection {
	cc := &ClientConnection{
		ID:         pkg.GenerateUniqueID(),
		Conn:       conn,
		incoming:   make(chan string),
		outgoing:   make(chan string),
		disconnect: make(chan bool),
		server:     s,
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

func (cc *ClientConnection) Close() {
	if cc.Conn != nil {
		cc.Conn.Close()
	}
	// close(cc.incoming)
	// close(cc.outgoing)
	// close(cc.disconnect)
	cc.server.leaving <- cc
	fmt.Println("Closing client connection")
}

func (cc *ClientConnection) Start() {
	defer cc.Close()
	go cc.processFile()
	go cc.writeResponse()
	<-cc.disconnect // Block until disconnect channel
}

func (cc *ClientConnection) processFile() {
	defer fmt.Println("Closing processFile")
	errors := cc.server.clientErrors

	reader := bufio.NewReader(cc.Conn)
	var sb strings.Builder
	// Process metadata first
	header, err := reader.ReadString('\n')
	if err != nil {
		errors <- NewCCError(cc, pkg.HandleErrorFormat("ClientConnection.ProcessRequest: Could not read packet header for the file", err))
		cc.disconnect <- true
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
			return
		}

		// Trim newline characters and append to the message
		trimmedLine := line[:len(line)-1] // Remove the newline character
		sb.WriteString(trimmedLine + "\n")
	}

	// Delay count
	// time.Sleep(5 * time.Second)

	// Count alphas
	count := 0
	fullMsg := sb.String()
	for _, c := range fullMsg {
		if unicode.IsLetter(c) {
			count++
		}
	}

	// Send response to outgoing channel for processing
	header = strings.TrimSpace(header)
	fmt.Printf("File: %s\nCount: %d\n", header, count)
	resp := fmt.Sprintf("%d\n", count)
	cc.outgoing <- resp
}

func (cc *ClientConnection) writeResponse() {
	// Keep looping/opening connection until client disconnects
	defer fmt.Println("Closing writeResponse")
	errors := cc.server.clientErrors
	writer := bufio.NewWriter(cc.Conn)

	for resp := range cc.outgoing {
		_, err := writer.WriteString(resp)
		if err != nil {
			errors <- NewCCError(cc, pkg.HandleErrorFormat("ClientConnection.ProcessRequest: Error writing to client", err))
			break
		}

		// Flush after writing the response
		err = writer.Flush()
		if err != nil {
			errors <- NewCCError(cc, pkg.HandleErrorFormat("ClientConnection.ProcessRequest: Error flushing data to client", err))
			break
		}

		// Finished communication
		break
	}
	cc.disconnect <- true
}
