package main

import (
	"bytes"
	"flag"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"os"
	"time"
)

type Config struct {
	Debug       bool   `yaml:"debug"`
	LogLevel    string `yaml:"logLevel"`
	Address     string `yaml:"address"`
	CameraPort  int    `yaml:"cameraPort"`
	MonitorPort int    `yaml:"monitorPort"`
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
	log.Info(config)

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatalf("invalid log level: %s", err)
	}
	log.SetLevel(level)

	// set up camera listener server
	serverAddr := net.TCPAddr{
		IP:   net.IPv4zero,
		Port: config.CameraPort,
	}
	log.Info("started camera listener on ", serverAddr.String())
	cameraListener, err := net.Listen("tcp", serverAddr.String())
	if err != nil {
		log.Error(err)
		return
	}
	defer cameraListener.Close()

	// perpetual loop for accepting camera connections
	for {
		c, err := cameraListener.Accept()
		if err != nil && err != io.EOF {
			log.Fatal("failure accepting connection on camera listener: ", err)
		} else {
			// spawn goroutine for each camera
			go func(conn net.Conn) {
				log.Trace("serving camera connection ", conn.RemoteAddr())
				defer conn.Close()
				cameraAddress := conn.RemoteAddr().String()
				window := gocv.NewWindow(cameraAddress)
				defer window.Close()

				// perpetual loop to get motion data from camera connection
				for {
					buffer := make([]byte, 1024*100) // 100KB
					n, err := conn.Read(buffer)
					if err != nil {
						log.Error("error reading data: ", err)
					} else {
						if n > 0 {
							data := bytes.Trim(buffer, "\x00")
							log.Trace("received ", len(data)/1024, "KB from ", cameraAddress)

							motion, err := gocv.IMDecode(data, gocv.IMReadUnchanged)
							if err != nil || motion.Empty() {
								log.Error("failure while decoding bytes to matrix: ", err)
							} else {
								window.IMShow(motion)
								window.WaitKey(1)
								//write to mounted USB disk here
							}
						}
					}
					time.Sleep(10 * time.Millisecond)
				}
			}(c)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
