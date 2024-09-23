package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
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
	MaxClients uint
	SocketFile string
}

func defaultOpts() DomainSocketServerOpts {
	return DomainSocketServerOpts{
		MaxClients: DEFAULT_MAX_CLIENTS,
		SocketFile: DEFAULT_SOCKET_FILE,
	}
}

func DSSWithMaxClients(n uint) DSSOptsFunc {
	return func(opts *DomainSocketServerOpts) {
		opts.MaxClients = n
	}
}

func DSSWithSocketFile(s string) DSSOptsFunc {
	return func(opts *DomainSocketServerOpts) {
		opts.SocketFile = s
	}
}

// DOMAINSOCKETSERVER STRUCT
type TeardownFunc func(*DomainSocketServer)

type DomainSocketServer struct {
	Opts          DomainSocketServerOpts
	connections   map[int64]*ClientConnection
	joining       chan *ClientConnection
	leaving       chan *ClientConnection
	clientErrors  chan *ClientConnectionError
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
		joining:       make(chan *ClientConnection),
		leaving:       make(chan *ClientConnection),
		clientErrors:  make(chan *ClientConnectionError),
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
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM) // Notifies c if os calls signal (no args means everything)
	go func(server *DomainSocketServer) {
		s := <-c // Blocking call. Cleanup will only be called if something is sent to c chan
		fmt.Println("OS Signal detected shutting down...")
		fmt.Println("Got signal: ", s)
		server.Shutdown()
		os.Exit(1)
	}(dss)

	return dss
}

func (dss *DomainSocketServer) Start() error {
	// Activate Listener
	listener, err := net.Listen("unix", dss.Opts.SocketFile)
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

	// Activate goroutine to handle All channels
	dss.listen()
	// Activate goroutine to handle incoming connections
	dss.handleConnections(listener)

	return nil
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

func (dss *DomainSocketServer) handleConnections(l net.Listener) {
	for {
		// Accept connection
		conn, err := l.Accept()
		if err != nil {
			// TODO: how to handle if client does not accept cause infinite loops
			msg := fmt.Sprintf("DomainSocketServer.handleConnections: Failed to accept client")
			dss.clientErrors <- NewCCError(nil, pkg.HandleErrorFormat(msg, err))
			continue
		}

		client := NewClientConnection(conn)
		dss.joining <- client
	}
}

func (dss *DomainSocketServer) join(cc *ClientConnection) {
	fmt.Println("Client connecting...")
	numClients := dss.NumCurrentClients()
	if numClients >= int(dss.Opts.MaxClients) {
		msg := fmt.Sprintf(
			"Sorry, we are currently at full capacity with %d clients. Please try again later.",
			numClients,
		)
		cc.WriteToClient(msg)
		return
	}

	// Establish client connected
	dss.connections[cc.ID] = cc
	numClients = dss.NumCurrentClients()
	fmt.Printf("Currently have %d clients...", numClients)

	// Goroutine the client request
	// All communication needs to be done through channels
	go cc.ProcessRequest(dss.leaving, dss.clientErrors)
}

func (dss *DomainSocketServer) leave(cc *ClientConnection) {
	// cleanup client resources
	defer cc.Close()
	delete(dss.connections, cc.ID)
	fmt.Printf("Client disconnecting...\nNumber of Clients now: %d\n", dss.NumCurrentClients())
}

func (dss *DomainSocketServer) handleClientError(ccErr *ClientConnectionError) {
	msg := fmt.Sprintf("Internal Client Error: ")
	log.Println(pkg.HandleErrorFormat(msg, ccErr.Error))
}

func (dss *DomainSocketServer) close() {
	// Cleanup server and destroy any used resources
	close(dss.joining)
	close(dss.leaving)
	close(dss.clientErrors)
	os.Remove(dss.Opts.SocketFile)
}

func (dss *DomainSocketServer) Shutdown() {
	// Shutdown in reverse order
	for i := len(dss.teardownFuncs) - 1; i >= 0; i-- {
		dss.teardownFuncs[i](dss)
	}
}

func (dss *DomainSocketServer) NumCurrentClients() int {
	return len(dss.connections)
}
