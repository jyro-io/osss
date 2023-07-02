package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel       string `yaml:"logLevel"`
	MonitorAddress string `yaml:"monitorAddress"`
	Port           int    `yaml:"port"`
	VideoFormat    string `yaml:"videoFormat"`
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

	// open connection to monitor server
	monitorAddr := net.TCPAddr{
		IP:   net.ParseIP(config.MonitorAddress),
		Port: config.Port,
	}
	monitor, err := net.Dial("tcp", monitorAddr.String())
	if err != nil {
		log.Fatalf("failed to dial TCP: %s", err)
	}
	defer monitor.Close()
	log.Info("connected to monitor on ", monitorAddr.String())

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error getting current working directory:", err)
		return
	}

	for {
		files, err := os.ReadDir(cwd)
		if err != nil {
			log.Error("error reading directory:", err)
			return
		}
		for _, file := range files {
			buffer := make([]byte, 0)
			if strings.Contains(file.Name(), config.VideoFormat) {
				fileHandle, err := os.Open(file.Name())
				if err != nil {
					log.Error("error opening file:", err)
					return
				}
				defer fileHandle.Close()

				n, err := fileHandle.Read(buffer)
				if err != nil && err != io.EOF {
					log.Error("error reading from file: ", err)
				}
				log.Debug(fmt.Sprintf("read %d bytes from stream: %s", n, buffer))

				n, err = monitor.Write(buffer)
				if err != nil {
					log.Fatalf("failed to send data: %s", err)
				}
				log.Debug(fmt.Sprintf("sent %d bytes to monitor feed: %s", n, &monitorAddr))
			}
		}
	}
}
