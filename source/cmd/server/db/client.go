package client

import (
	"bufio"
	"github.com/google/uuid"
	"log"
	"net"
	"strings"
)

const (
	DefaultClientName = "Anonymous"
)

// Client represent a connection to the server. It abstracts the idea of a connection
// into incoming and outgoing channels as well storing information about the client's state
type Client struct {
	ID       string
	Name     string
	Incoming chan *Message
	Outgoing chan string
	Conn     net.Conn
	Reader   *bufio.Reader
	Writer   *bufio.Writer
}

func NewClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	client := &Client{
		ID:       uuid.New().String(),
		Name:     DefaultClientName,
		Incoming: make(chan *Message),
		Outgoing: make(chan string),
		Conn:     conn,
		Reader:   reader,
		Writer:   writer,
	}

	client.Listen()
	return client
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

// Responsible for reading string from the Client's sockets, formats into Messages,
// and puts them into the Client's incoming channel.
func (client *Client) Read() {
	for {
		str, err := client.Reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		message := NewMessage(client, strings.TrimSuffix(str, "\n"))
		client.Incoming <- message
	}
	close(client.Incoming)
	log.Println("Client read thread closed")
}

// Responsible reading in messages from the Client's outoing channel and writes
// to the Client's socket
func (client *Client) Write() {
	for str := range client.Outgoing {
		_, err := client.Writer.WriteString(str)
		if err != nil {
			log.Println(err)
			break
		}
		err = client.Writer.Flush()
		if err != nil {
			log.Println(err)
			break
		}
	}
	log.Println("Client write thread closed")
}

// Quit Closes the client's connection
func (client *Client) Quit() {
	// TODO: Handle close error
	client.Conn.Close()
}
