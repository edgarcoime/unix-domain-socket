package server

import (
	"fmt"
	"net"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

type ClientConnection struct {
	ID     int64
	Conn   net.Conn
	Buffer []byte
}

func NewClientConnection(conn net.Conn) *ClientConnection {
	cc := &ClientConnection{
		ID:     pkg.GenerateUniqueID(),
		Conn:   conn,
		Buffer: make([]byte, 4096),
	}

	return cc
}

func (cc *ClientConnection) ProcessRequest() error {
	fmt.Println("ClientConnection processing request...")
	defer cc.Close()

	// Store user request in a buffer
	_, err := cc.Conn.Read(cc.Buffer)
	if err != nil {
		msg := "ClientConnection.ProcessRequest: Could not read client message through buffer."
		return pkg.HandleErrorFormat(msg, err)
	}

	// Client will just send filename
	fileName := string(cc.Buffer)
	m, err := pkg.CheckFileExists(fileName)
	if err != nil {
		msg := fmt.Sprintf("ClientConnection.ProcessRequest: File does not exist or invalid name.")
		return pkg.HandleErrorFormat(msg, err)
	}

	// Echo message back to user
	convertedMsg := []byte(m)
	_, err = cc.Conn.Write(convertedMsg)
	if err != nil {
		msg := "ClientConnection.ProcessRequest: Could not write message back to Client."
		return pkg.HandleErrorFormat(msg, err)
	}

	fmt.Println(m)
	fmt.Println(convertedMsg)
	fmt.Println("ClientConnection processing request concluding...")
	return nil
}

func (cc *ClientConnection) Close() {
	cc.Conn.Close()
	fmt.Println("Closing client connection")
}
