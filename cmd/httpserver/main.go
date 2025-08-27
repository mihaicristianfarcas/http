package main

import (
	"http-from-scratch/internal/request"
	"http-from-scratch/internal/response"
	"http-from-scratch/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "skill issue\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "my bad G\n",
			}
		default:
			w.Write([]byte("all good ong\n"))
			return nil
		}
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
