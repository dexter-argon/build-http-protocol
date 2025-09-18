package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var CRLF []byte = []byte("\r\n")
var ERROR_MALFORMED_FIELD_LINE error = fmt.Errorf("malformed field line")
var ERROR_MALFORMED_FIELD_NAME error = fmt.Errorf("malformed field name")

func isToken(str string) bool {
	result := true
outer:
	for _, ch := range str {
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			result = true
			continue
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			result = true
		default:
			result = false
			break outer
		}
	}
	return result && len(str) > 0
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(fieldName string) (string, bool) {
	value, ok := h.headers[strings.ToLower(fieldName)]
	return value, ok
}

func (h *Headers) Set(fieldName, fieldValue string) {
	fieldName = strings.ToLower(fieldName)
	value, ok := h.headers[fieldName]
	if !ok {
		h.headers[fieldName] = fieldValue
		return
	}
	h.headers[fieldName] = fmt.Sprintf("%s,%s", value, fieldValue)
}

func (h *Headers) ForEach(cb func(u, v string)) {
	for k, v := range h.headers {
		cb(k, v)
	}
}

func (h Headers) parseHeader(fieldLine []byte) (string, string, error) {
	fields := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(fields) != 2 {
		return "", "", ERROR_MALFORMED_FIELD_LINE
	}

	key := fields[0]
	value := bytes.TrimSpace(fields[1])
	if bytes.HasSuffix(key, []byte(" ")) {
		return "", "", ERROR_MALFORMED_FIELD_NAME
	}

	return string(key), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data, CRLF)
		if idx == -1 {
			break
		}

		// empty header
		if idx == 0 {
			done = true
			read += idx + len(CRLF)
			break
		}

		fieldName, fieldValue, err := h.parseHeader(data[:idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken(fieldName) {
			return 0, false, ERROR_MALFORMED_FIELD_NAME
		}

		h.Set(fieldName, fieldValue)
		read += idx + len(CRLF)
		data = data[idx+len(CRLF):]
	}

	return read, done, nil
}
