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
	Opts DomainSocketServerOpts
}

func NewDomainSocketServer(opts ...DSSOptsFunc) *DomainSocketServer {
	// Set default options but also check for other custom opts
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return &DomainSocketServer{
		Opts: o,
	}
}

func (dss *DomainSocketServer) Listen() error {
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
		s.Close()
		err := l.Close()
		if err != nil {
			log.Fatal(pkg.HandleErrorFormat("DomainSocketServer.Listen: Error shutting down listener", err))
		}
	}(listener, dss)

	var failedConns []net.Addr
	var clientCount uint16 = 0
	for {
		// Accept inc connections
		conn, err := listener.Accept()
		if err != nil {
			// TODO: Should we stop server if client cannot connect?
			// Move on to next client and try to accept next?
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

		fmt.Printf("Client %d LEAVES\n", clientCount)
		clientCount -= 1
	}
}

func (dss *DomainSocketServer) ProcessFile(filepath string) (string, error) {
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

func (dss *DomainSocketServer) Close() {
	// Cleanup server and destroy any used resources
	// Cleanup Socket file
	fmt.Println("Server cleanup starting")
	os.Remove(dss.Opts.Socket)
}
