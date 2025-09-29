package main

import (
	"build-http-protocol/internal/request"
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
	listner, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error starting the server. Error: ", err.Error())
	}
	defer listner.Close()

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal("error accepting a connection")
		}
		go ConnectionHandler(conn)
	}
}

func ConnectionHandler(conn net.Conn) {
	defer conn.Close()
	fmt.Println("handling connection...")

	r, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatal("error parsing the request")
	}
	fmt.Printf("Request line:\n")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	r.Headers.ForEach(func(c, v string) {
		fmt.Printf("- %s:%v\n", c, v)
	})
	fmt.Printf("Body:\n")
	fmt.Printf("%s\n", r.Body)
}
