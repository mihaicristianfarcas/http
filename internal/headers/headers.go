package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// a header must contain only:
// Uppercase letters: A-Z
// Lowercase letters: a-z
// Digits: 0-9
// Special characters: !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~
func isToken(str []byte) bool {
	tokenPattern := regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)
	return tokenPattern.Match(str)
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrBadFieldLine
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", ErrBadFieldName
	}

	return string(name), string(value), nil
}

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")

var ErrBadFieldLine = fmt.Errorf("bad field line")
var ErrBadFieldName = fmt.Errorf("bad field name")
var ErrBadHeaderName = fmt.Errorf("bad header name")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0

	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		// Empty header
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, ErrBadHeaderName
		}

		read += idx + len(rn)
		h.Set(name, value)
	}

	return read, done, nil

}
