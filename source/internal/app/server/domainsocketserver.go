package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	DEFAULT_MAX_CLIENTS = pkg.DEFAULT_MAX_CLIENTS
	DEFAULT_SOCKET_FILE = pkg.DEFAULT_SOCKET_FILE
)

// DEFAULTS FOR DOMAINSOCKETSERVER CONSTRUCTOR
type DSSOptsFunc func(*DomainSocketServerOpts)

type DomainSocketServerOpts struct {
	MaxConn uint
	Socket  string
}

func defaultOpts() DomainSocketServerOpts {
	return DomainSocketServerOpts{
		MaxConn: DEFAULT_MAX_CLIENTS,
		Socket:  DEFAULT_SOCKET_FILE,
	}
}

func DSSWithMaxConn(n uint) DSSOptsFunc {
	return func(opts *DomainSocketServerOpts) {
		opts.MaxConn = n
	}
}

func DSSWithSocket(s string) DSSOptsFunc {
	return func(opts *DomainSocketServerOpts) {
		opts.Socket = s
	}
}

// DOMAINSOCKETSERVER STRUCT
type DomainSocketServer struct {
	Opts         DomainSocketServerOpts
	connections  map[int64]*ClientConnection
	joining      chan *ClientConnection
	leaving      chan *ClientConnection
	clientErrors chan error
}

func NewDomainSocketServer(opts ...DSSOptsFunc) *DomainSocketServer {
	// Set default options but also check for other custom opts
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return &DomainSocketServer{
		Opts: o,
		// Handle non-negotiable attributes
		connections:  make(map[int64]*ClientConnection),
		joining:      make(chan *ClientConnection), // Incoming clients that need to be processed
		leaving:      make(chan *ClientConnection), // Outgoing clients that need to be exited
		clientErrors: make(chan error),
	}
}

func (dss *DomainSocketServer) listen() {
	// Goroutine responsible for handling any clients coming in through the channels
	go func() {
		for {
			select {
			case conn := <-dss.joining:
				dss.join(conn)
			case conn := <-dss.leaving:
				dss.leave(conn)
			case err := <-dss.clientErrors:
				dss.handleClientError(err)
			}
		}
	}()
}

func (dss *DomainSocketServer) join(cc *ClientConnection) {
	if len(dss.connections) >= int(dss.Opts.MaxConn) {
		cc.Close()
	}

	// Establish client connected
	dss.connections[cc.ID] = cc

	// Process client request and handle bubble up error
	err := cc.ProcessRequest()
	if err != nil {
		dss.clientErrors <- err
		dss.leaving <- cc
	}

	dss.leaving <- cc
}

func (dss *DomainSocketServer) leave(cc *ClientConnection) {
	// cleanup resources
	delete(dss.connections, cc.ID)
	cc.Close()
}

func (dss *DomainSocketServer) handleClientError(err error) {
	msg := fmt.Sprintf("Internal Client Error: ")
	log.Println(pkg.HandleErrorFormat(msg, err))
}

func (dss *DomainSocketServer) Start() error {
	// Setup Connection
	listener, err := net.Listen("unix", dss.Opts.Socket)
	if err != nil {
		return pkg.HandleErrorFormat("DomainSocketServer.Listen: Error listening to socket", err)
	}

	// Setup Tear Down function to catch signal interrupts
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(l net.Listener, s *DomainSocketServer) {
		<-c
		s.close()
		err := l.Close()
		if err != nil {
			log.Fatal(pkg.HandleErrorFormat("DomainSocketServer.Listen: Error shutting down listener", err))
		}
	}(listener, dss)

	// Initiate server listening to communication channels in a seperate goroutine
	dss.listen()

	// Initiate locking while loop to parse new connections and/or leave requests
	for {
		// Accept inc connections and handle client errors
		conn, err := listener.Accept()
		if err != nil {
			msg := fmt.Sprintf("DomainSocketServer.Listen: Failed to accept client %s", conn.LocalAddr().String())
			dss.clientErrors <- pkg.HandleErrorFormat(msg, err)
		}

		// Instantiate connection
		newCC := NewClientConnection(conn)
		dss.joining <- newCC
	}

	// Locking loop to handle incoming requests and handling channels

	var failedConns []net.Addr
	var clientCount uint16 = 0
	for {
		// Accept inc connections
		conn, err := listener.Accept()
		if err != nil {
			msg := fmt.Sprintf("DomainSocketServer.Listen: Failed to accept client %s", conn.LocalAddr().String())
			log.Println(pkg.HandleErrorFormat(msg, err))
			failedConns = append(failedConns, conn.LocalAddr())
			continue
		}

		clientCount += 1
		fmt.Printf("Client %d JOINED\n", clientCount)

		// Instantiate client connection
		// create go routine to parse client messages
		// Handle simple echo server
		go func(c net.Conn) {
			defer c.Close()

			// Create buffer for incoming data
			buf := make([]byte, 4096)

			// Read data from connection
			n, err := c.Read(buf)
			if err != nil {
				// Error from buffer should d/c client not shut down whole server
				log.Printf(err.Error())
				return
			}

			// Echo data back to the client
			_, err = conn.Write(buf[:n])
			if err != nil {
				// Error from buffer should d/c client not shut down whole server
				log.Printf(err.Error())
				return
			}
		}(conn)
	}
}

func (dss *DomainSocketServer) processFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	// Attempt to open file, handle error, and defer close
	if err != nil {
		return "", pkg.HandleErrorFormat(
			fmt.Sprintf("DomainSocketServer.ProcessFile: Error opening file in \"%s\"", filepath),
			err,
		)
	}
	defer f.Close()

	// Scan first line as sample, handle error
	var sb strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", pkg.HandleErrorFormat(
			fmt.Sprintf("DomainSocketServer.ProcessFile: Error Scanning lines for file in \"%s\"", filepath),
			err,
		)
	}

	return sb.String(), nil
}

func (dss *DomainSocketServer) close() {
	// Cleanup server and destroy any used resources
	// Cleanup Socket file
	fmt.Println("Server cleanup starting")
	os.Remove(dss.Opts.Socket)
}
