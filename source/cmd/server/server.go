package main

import (
	"fmt"
	"log"

	"github.com/edgarcoime/domainsocket/internal/app/server"
	"github.com/edgarcoime/domainsocket/internal/pkg"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("Server Application\nMax Clients %d\n", pkg.MaxClients)

	// Instantiate client
	clientConn, err := server.NewClientConn()
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate Domain Socket Server
	dss, err := server.NewDomainSocketServer()
	if err != nil {
		log.Fatal(err)
	}

	// instantiate server

	// Activate server loop to look for connection
	// -> Any connectinos to the server instantiate a ClientConn struct
}
