package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

func logError(error error) {
	log.Println("osss: monitor: error:", error)
}

func logLine(message string) {
	log.Println(fmt.Sprintf("osss: monitor: %s", message))
}

type Config struct {
	Monitor struct {
		cameraPort  int
		monitorPort int
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
	logLine("started")
	config := getConfig()
	logLine("loaded config.yaml")
	spew.Dump(config)

	// start camera stream listening server
	cameraEndpoint := fmt.Sprintf("0.0.0.0:%s", strconv.Itoa(config.Monitor.cameraPort))
	cameraListener, err := net.Listen("tcp", cameraEndpoint)
	if err != nil {
		logError(err)
		return
	}
	logLine(fmt.Sprintf("started camera listener on %s", cameraEndpoint))

	// start localhost camera stream monitoring server
	monitorEndpoint := fmt.Sprintf("0.0.0.0:%s", strconv.Itoa(config.Monitor.cameraPort))
	monitorListener, err := net.Listen("tcp", monitorEndpoint)
	if err != nil {
		logError(err)
		return
	}
	logLine(fmt.Sprintf("started monitor listener on %s", monitorEndpoint))

	for {
		conn, err := cameraListener.Accept()
		if err != nil {
			logError(err)
			return
		}
		logLine(fmt.Sprintf("accepted camera connection from: %s", conn.RemoteAddr()))
		conn, err = monitorListener.Accept()
		if err != nil {
			logError(err)
			return
		}
		logLine(fmt.Sprintf("accepted monitor connection from: %s", conn.RemoteAddr()))
		// watch camera streams for data
		// switch localhost:7777 stream to most recently active camera
	}
}
