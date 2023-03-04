package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

func logError(error error) {
	log.Println("osss: monitor: error:", error)
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
	log.Println("osss: monitor: info: started")
	config := getConfig()
	log.Println("osss: monitor: info: loaded config.yaml:")
	spew.Dump(config)

	// create wifi network from config
	listener, err := net.Listen("tcp", "0.0.0.0:7777")
	if err != nil {
		logError(err)
		return
	}
	log.Println("osss: monitor: listening on 0.0.0.0:7777")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logError(err)
			return
		}
		log.Printf("osss: monitor: accepted connection from: %v\n", conn.RemoteAddr())
		// accept video uploads from clients
	}
}
