package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	DEFAULT_SOCKET_FILE = pkg.DEFAULT_SOCKET_FILE
	DEFAULT_FILEPATH    = ""
)

type ClientOptsFunc func(*ClientOpts)

type ClientOpts struct {
	SocketFile string
	Filepath   string
}

func defaultOpts() ClientOpts {
	return ClientOpts{
		SocketFile: DEFAULT_SOCKET_FILE,
		Filepath:   DEFAULT_FILEPATH,
	}
}

func withSocketFile(s string) ClientOptsFunc {
	return func(opts *ClientOpts) {
		opts.SocketFile = s
	}
}

func withFilepath(s string) ClientOptsFunc {
	return func(opts *ClientOpts) {
		opts.Filepath = s
	}
}

func NewClientOpts(opts ...ClientOptsFunc) *ClientOpts {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}
	return &o
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var paramSocket string
	var paramFilename string

	flag.StringVar(
		&paramSocket, "s", DEFAULT_SOCKET_FILE,
		"A valid path to the socket file that the client will attempt to bind and listen to (ex. \"/tmp/example.sock\").",
	)
	flag.StringVar(
		&paramFilename, "f", DEFAULT_FILEPATH,
		"A valid path to a file that the client will request the server to check",
	)

	// Parse flags
	flag.Parse()
	var opts []ClientOptsFunc
	opts = append(opts, withSocketFile(pkg.StringInputParser(paramSocket)))
	// Parse required arguments
	if paramFilename == "" {
		log.Fatal("Missing required parameter socket (-s), the client needs a path to a file to ask the server about.")
	}
	// Create options struct
	opts = append(opts, withSocketFile(pkg.StringInputParser(paramSocket)))
	opts = append(opts, withFilepath(pkg.StringInputParser(paramFilename)))
	clientOptions := NewClientOpts(opts...)

	// establish connection
	conn, err := net.Dial("unix", clientOptions.SocketFile)
	if err != nil {
		log.Fatalf("Failed to connect to the socket: %s", err)
	}
	defer conn.Close()

	// Write message to the server
	outboundMsg := []byte(clientOptions.Filepath)
	_, err = conn.Write(outboundMsg)
	if err != nil {
		log.Fatalf("Failed to write to the socket: %s", err)
	}

	// Read inbound message from the server
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from socket: %s", err)
	}

	inboundMsg := string(buf[:n])
	fmt.Printf("Server Response: %s\n", inboundMsg)
}
