package main

import (
	"fmt"
	"io"
	"net"
	"strings"
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

	// Create or open an AOF file
	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	// Read the AOF file and rebuild the database
	aof.Read(func(value Value) {
		// Get the command
		value.array = value.array[len(value.array)/2:]
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		// Get the handler for the command
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		// Execute the handler
		handler(args)
	})

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

		if msg.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(msg.array) == 0 {
			fmt.Println("Invalid request, empty array")
			continue
		}

		// Get the command
		msg.array = msg.array[len(msg.array)/2:]
		command := strings.ToUpper(msg.array[0].bulk)
		args := msg.array[1:]
		writer := NewWriter(conn)

		// Get the handler for the command
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		// Write to the AOF file
		if command == "SET" || command == "HSET" {
			aof.Write(msg)
		}

		// Execute the handler
		result := handler(args)
		writer.Write(result)
	}
}
