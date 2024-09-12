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
	DEFAULT_MAX_CLIENTS = 10
)

type DomainSocketServer struct {
	Socket     string
	MaxClients uint
}

func NewDomainSocketServer() (*DomainSocketServer, error) {
	dss := &DomainSocketServer{
		Socket:     pkg.DEFAULT_SOCKET,
		MaxClients: DEFAULT_MAX_CLIENTS,
	}
	// Upon creating DSS should listen right away
	dss.Listen()

	return dss, nil
}

func (dss *DomainSocketServer) Listen() error {
	// Setup Connection
	listener, err := net.Listen("unix", pkg.DEFAULT_SOCKET)
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

		// Instantiate client connection
		// create go routine to parse client messages
	}
}

func (dss *DomainSocketServer) Close() {
	// Cleanup server and destroy any used resources
	fmt.Println("Closing server")

	// Cleanup Socket file
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(dss.Socket)
		os.Exit(1)
	}()
}
