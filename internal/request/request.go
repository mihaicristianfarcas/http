package request

import (
	"bytes"
	"fmt"
	"http-from-scratch/internal/headers"
	"io"
	"strconv"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
	StateBody    parserState = "body"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
	Body        string
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "1.1"
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

var ErrBadRequestLine = fmt.Errorf("bad request line")
var ErrUnsupportedHTTPVersion = fmt.Errorf("unsupported HTTP version")
var ErrRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func ParseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrBadRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrBadRequestLine
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	if !rl.ValidHTTP() {
		return nil, 0, ErrUnsupportedHTTPVersion
	}

	return rl, read, nil
}

func (r *Request) HasBody() bool {
	// TODO update method when doing chunked encoding
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

dance:
	for {
		currentData := data[read:]

		if len(currentData) == 0 {
			break dance
		}

		switch r.state {
		case StateError:
			return 0, ErrRequestInErrorState
		case StateInit:
			rl, n, err := ParseRequestLine((data[read:]))
			if err != nil {
				r.state = StateError
			}

			if n == 0 {
				break dance
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break dance
			}

			read += n

			// Normally we would not get an EOF after reading data
			// so this is a workaround/transition
			if done {
				if r.HasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				panic("chunked not implemented")
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break dance

		default:
			panic("certified bad programming moment")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: buffer could get overrun; e.g. header/body that exceeds 1k
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		// TODO handle error
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
