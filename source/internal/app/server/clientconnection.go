package server

import (
	"fmt"
	"net"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

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

func (cc *ClientConnection) ProcessRequest(leaving chan *ClientConnection, errors chan error) {
	fmt.Println("ClientConnection processing request...")
	buf := make([]byte, 4096)

	// NEED n so that null bytes are ommitted when converting to string
	n, err := cc.Conn.Read(buf)
	if err != nil {
		msg := "ClientConnection.ProcessRequest: Could not read client message through buffer."
		cc.WriteToClient("Could not read incoming message.")
		errors <- pkg.HandleErrorFormat(msg, err)
		return
	}

	// Only slices can be converted to string not []byte
	filepath := string(buf[:n])

	// Client will just send filename
	m, err := pkg.CheckFileExists(filepath)
	if err != nil {
		msg := fmt.Sprintf("ClientConnection.ProcessRequest: File does not exist or invalid name.")
		cc.WriteToClient(m)
		errors <- pkg.HandleErrorFormat(msg, err)
		return
	}

	// Echo message back to user
	err = cc.WriteToClient(m)
	if err != nil {
		errors <- err
		return
	}

	fmt.Println(m)
	fmt.Println("ClientConnection processing request concluding...")
	leaving <- cc
}

func (cc *ClientConnection) Close() {
	cc.Conn.Close()
	fmt.Println("Closing client connection")
}
