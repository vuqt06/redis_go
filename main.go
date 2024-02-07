package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Start a TCP listener so that any Client can communicate with the server
	fmt.Println("Listening on port 6379...")

	// Create a new server
	server, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Accept any incoming connection
	conn, err := server.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close() // close the connection when finished

	// Read the data from the connection
	for {
		buf := make([]byte, 1024)

		// Read message from client
		_, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading fro the client: ", err.Error())
			os.Exit(1)
		}

		// ignore request and send back a PONG
		conn.Write([]byte("+OK\r\n"))
	}
}
