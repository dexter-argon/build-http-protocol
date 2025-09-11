package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {

	addr, err := net.ResolveUDPAddr("udp", ":42070")
	if err != nil {

	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error dialing UDP server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		// Print the prompt character.
		fmt.Print("> ")

		// Read a line from the console.
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from stdin: %v", err)
			continue
		}

		// Check for exit command.
		if line == "exit\n" {
			log.Println("Exiting program.")
			return
		}

		// Write the line to the UDP connection.
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("Error writing to UDP connection: %v", err)
			continue
		}
	}
}
