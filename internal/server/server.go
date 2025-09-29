package server

import (
	"build-http-protocol/internal/request"
	"build-http-protocol/internal/response"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

func writeErrors(w io.ReadWriteCloser, err *HandlerError) {
	h := response.GetDefaultHeaders(len(err.Message))

	response.WriteStatusLine(w, err.StatusCode)
	response.WriteHeaders(w, h)

	w.Write([]byte(err.Message))
}

func handleConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	// 1. parse the request from connection
	req, err := request.RequestFromReader(conn)

	if err != nil {
		writeErrors(conn, &HandlerError{
			Message:    err.Error(),
			StatusCode: response.StatusBadRequest,
		})
		return
	}

	// 2. create empty bytes buffer for handler to write to.
	writer := bytes.NewBuffer([]byte{})

	// 3. call handler function
	// 4. if handler errs then write the error message to connection
	handleError := s.handler(writer, req)
	if handleError != nil {
		writeErrors(conn, &HandlerError{
			StatusCode: handleError.StatusCode,
			Message:    handleError.Message,
		})
		return
	}

	// 5. if handler succeeds
	fmt.Printf("We are handling your request CLIENT!")
	body := writer.Bytes()
	responseHeaders := response.GetDefaultHeaders(0)
	responseHeaders.Replace("Content-Length", strconv.Itoa(len(body)))
	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, responseHeaders)
	conn.Write(body)
	conn.Close()
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
