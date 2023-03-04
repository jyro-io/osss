package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/dhowden/raspicam"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:7777")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v", err)
		return
	}
	log.Println("Listening on 0.0.0.0:7777")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "accept: %v", err)
			return
		}
		log.Printf("Accepted connection from: %v\n", conn.RemoteAddr())
		go func() {
			s := raspicam.Vid()
			errCh := make(chan error)
			go func() {
				for x := range errCh {
					fmt.Fprintf(os.Stderr, "%v\n", x)
				}
			}()
			log.Println("Capturing image...")
			raspicam.Capture(s, conn, errCh)
			log.Println("Done")
			conn.Close()
		}()
	}
}
