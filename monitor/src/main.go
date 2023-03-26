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

	// start camera stream listening server
	cameraListener, err := net.Listen("tcp", "0.0.0.0:7776")
	if err != nil {
		logError(err)
		return
	}
	log.Println("osss: monitor: started camera listener on 0.0.0.0:7776")

	// start localhost camera stream monitoring server
	monitorListener, err := net.Listen("tcp", "0.0.0.0:7777")
	if err != nil {
		logError(err)
		return
	}
	log.Println("osss: monitor: started camera listener on 0.0.0.0:7777")

	for {
		conn, err := cameraListener.Accept()
		if err != nil {
			logError(err)
			return
		}
		conn, err = monitorListener.Accept()
		if err != nil {
			logError(err)
			return
		}
		log.Printf("osss: monitor: accepted connection from: %v\n", conn.RemoteAddr())
		// watch camera streams for data
		// switch localhost:7777 stream to most recently active camera
	}
}
