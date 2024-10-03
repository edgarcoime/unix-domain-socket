package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgarcoime/domainsocket/internal/pkg"
)

const (
	CONN_TYPE = pkg.SERVER_TYPE
)

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
	sc.Conn.Close()
	print("closing")
}

func (sc *ServerConnection) ProcessRequest() {
	// CONNECTED TO SERVER NOW
	defer sc.Close()
	file, err := os.Open(sc.Filepath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	fmt.Println("found file")

	fileReader := bufio.NewReader(file)
	writer := bufio.NewWriter(sc.Conn)
	// reader := bufio.NewReader(sc.Conn)

	// Create first packet
	_, err = writer.WriteString(sc.Filepath + "\n")
	if err != nil {
		log.Fatalf("Could not send header for the packets to the server\n")
	}

	// Loop through chunks of the file and send chunks to the server
	for {
		line, err := fileReader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				if len(line) > 0 {
					// send last line if doesn't work with new line
					_, writeErr := writer.WriteString(line)
					if writeErr != nil {
						log.Printf("Error sending last line: %v\n", writeErr)
						break
					}
				}
				break
			}
			log.Fatalf("Error reading the file: %s\n", err)
		}

		// Send the line to the server
		_, err = writer.WriteString(line)
		if err != nil {
			log.Fatalf("Error sending data: %v\n", err)
		}

		// Flush the buffered data to ensure it's sent immediately
		err = writer.Flush()
		if err != nil {
			log.Fatalf("Error flushing data: %v\n", err)
		}
	}
}
