package main

import (
	"fmt"
	"log"

	"github.com/edgarcoime/domainsocket/internal/app/server"
)

const (
	MAX_CLIENTS = 10
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("Server Application\nMax Clients %d\n")

	// Instantiate Domain Socket Server
	dss, err := server.NewDomainSocketServer()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", dss.Socket)

	// instantiate server

	// Activate server loop to look for connection
	// -> Any connectinos to the server instantiate a ClientConn struct
}
