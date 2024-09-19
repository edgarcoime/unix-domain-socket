package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		log.Fatal("Please provide a filepath relative or absolute")
	}

	filepath := os.Args[1]

	fmt.Println()
	path, _ := os.Getwd()
	fmt.Printf("%s\n", path)

	contents, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("File reading error ", err)
		return
	}

	fmt.Println("Contents of file: ", string(contents))

	// Process command line args and find out what options are needed

	// Instantiate server with options
	// Input arg functions in server
	// Handle any errors from server to user here
	// dss := server.NewDomainSocketServer()

	// With newly instantiated server listen
	// Defer cleanup
	// fmt.Printf("%+v\n", dss)

	// fmt.Println()
	// path, _ := os.Getwd()
	// fmt.Printf("%s\n", path)

	// s, err := dss.ProcessFile("~/text.txt")
	// if err != nil {
	//	log.Fatal(err)
	// }

	// fmt.Println(s)
	// Instantiate Domain Socket Server
	// Activate server loop to look for connection
	// -> Any connectinos to the server instantiate a ClientConn struct
}
