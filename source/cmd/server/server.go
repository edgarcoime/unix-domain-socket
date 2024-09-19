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
	// filePath := os.Args[1]

	hasCustomSocket := true
	var opts []server.DSSOptsFunc
	if hasCustomSocket {
		opts = append(opts, server.DSSWithSocket("/tmp/mysock.sock"))
	}

	dss := server.NewDomainSocketServer(opts...)
	// Debugging
	currentPath, _ := os.Getwd()
	fmt.Printf("%+v\n", dss)
	fmt.Printf("%s\n", currentPath)

	// Activate the server and handle error
	err := dss.Start()
	if err != nil {
		log.Fatal(err)
	}
}
