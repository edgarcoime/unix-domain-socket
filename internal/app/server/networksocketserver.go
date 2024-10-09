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
	SERVER_TYPE         = pkg.SERVER_TYPE
)

// DEFAULTS FOR DOMAINSOCKETSERVER CONSTRUCTOR
type ServerOptsFunc func(*ServerOpts)

type ServerOpts struct {
	Address    string
	Port       string
	MaxClients uint
}

func defaultOpts() ServerOpts {
	return ServerOpts{
		Address:    "",
		Port:       "",
		MaxClients: DEFAULT_MAX_CLIENTS,
	}
}

func ServerWithMaxClients(n uint) ServerOptsFunc {
	return func(opts *ServerOpts) {
		opts.MaxClients = n
	}
}

func ServerWithAddress(s string) ServerOptsFunc {
	return func(opts *ServerOpts) {
		opts.Address = s
	}
}

func ServerWithPort(s string) ServerOptsFunc {
	return func(opts *ServerOpts) {
		opts.Port = s
	}
}

// DOMAINSOCKETSERVER STRUCT
type TeardownFunc func(*NetworkSocketServer)

type NetworkSocketServer struct {
	Opts          ServerOpts
	connections   map[int64]*ClientConnection
	joining       chan *ClientConnection
	leaving       chan *ClientConnection
	clientErrors  chan *ClientConnectionError
	teardownFuncs []TeardownFunc
}

func NewNetworkSocketServer(opts ...ServerOptsFunc) *NetworkSocketServer {
	// Set default options but also check for other custom opts
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	server := &NetworkSocketServer{
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
		return func(s *NetworkSocketServer) {
			fmt.Println("Teardown: dss.close")
			s.close()
		}
	}()
	server.teardownFuncs = append(server.teardownFuncs, teardownFunc)

	// Setup Teardown lifeline
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM) // Notifies c if os calls signal (no args means everything)
	go func(server *NetworkSocketServer) {
		s := <-c // Blocking call. Cleanup will only be called if something is sent to c chan
		fmt.Println("OS Signal detected shutting down...")
		fmt.Println("Got signal: ", s)
		server.Shutdown()
		os.Exit(1)
	}(server)

	return server
}

func (s *NetworkSocketServer) Start() error {
	// Activate Listener
	fullAddr := fmt.Sprintf("%s:%s", s.Opts.Address, s.Opts.Port)
	tcpAddr, err := net.ResolveTCPAddr(SERVER_TYPE, fullAddr)
	if err != nil {
		log.Fatalf("Invalid Address format please check address format of the following: %#v\n%s\n", tcpAddr, err)
	}

	listener, err := net.ListenTCP(SERVER_TYPE, tcpAddr)
	if err != nil {
		log.Fatalf("Error occured during net.Listen: %s\n", err)
	}

	teardownFunc := func(l net.Listener) TeardownFunc {
		return func(s *NetworkSocketServer) {
			fmt.Println("Teardown: closing net connection")
			err := l.Close()
			if err != nil {
				log.Printf("Error occured while closing net connection: %s\n", err)
			}
		}
	}(listener)
	s.teardownFuncs = append(s.teardownFuncs, teardownFunc)

	// Activate goroutine to handle All channels
	s.listen()
	// Activate goroutine to handle incoming connections
	s.handleConnections(listener)

	return nil
}

func (s *NetworkSocketServer) listen() {
	// Goroutine responsible for handling any clients coming in through the channels
	go func() {
		for {
			select {
			case conn := <-s.joining:
				s.join(conn)
			case conn := <-s.leaving:
				s.leave(conn)
			case err := <-s.clientErrors:
				s.handleClientError(err)
			}
		}
	}()
}

func (s *NetworkSocketServer) handleConnections(l net.Listener) {
	for {
		// Accept connection
		conn, err := l.Accept()
		if err != nil {
			// TODO: how to handle if client does not accept cause infinite loops
			msg := fmt.Sprintf("DomainSocketServer.handleConnections: Failed to accept client")
			s.clientErrors <- NewCCError(nil, pkg.HandleErrorFormat(msg, err))
			continue
		}

		client := NewClientConnection(conn, s)
		s.joining <- client
	}
}

func (s *NetworkSocketServer) join(cc *ClientConnection) {
	fmt.Println("Client connecting...")
	numClients := s.NumCurrentClients()
	if numClients >= int(s.Opts.MaxClients) {
		msg := fmt.Sprintf(
			"Sorry, we are currently at full capacity with %d clients. Please try again later.",
			numClients,
		)
		cc.WriteToClient(msg)
		return
	}

	// Establish client connected
	s.connections[cc.ID] = cc
	numClients = s.NumCurrentClients()
	fmt.Printf("Currently have %d clients...\n", numClients)

	// Goroutine the client request
	// All communication needs to be done through channels
	// go cc.ProcessRequest(dss.leaving, dss.clientErrors)
	go cc.Start()
}

func (s *NetworkSocketServer) leave(cc *ClientConnection) {
	// cleanup client resources
	delete(s.connections, cc.ID)
	fmt.Printf("Client disconnecting...\nNumber of Clients now: %d\n", s.NumCurrentClients())
}

func (s *NetworkSocketServer) handleClientError(ccErr *ClientConnectionError) {
	msg := fmt.Sprintf("Internal Client Error: ")
	log.Println(pkg.HandleErrorFormat(msg, ccErr.Error))
}

func (s *NetworkSocketServer) close() {
	// Cleanup server and destroy any used resources
	close(s.joining)
	close(s.leaving)
	close(s.clientErrors)
}

func (s *NetworkSocketServer) Shutdown() {
	// Shutdown in reverse order
	for i := len(s.teardownFuncs) - 1; i >= 0; i-- {
		s.teardownFuncs[i](s)
	}
}

func (s *NetworkSocketServer) NumCurrentClients() int {
	return len(s.connections)
}
