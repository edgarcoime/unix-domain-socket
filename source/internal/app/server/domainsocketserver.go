package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

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
type TeardownFunc func(*DomainSocketServer)

type DomainSocketServer struct {
	Opts          DomainSocketServerOpts
	connections   map[int64]*ClientConnection
	joining       chan *ClientConnection
	leaving       chan *ClientConnection
	clientErrors  chan error
	teardownFuncs []TeardownFunc
}

func NewDomainSocketServer(opts ...DSSOptsFunc) *DomainSocketServer {
	// Set default options but also check for other custom opts
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	dss := &DomainSocketServer{
		Opts: o,
		// Handle non-negotiable attributes
		connections:   make(map[int64]*ClientConnection),
		joining:       make(chan *ClientConnection), // Incoming clients that need to be processed
		leaving:       make(chan *ClientConnection), // Outgoing clients that need to be exited
		clientErrors:  make(chan error),
		teardownFuncs: []TeardownFunc{},
	}

	// Teardown will at least have dss.Close
	teardownFunc := func() TeardownFunc {
		return func(s *DomainSocketServer) {
			fmt.Println("Teardown: dss.close")
			s.close()
		}
	}()
	dss.teardownFuncs = append(dss.teardownFuncs, teardownFunc)

	// Setup Teardown lifeline
	c := make(chan os.Signal, 1)
	signal.Notify(c) // Notifies c if os calls signal (no args means everything)
	go func(server *DomainSocketServer) {
		<-c
		fmt.Println("OS Signal interrupt shutting down...")
		server.Shutdown()
		os.Exit(1)
	}(dss)

	return dss
}

func (dss *DomainSocketServer) Shutdown() {
	// Shutdown in reverse order
	for i := len(dss.teardownFuncs) - 1; i >= 0; i-- {
		dss.teardownFuncs[i](dss)
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
	fmt.Println("Client connecting...")
	if len(dss.connections) >= int(dss.Opts.MaxConn) {
		cc.Close()
	}

	// Establish client connected
	dss.connections[cc.ID] = cc

	// Process client request and handle bubble up error
	err := cc.ProcessRequest()
	if err != nil {
		dss.handleClientError(err)
	}

	dss.leave(cc)
	fmt.Println("Client connecting ends...")
}

func (dss *DomainSocketServer) leave(cc *ClientConnection) {
	fmt.Println("Client disconnecting...")
	// cleanup resources
	delete(dss.connections, cc.ID)
	cc.Close()
}

func (dss *DomainSocketServer) handleClientError(err error) {
	fmt.Println("Handling client error...")
	msg := fmt.Sprintf("Internal Client Error: ")
	log.Println(pkg.HandleErrorFormat(msg, err))
	fmt.Println("")
}

func (dss *DomainSocketServer) Start() error {
	defer dss.Shutdown()

	fmt.Println("Starting server...")

	// Activate Listener
	listener, err := net.Listen("unix", dss.Opts.Socket)
	if err != nil {
		log.Fatalf("Error occured during net.Listen: %s\n", err)
	}

	teardownFunc := func(l net.Listener) TeardownFunc {
		return func(s *DomainSocketServer) {
			fmt.Println("Teardown: closing net connection")
			err := l.Close()
			if err != nil {
				log.Printf("Error occured while closing net connection: %s\n", err)
			}
		}
	}(listener)
	dss.teardownFuncs = append(dss.teardownFuncs, teardownFunc)

	fmt.Println(len(dss.teardownFuncs))

	// Communication loop for echo server
	for {
		// Accept connection
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Handle connection in a goroutine
		go func(nc net.Conn) {
			defer nc.Close()

			// Create buffer for incoming data
			buf := make([]byte, 4096)

			// NEED n so that null bytes are ommitted when converting to string
			n, err := nc.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}

			// only Slices can be converted to string not []byte
			filepath := string(buf[:n])

			// Search for file
			msg, err := pkg.CheckFileExists(filepath)
			if err != nil {
				log.Println(err)
				msg = "File does not exist. Please check the name and directory."
			}
			println("Server Msg: ", msg)

			// convert msg into bytes
			outgoing := []byte(msg)

			// Echo back message to client connection
			_, err = nc.Write(outgoing)
			if err != nil {
				log.Println(err)
			}
		}(conn)
	}
}

func (dss *DomainSocketServer) close() {
	// Cleanup server and destroy any used resources
	close(dss.joining)
	close(dss.leaving)
	close(dss.clientErrors)
	os.Remove(dss.Opts.Socket)
}
