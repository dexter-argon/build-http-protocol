package response

import (
	"build-http-protocol/internal/headers"
	"fmt"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const HTTP_VERSION = "HTTP/1.1"

type Response struct {
}

type WriterState string

const (
	StateStatusCode WriterState = "StatusCode"
	StateHeaders    WriterState = "Headers"
	StateBody       WriterState = "Body"
)

type Writer struct {
	writerState WriterState
	conn        io.Writer
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		writerState: StateStatusCode,
		conn:        conn,
	}
}

func (w *Writer) WriteToResponse(b []byte) (int, error) {
	err := w.WriteStatusLine(StatusOK)
	if err != nil {
		return 0, err
	}
	err = w.WriteHeaders(GetDefaultHeaders(len(b)))
	if err != nil {
		return 0, err
	}
	return w.write(b)
}

func (w *Writer) write(b []byte) (int, error) {
	return w.conn.Write(b)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != StateStatusCode {
		return fmt.Errorf("invalid Writer State for writing StatusLine")
	}

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

	_, err := w.write(statusLine)
	if err == nil {
		w.writerState = StateHeaders
	}

	return err
}

func (w *Writer) WriteHeaders(headers *headers.Headers) error {
	if w.writerState != StateHeaders {
		return fmt.Errorf("invalid writer state for writing headers")
	}
	var err error = nil
	var bytes []byte = []byte{}
	headers.ForEach(func(n, v string) {
		bytes = fmt.Appendf(bytes, "%s: %s\r\n", n, v)
	})
	bytes = fmt.Append(bytes, "\r\n")
	_, err = w.write(bytes)

	if err == nil {
		w.writerState = StateBody
	}
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != StateBody {
		return 0, fmt.Errorf("invalid writer state for writing body")
	}

	return w.write(p)
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}
