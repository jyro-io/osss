package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel       string `yaml:"logLevel"`
	MonitorAddress string `yaml:"address"`
	Port           int    `yaml:"port"`
	Stream         string `yaml:"stream"`
}

func getConfig(file string) Config {
	handle, err := os.Open(file)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer handle.Close()

	content, err := io.ReadAll(handle)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
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
	log.Debug(*configFile)
	config := getConfig(*configFile)
	log.Debug(config)

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatalf("invalid log level: %s", err)
	}
	log.SetLevel(level)

	// set up monitor server
	monitorAddr := net.UDPAddr{
		IP:   net.ParseIP(config.MonitorAddress),
		Port: config.Port,
	}
	monitorListener, err := net.DialUDP("udp", nil, &monitorAddr)
	if err != nil {
		log.Fatalf("failed to dial UDP: %s", err)
	}
	defer monitorListener.Close()
	log.Info("connected to monitor on ", monitorAddr.String())

	for {
		file, err := os.Open(config.Stream)
		if err != nil {
			log.Error("error opening stream at ", config.Stream, " ", err)
		}
		defer file.Close()

		buffer := make([]byte, 4096)

		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Error("error reading from file: ", err)
		}
		log.Debug(fmt.Sprintf("read %d bytes from stream: %s", n, buffer))

		n, err = monitorListener.Write(buffer)
		if err != nil {
			log.Fatalf("failed to send data: %s", err)
		}
		log.Debug(fmt.Sprintf("sent %d bytes to monitor feed: %s", n, &monitorAddr))
	}
}
