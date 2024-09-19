package server

import (
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
	defer cc.Close()

	// Store user request in a buffer
	m, err := cc.Conn.Read(cc.Buffer)
	if err != nil {
		msg := "ClientConnection.ProcessRequest: Could not read client message through buffer."
		return pkg.HandleErrorFormat(msg, err)
	}

	// Echo message back to user
	_, err = cc.Conn.Write(cc.Buffer[:m])
	if err != nil {
		msg := "ClientConnection.ProcessRequest: Could not write message back to Client."
		return pkg.HandleErrorFormat(msg, err)
	}

	return nil
}

func (cc *ClientConnection) Close() {
	cc.Conn.Close()
}
