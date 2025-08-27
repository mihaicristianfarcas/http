package response

import (
	"fmt"
	"http-from-scratch/internal/headers"
	"io"
)

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusNotFound            StatusCode = 404
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	b := []byte{}

	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})

	b = fmt.Append(b, "\r\n")
	_, err := w.Write(b)

	return err
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {

	statusLine := []byte{}

	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")

	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")

	case StatusNotFound:
		statusLine = []byte("HTTP/1.1 404 Not Found\r\n")

	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")

	default:
		return fmt.Errorf("unknown status code")
	}

	_, err := w.Write(statusLine)
	return err
}
