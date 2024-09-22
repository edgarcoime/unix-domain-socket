package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	DEFAULT_SOCKET_FILE = pkg.DEFAULT_SOCKET_FILE
	DEFAULT_FILEPATH    = ""
	MAX_CMD_ARGUMENTS   = 4
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

func ParseFlags() *ClientOpts {
	// Parse arguments and user input
	var paramSocket string
	var paramFilename string

	// Validate max amount of args
	if len(os.Args) > MAX_CMD_ARGUMENTS+1 {
		msg := fmt.Sprintf(`
The Client application only allows for %d arguments including flags.
Please supply at least the desired filename to run the program or use the following helper:
			-h : Supplies a help menu for arguments
`, MAX_CMD_ARGUMENTS)
		log.Fatal(msg)
	}

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
	if paramFilename == "" {
		log.Fatal("Missing required parameter socket (-f), the client needs a path to a file to ask the server about.")
	}

	// Create options struct
	opts = append(opts, withSocketFile(pkg.StringInputParser(paramSocket)))
	opts = append(opts, withFilepath(pkg.StringInputParser(paramFilename)))
	return NewClientOpts(opts...)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	opts := ParseFlags()

	msg := `Starting up Client with the following configurations:
	Socketfile: %s
	Filepath: %s
Attempting to connect to server now...`
	fmt.Printf(msg, opts.SocketFile, opts.Filepath)
	fmt.Println("")

	// establish connection
	conn, err := net.Dial("unix", opts.SocketFile)
	if err != nil {
		log.Fatalf("Failed to connect to the socket: %s", err)
	}

	// intercept os signal cleanup functions
	// REMEMBER IF THERE IS AN OS EXIT YOU MUST SET THIS UP
	os_c := make(chan os.Signal, 1)
	signal.Notify(os_c, syscall.SIGINT, syscall.SIGTERM)
	defer conn.Close()
	go func(c net.Conn) {
		s := <-os_c
		fmt.Println("Sig call shutdown")
		fmt.Println("Os signal: ", s)
		c.Close()
		os.Exit(1)
	}(conn)

	fmt.Println("writing start")
	// Write message to the server
	outboundMsg := []byte(opts.Filepath)
	_, err = conn.Write(outboundMsg)
	if err != nil {
		log.Fatalf("Failed to write to the socket: %s", err)
	}
	fmt.Println("writing end")

	fmt.Println("reading start")
	// Read inbound message from the server
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from socket: %s", err)
	}
	fmt.Println("reading end")

	inboundMsg := string(buf[:n])
	fmt.Printf("Server Response: %s\n", inboundMsg)
}
