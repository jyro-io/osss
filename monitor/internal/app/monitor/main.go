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
	LogLevel string `yaml:"logLevel"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
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

type Camera struct {
	Address *net.UDPAddr
	Buffer  []byte
}

func addCamera(addr *net.UDPAddr, cameras []Camera) []Camera {
	for _, camera := range cameras {
		if camera.Address == addr {
			return cameras
		}
	}
	cameras = append(cameras, Camera{Address: addr, Buffer: make([]byte, 1024)})
	log.Debug("added camera: ", addr)
	return cameras
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("started")

	configFile := flag.String("config-file", "configs/config.yaml", "config file location")
	flag.Parse()
	log.Debug(*configFile)
	config := getConfig(*configFile)
	log.Debug(config)

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatalf("invalid log level: %s", err)
	}
	log.SetLevel(level)

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", config.Address, strconv.Itoa(config.Port)))
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Info("started camera listener on ", serverAddr.String())

	netBuffer := make([]byte, 1024*1024) // 1GB buffer
	cameras := []Camera{}
	for {
		n, addr, err := conn.ReadFromUDP(netBuffer)
		if err != nil {
			log.Error("error reading from UDP: ", err.Error())
			continue
		}
		addCamera(addr, cameras)

		go func(data []byte, addr *net.UDPAddr) {
			log.Debug(fmt.Sprintf("received %d bytes from %s", len(data), addr.String()))
		}(netBuffer[:n], addr)

		// switch localhost:7777 stream to most recently active camera
	}
}
