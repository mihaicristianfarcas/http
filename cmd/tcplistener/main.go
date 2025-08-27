package main

import (
	"fmt"
	"http-from-scratch/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("error", "error", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Printf("request line:\n")
		fmt.Printf("- method: %s\n", r.RequestLine.Method)
		fmt.Printf("- target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- ver: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("headers:\n")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})
		fmt.Printf("body:\n")
		fmt.Printf("%s\n", r.Body)
	}
}
