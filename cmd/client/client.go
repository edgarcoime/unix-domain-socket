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
	DEFAULT_ADDR        = pkg.DEFAULT_SERVER_ADDR
	CONN_TYPE           = pkg.SERVER_TYPE
	DEFAULT_FILEPATH    = ""
	MAX_CMD_ARGUMENTS   = 4
)

type ClientOptsFunc func(*ClientOpts)

type ClientOpts struct {
	Address  string
	Port     string
	Filepath string
}

func defaultOpts() ClientOpts {
	return ClientOpts{
		Address:  DEFAULT_ADDR,
		Port:     "",
		Filepath: DEFAULT_FILEPATH,
	}
}

func withAddress(s string) ClientOptsFunc {
	return func(opts *ClientOpts) {
		opts.Address = s
	}
}

func withPort(s string) ClientOptsFunc {
	return func(opts *ClientOpts) {
		opts.Port = s
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

type ServerConnection struct {
	LocalAddr  *net.TCPAddr
	RemoteAddr *net.TCPAddr
	Conn       *net.TCPConn
	Filepath   string
}

func NewServerConnection(r *net.TCPAddr, fp string) *ServerConnection {
	return &ServerConnection{
		LocalAddr:  nil,
		RemoteAddr: r,
		Conn:       nil,
		Filepath:   fp,
	}
}

func (sc *ServerConnection) Start() {
	// Attempt connection
	var localAddr *net.TCPAddr = nil
	conn, err := net.DialTCP(CONN_TYPE, localAddr, sc.RemoteAddr)
	if err != nil {
		log.Fatalf("Error occured during net.DialTCP: \nPlease check the remote address (%s).\n%s\n", sc.RemoteAddr, err)
	}

	sc.LocalAddr = localAddr
	sc.Conn = conn

	// REMEMBER IF THERE IS AN OS EXIT YOU MUST SET THIS UP
	os_c := make(chan os.Signal, 1)
	signal.Notify(os_c, syscall.SIGINT, syscall.SIGTERM)
	go func(c net.Conn) {
		s := <-os_c
		fmt.Println("Sig call shutdown")
		fmt.Println("Os signal: ", s)
		sc.Close()
		os.Exit(1)
	}(conn)
}

func (sc *ServerConnection) Close() {
	sc.Close()
}

func (sc *ServerConnection) ProcessRequest() {
	// CONNECTED TO SERVER NOW
	file, err := os.Open(sc.Filepath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Read file
	content, err := pkg.ReadTextFileContents(sc.Filepath)
	if err != nil {
		log.Fatalf("Failed to read and open the file at path: %s\n", sc.Filepath)
	}

	// Loop through chunks of the file and send chunks to the server

	// Write message to the server
	outboundMsg := []byte(content)
	_, err = sc.Conn.Write(outboundMsg)
	if err != nil {
		log.Fatalf("Failed to write to the socket: %s\n", err)
	}

	// Read inbound message from the server
	buf := make([]byte, 4096)
	n, err := sc.Conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from socket: %s\n", err)
	}

	buf.len

	inboundMsg := string(buf[:n])
	fmt.Printf("Server Response: %s\n", inboundMsg)
}

func ParseFlags() *ClientOpts {
	// Parse arguments and user input
	// Set Connection Type flags
	var paramFilename string
	var paramAddress string
	var paramPort string

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
		&paramAddress, "a", DEFAULT_ADDR,
		"A valid address the client will attempt to connect to. Otherwise will default to loopback or local address (ie. 0.0.0.0)",
	)
	flag.StringVar(
		&paramPort, "p", "",
		"A valid port the server will bind and listen to.",
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
	if paramPort == "" {
		log.Fatal("Missing required parameter port (-p), The server needs a port to listen to.")
	}

	// Create options struct
	opts = append(opts, withFilepath(pkg.StringInputParser(paramFilename)))
	opts = append(opts, withAddress(paramAddress))
	opts = append(opts, withPort(paramPort))
	return NewClientOpts(opts...)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	opts := ParseFlags()

	msg := `Starting up Client with the following configurations:
	Address: %s
	Port: %s
	Filepath: %s
Attempting to connect to server now...`
	fmt.Printf(msg, opts.Address, opts.Port, opts.Filepath)
	fmt.Println("")

	// Parse address
	fullAddr := fmt.Sprintf("%s:%s", opts.Address, opts.Port)
	remoteAddr, err := net.ResolveTCPAddr(CONN_TYPE, fullAddr)
	if err != nil {
		log.Fatalf("Invalid Address format please check address format of the following: %#v\n%s\n", fullAddr, err)
	}

	sc := NewServerConnection(remoteAddr, opts.Filepath)
	sc.Start()
	sc.ProcessRequest()
}
