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
	MAX_CMD_ARGUMENTS   = 6
	DEFAULT_MAX_CLIENTS = pkg.DEFAULT_MAX_CLIENTS
	DEFAULT_SERVER_ADDR = pkg.DEFAULT_SERVER_ADDR
	TYPE                = pkg.SERVER_TYPE
)

type ServerParams struct {
	MaxClients uint
	Port       string
	Address    string
}

func NewServerParams() *ServerParams {
	return &ServerParams{
		MaxClients: DEFAULT_MAX_CLIENTS,
		Port:       "",
		Address:    "",
	}
}

func (sp *ServerParams) SetAddress(s string) *ServerParams {
	sp.Address = s
	return sp
}

func (sp *ServerParams) SetPort(s string) *ServerParams {
	sp.Port = s
	return sp
}

func (sp *ServerParams) SetMaxClient(n uint) *ServerParams {
	sp.MaxClients = n
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

	// Set Connection Type flags
	var paramAddress string
	var paramPort string

	// Set Server Flags
	var paramMaxClient uint

	// Parse flags
	flag.StringVar(
		&paramAddress, "a", DEFAULT_SERVER_ADDR,
		// NOTE: Check if need to say valid ip address
		"A valid address the server will bind and listen to. Otherwise will default to loopback or local address (ie. 0.0.0.0)",
	)
	flag.StringVar(
		&paramPort, "p", "",
		"A valid port the server will bind and listen to.",
	)
	flag.UintVar(
		&paramMaxClient, "n", DEFAULT_MAX_CLIENTS,
		"The number of client connections the server will simultaneously handle.",
	)
	flag.Parse()
	if paramPort == "" {
		log.Fatal("Missing required parameter port (-p), The server needs a port to listen to.")
	}

	// Manage optional params
	serverOptions := NewServerParams().SetAddress(paramAddress).SetPort(paramPort).SetMaxClient(paramMaxClient)
	var serverOpts []server.ServerOptsFunc
	serverOpts = append(serverOpts, server.ServerWithAddress(serverOptions.Address))
	serverOpts = append(serverOpts, server.ServerWithPort(serverOptions.Port))
	serverOpts = append(serverOpts, server.ServerWithMaxClients(serverOptions.MaxClients))

	// Create and start server
	networkServer := server.NewNetworkSocketServer(serverOpts...)
	msg := `Starting Network Socket Server (%s) with the following configurations:
	Address: %s
	Port: %s
	MaxClients: %d
Listening to requests now...`
	fmt.Printf(msg, TYPE, networkServer.Opts.Address, networkServer.Opts.Port, networkServer.Opts.MaxClients)
	fmt.Println("")
	err := networkServer.Start()
	if err != nil {
		log.Fatal(err)
	}
}
