package server

import (
	"fmt"
	"http-from-scratch/internal/request"
	"http-from-scratch/internal/response"
	"io"
	"net"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed  bool
	handler Handler
}

func handleConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, r)
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

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:  false,
		handler: handler,
	}

	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
