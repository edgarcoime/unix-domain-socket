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
	listener, err := net.Listen("unix", pkg.DEFAULT_SOCKET_FILE)
	if err != nil {
		return err
	}

	// Setup Tear Down
	defer dss.Close()
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	for {
		// Accept inc connections
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		fmt.Printf("%+v\n", conn)

		// Instantiate client connection
		// create go routine to parse client messages
	}
}

func (dss *DomainSocketServer) ProcessFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	// Attempt to open file, handle error, and defer close
	if err != nil {
		return "", fmt.Errorf("ProcessFile: Error opening file %s: %w", filepath, err)
	}
	defer f.Close()

	// Scan first line as sample, handle error
	var sb strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ProcessFile: Error scanning lines for file %s: %w", filepath, err)
	}

	return sb.String(), nil
}

func (dss *DomainSocketServer) Close() {
	// Cleanup server and destroy any used resources
	// Cleanup Socket file
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()
}
