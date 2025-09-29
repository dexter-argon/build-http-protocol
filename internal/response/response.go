package response

import (
	"build-http-protocol/internal/headers"
	"fmt"
	"io"
	"strconv"
)

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const HTTP_VERSION = "HTTP/1.1"

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	fmt.Println("inside the WriteStatusLine")
	statusLine := []byte{}

	switch statusCode {
	case StatusOK:
		statusLine = []byte(HTTP_VERSION + " 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte(HTTP_VERSION + " 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte(HTTP_VERSION + " 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("unrecognized status code")
	}

	_, err := w.Write(statusLine)

	return err
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	var err error = nil
	var bytes []byte = []byte{}
	h.ForEach(func(n, v string) {
		bytes = fmt.Appendf(bytes, "%s: %s\r\n", n, v)
	})
	bytes = fmt.Append(bytes, "\r\n")
	_, err = w.Write(bytes)
	return err
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}
