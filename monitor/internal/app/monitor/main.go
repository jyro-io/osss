package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Monitor struct {
		LogLevel string `yaml:"logLevel"`
		Address  string `yaml:"address"`
		Port     int    `yaml:"port"`
	} `yaml:"monitor"`
}

func getConfig(file string) Config {
	handle, err := os.Open(file)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer handle.Close()

	content, err := io.ReadAll(handle)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}

	c := Config{}
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		log.Error(err)
	}

	return c
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("started")

	configFile := flag.String("config-file", "configs/config.yaml", "config file location")
	flag.Parse()
	log.Info(*configFile)
	config := getConfig(*configFile)
	log.Info(config)

	level, err := log.ParseLevel(config.Monitor.LogLevel)
	if err != nil {
		log.Fatalf("invalid log level: %s", err)
	}
	log.SetLevel(level)

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", config.Monitor.Address, strconv.Itoa(config.Monitor.Port)))
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	cameraBuffer := make([]byte, 1024)
	log.Info(fmt.Sprintf("started camera listener on %s", serverAddr.String()))

	for {
		n, addr, err := conn.ReadFromUDP(cameraBuffer)
		if err != nil {
			log.Error(fmt.Sprintf("error reading from UDP: %s", err.Error()))
			continue
		}
		log.Debug(fmt.Sprintf("received %d bytes from %s", n, addr.String()))
		// watch camera streams for data
		// switch localhost:7777 stream to most recently active camera
	}
}
