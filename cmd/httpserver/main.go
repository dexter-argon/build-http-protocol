package main

import (
	"build-http-protocol/internal/headers"
	"build-http-protocol/internal/request"
	"build-http-protocol/internal/response"
	"build-http-protocol/internal/server"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

func toString(b []byte) string {
	out := ""
	for _, r := range b {
		out += fmt.Sprintf("%02x", r)
	}
	return out
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerError {
		body := request200()
		status := response.StatusOK
		h := response.GetDefaultHeaders(0)

		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = request400()
			status = response.StatusBadRequest
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = request500()
			status = response.StatusInternalServerError
		} else if req.RequestLine.RequestTarget == "/video" {
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(f)
			return nil
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			fmt.Printf("httpbin.org endpoint we're hitting %s\n", target[len("/httpbin/"):])
			if err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    err.Error(),
				}
			}
			h.Delete("Content-Length")
			h.Set("Transfer-Encoding", "chunked")
			h.Replace("Content-Type", "text/plain")
			h.Set("Trailer", "X-Content-SHA256")
			h.Set("Trailer", "X-Content-Length")
			w.WriteStatusLine(status)
			w.WriteHeaders(h)

			fullBody := []byte{}
			for {
				buf := make([]byte, 32)
				n, err := res.Body.Read(buf)
				if err != nil {
					break
				}
				w.WriteBody(fmt.Appendf([]byte{}, "%x\r\n", n))
				w.WriteBody(buf[:n])
				w.WriteBody([]byte("\r\n"))
				fullBody = append(fullBody, buf[:n]...)
			}
			w.WriteBody([]byte("0\r\n"))
			trailers := headers.NewHeaders()
			out := sha256.Sum256(fullBody)
			trailers.Set("X-Content-SHA256", toString(out[:]))
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
			w.WriteHeaders(trailers)
			w.WriteBody([]byte("\r\n"))
			return nil
		}

		err := w.WriteStatusLine(status)
		if err != nil {
			return newHandlerError(response.StatusInternalServerError, err.Error())
		}
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")
		err = w.WriteHeaders(h)
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
