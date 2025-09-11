package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	var line string
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			str := string(data[:n])
			if idx := strings.Index(str, "\n"); idx != -1 {
				line += str[:idx]
				out <- line
				data = data[idx+1:]
				line = ""
			}

			line += string(data)
		}

		if len(line) > 0 {
			out <- line
		}
	}()

	return out
}

func main() {
	// f, err := os.Open("message.txt")
	// if err != nil {
	// 	log.Fatal("error opening the file message.txt")
	// }

	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error starting the server. Error: ", err.Error())
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting a connection. Error message: %s", err.Error())
			continue
		}
		go ConnectionHandler(conn)
	}
}

func ConnectionHandler(conn net.Conn) {
	// Handling connection
	fmt.Println("handling connection...")
	lines := getLinesChannel(conn)

	for line := range lines {
		fmt.Printf("Read: %s\n", line)
	}
}
