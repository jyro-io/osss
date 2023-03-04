package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	listener, err := net.DialTCP("tcp", "0.0.0.0:7777")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v", err)
		return
	}
	log.Println("Listening on 0.0.0.0:7777")
}
