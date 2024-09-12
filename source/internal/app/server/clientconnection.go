package server

import (
	"bufio"
	"net"
)

type ClientConnection struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func NewClientConnection(conn net.Conn) (*ClientConnection, error) {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	cc := &ClientConnection{
		Conn:   conn,
		Writer: writer,
		Reader: reader,
	}

	return cc, nil
}

func (cc *ClientConnection) Listen() error {
	defer cc.Close()
	return nil
}

func (cc *ClientConnection) Read() error {
	return nil
}

func (cc *ClientConnection) Write() error {
	return nil
}

func (cc *ClientConnection) Close() {
	cc.Conn.Close()
}
