package main

import (
	"flag"
	"io"
	"net"
	"os"

	"github.com/dhowden/raspicam"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel       string `yaml:"logLevel"`
	MonitorAddress string `yaml:"address"`
	Port           int    `yaml:"port"`
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
		// wait for movement
		v := raspicam.NewVid()
		errCh := make(chan error)
		go func() {
			for x := range errCh {
				log.Error(x)
			}
		}()

		log.Info("capturing video...")
		raspicam.Capture(v, conn, errCh)
		log.Info("done")
	}
}
