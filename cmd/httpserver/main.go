package main

import (
	"build-http-protocol/internal/request"
	"build-http-protocol/internal/response"
	"build-http-protocol/internal/server"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func request400() []byte {
	return []byte(`<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>`)
}

func request500() []byte {
	return []byte(`<html> <head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>`)
}

func request200() []byte {
	return []byte(`<html><head><title>200 OK</title></head><body> <h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>`)
}

const port = 42069

func newHandlerError(statusCode response.StatusCode, message string) *server.HandlerError {
	return &server.HandlerError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerError {
		body := request200()
		status := response.StatusOK
		headers := response.GetDefaultHeaders(0)

		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = request400()
			status = response.StatusBadRequest
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = request500()
			status = response.StatusInternalServerError
		}

		err := w.WriteStatusLine(status)
		if err != nil {
			return newHandlerError(response.StatusInternalServerError, err.Error())
		}
		headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		headers.Replace("Content-Type", "text/html")
		err = w.WriteHeaders(headers)
		if err != nil {
			return newHandlerError(response.StatusInternalServerError, err.Error())
		}
		_, err = w.WriteBody(body)
		if err != nil {
			return newHandlerError(response.StatusInternalServerError, err.Error())
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
