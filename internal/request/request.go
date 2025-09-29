package request

import (
	"build-http-protocol/internal/headers"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type parseState string

const (
	StateInit    parseState = "init"
	StateHeaders parseState = "headers"
	StateBody    parseState = "body"
	StateDone    parseState = "done"
	StateError   parseState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	state       parseState
}

func getInt(header *headers.Headers, key string, defaultValue int) int {
	valStr, exists := header.Get(key)

	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func (r *Request) error() bool {
	return r.state == StateError
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("request is in error state")
var CRLF = []byte("\r\n")
var SPACE = " "

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			read += n
			if n == 0 || !done {
				break outer
			}

			if done {
				r.state = StateBody
			}
		case StateBody:
			contentLength := getInt(r.Headers, "content-length", 0)
			if contentLength == 0 {
				r.state = StateDone
				break
			}

			bytesToRead := min(contentLength-len(r.Body), len(data[read:]))
			r.Body += string(data[read : read+bytesToRead])

			read += bytesToRead

			if len(r.Body) == contentLength {
				r.state = StateDone
			}
			break outer

		case StateDone:
			return read, nil
		default:
			panic("somehow we have programmed poorly")
		}
	}

	return read, nil
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, CRLF)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(CRLF)

	parts := bytes.Split(startLine, []byte(SPACE))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buf := make([]byte, 1024)
	bufIdx := 0
	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil {
			return nil, err
		}
		bufIdx += n
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}
		// why though? because it'll not read all the data available
		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}
	return request, nil
}
