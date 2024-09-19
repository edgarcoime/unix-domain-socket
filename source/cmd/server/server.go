package main

import (
	"fmt"
	"log"
	"os"

	"github.com/edgarcoime/domainsocket/internal/app/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Process command line args and find out what options are needed
	// Handle opening file logic
	if len(os.Args) < 2 {
		log.Fatal("Please provide a filepath relative or absolute")
	}
	filePath := os.Args[1]

	hasCustomSocket := false
	var opts []server.DSSOptsFunc
	if hasCustomSocket {
		opts = append(opts, server.DSSWithSocket("/tmp/myCustomSocket"))
	}

	// Instantiate server with options
	// Input arg functions in server
	// Handle any errors from server to user here
	dss := server.NewDomainSocketServer(opts...)
	// Debugging
	fmt.Printf("%+v\n", dss)
	currentPath, _ := os.Getwd()
	fmt.Printf("%s\n", currentPath)

	// Activate the server
	err := dss.Activate()
	if err != nil {
		log.Fatal(err)
	}

	// With newly instantiated server listen
	// Defer cleanup

	fmt.Println(s)
	// Instantiate Domain Socket Server
	// Activate server loop to look for connection
	// -> Any connectinos to the server instantiate a ClientConn struct
}
