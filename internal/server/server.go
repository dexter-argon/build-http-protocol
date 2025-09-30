package server

import (
	"build-http-protocol/internal/request"
	"build-http-protocol/internal/response"
	"fmt"
	"net"
)

type Handler func(w *response.Writer, req *request.Request) *HandlerError

type Server struct {
	handler Handler
	closed  bool
}
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed {
			return
		}
		if err != nil {
			return
		}
		go handleConnection(s, conn)
	}
}

func writeErrors(w *response.Writer, err *HandlerError) {
	h := response.GetDefaultHeaders(len(err.Message))
	w.WriteStatusLine(err.StatusCode)
	w.WriteHeaders(h)
	w.WriteBody([]byte(err.Message))
}

func handleConnection(s *Server, conn net.Conn) {
	defer conn.Close()
	// 1. parse the request from connection
	writer := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)

	if err != nil {
		writeErrors(writer, &HandlerError{
			Message:    err.Error(),
			StatusCode: response.StatusBadRequest,
		})
		return
	}

	// 2. create empty bytes buffer for handler to write to.
	// 3. call handler function
	// 4. if handler errs then write the error message to connection

	handleError := s.handler(writer, req)
	if handleError != nil {
		writeErrors(writer, handleError)
		return
	}

	// 5. if handler succeeds
	fmt.Printf("We are handling your request CLIENT!\n")
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		handler: handler,
		closed:  false,
	}
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
