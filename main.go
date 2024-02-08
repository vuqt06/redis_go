package main

import (
	"fmt"
	"io"
	"net"
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
		// Create a buffer to read the data
		resp := NewResp(conn)

		// Read message from client
		msg, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
				return
			}
			fmt.Println(err)
			return
		}

		// Print the message
		fmt.Println("Message:", msg)

		// ignore request and send back a PONG
		conn.Write([]byte("+OK\r\n"))
	}
}
