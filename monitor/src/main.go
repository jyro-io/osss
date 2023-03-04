package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

func logError(error error) {
	fmt.Println("osss: monitor: error:", error)
}

type Config struct {
	Monitor struct {
	}
}

func getConfig() Config {
	c := Config{}
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logError(err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logError(err)
	}
	return c
}

func main() {
	fmt.Println("osss: monitor: info: started")
	config := getConfig()
	fmt.Println("osss: monitor: info: loaded config.yaml:")
	spew.Dump(config)
	// create wifi network from config
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
		// accept video uploads from clients
	}
}
