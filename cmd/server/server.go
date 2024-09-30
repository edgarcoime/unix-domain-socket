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
	DEFAULT_MAX_CLIENTS = pkg.DEFAULT_MAX_CLIENTS
	MAX_CMD_ARGUMENTS   = 4
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
		&paramAddress, "a", "",
		// NOTE: Check if need to say valid ip address
		"A valid address the server will bind and listen to. Otherwise will default to local port",
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
	var dssOpts []server.DSSOptsFunc
	dssOpts = append(dssOpts, server.DSSWithAddress(serverOptions.Address))
	dssOpts = append(dssOpts, server.DSSWithPort(serverOptions.Port))
	dssOpts = append(dssOpts, server.DSSWithMaxClients(serverOptions.MaxClients))

	// Create and start server
	dss := server.NewDomainSocketServer(dssOpts...)
	msg := `Starting up Domain Socket Server with the following configurations:
	Address: %s
	Port: %s
	MaxClients: %d
Listening to requests now...`
	fmt.Printf(msg, dss.Opts.Address, dss.Opts.Port, dss.Opts.MaxClients)
	fmt.Println("")
	err := dss.Start()
	if err != nil {
		log.Fatal(err)
	}
}
