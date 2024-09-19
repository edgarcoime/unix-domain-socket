package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/edgarcoime/domainsocket/internal/app/server"
	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	DEFAULT_SOCKET_FILE = pkg.DEFAULT_SOCKET_FILE
	DEFAULT_MAX_CLIENTS = pkg.DEFAULT_MAX_CLIENTS
	MAX_CMD_ARGUMENTS   = 4
)

type ServerParams struct {
	socketFile string
	maxClients uint
}

func NewServerParams() *ServerParams {
	return &ServerParams{
		socketFile: DEFAULT_SOCKET_FILE,
		maxClients: DEFAULT_MAX_CLIENTS,
	}
}

func (sp *ServerParams) SetSocketFile(s string) *ServerParams {
	sp.socketFile = s
	return sp
}

func (sp *ServerParams) SetMaxClient(n uint) *ServerParams {
	sp.maxClients = n
	return sp
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// Validate max amount of args
	if len(os.Args) > MAX_CMD_ARGUMENTS+1 {
		msg := fmt.Sprintf(`
The Server application only allows for %d arguments including flags.
You do not need to supply a flag to run the program but look at the -h docs to change functionality.
			-h : Supplies a help menu for arguments
`, MAX_CMD_ARGUMENTS)
		log.Fatal(msg)
	}

	// Set flags
	var paramSocket string
	var paramMaxClient uint

	// Parse flags
	flag.StringVar(
		&paramSocket, "s", DEFAULT_SOCKET_FILE,
		"A valid path to the socket file the server will bind and listen to (ex. \"/tmp/example.sock\").",
	)
	flag.UintVar(
		&paramMaxClient, "n", DEFAULT_MAX_CLIENTS,
		"The number of client connections the server will simultaneously handle.",
	)
	flag.Parse()

	// Manage optional params
	serverOptions := NewServerParams().SetSocketFile(paramSocket).SetMaxClient(paramMaxClient)
	var dssOpts []server.DSSOptsFunc
	dssOpts = append(dssOpts, server.DSSWithSocket(serverOptions.socketFile))
	dssOpts = append(dssOpts, server.DSSWithMaxConn(serverOptions.maxClients))

	// Create and start server
	dss := server.NewDomainSocketServer(dssOpts...)
	msg := `Starting up Domain Socket Server with the following configurations:
	Socketfile: %s
	MaxClients: %d
Listening to requests now...`
	fmt.Printf(msg, dss.Opts.Socket, dss.Opts.MaxConn)
	fmt.Println("")
	err := dss.Start()
	if err != nil {
		log.Fatal(err)
	}
}
