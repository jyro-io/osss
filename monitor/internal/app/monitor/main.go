package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel    string `yaml:"logLevel"`
	Address     string `yaml:"address"`
	CameraPort  int    `yaml:"cameraPort"`
	MonitorPort int    `yaml:"monitorPort"`
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

func addDataToCameraBuffer(data []byte, addr *net.UDPAddr, cameras []Camera) []Camera {
	for index, camera := range cameras {
		if camera.Address == addr {
			cameras[index].Buffer = append(cameras[index].Buffer, data...)
			log.Debug("added data to camera buffer: ", addr)
		}
	}
	return cameras
}

func findLongestCameraBufferIndex(cameras []Camera) int {
	longestIndex := -1
	longestBuffer := 0
	for index, camera := range cameras {
		if len(camera.Buffer) > longestBuffer {
			longestIndex = index
			longestBuffer = len(camera.Buffer)
		}
	}
	return longestIndex
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

	// set up camera listener
	serverAddr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: config.CameraPort,
	}
	cameraListener, err := net.ListenUDP("udp", &serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer cameraListener.Close()
	log.Info("started camera listener on ", serverAddr.String())

	// set up monitor feed
	monitorAddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: config.MonitorPort,
	}
	monitorFeed, err := net.DialUDP("udp", nil, &monitorAddr)
	if err != nil {
		log.Fatalf("failed to dial UDP: %s", err)
	}
	defer monitorFeed.Close()
	log.Info("started monitor feed on ", monitorAddr.String())

	cameraBuffer := make([]byte, 1024)
	cameras := []Camera{}
	longestBufferIndex := -1
	for {
		err = cameraListener.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		if err != nil {
			log.Fatal("error setting read deadline on camera listener: ", err)
		}
		n, addr, err := cameraListener.ReadFromUDP(cameraBuffer)
		if err != nil {
			log.Trace("error reading from camera listener: ", err)
			continue
		}
		if n > 0 {
			log.Debug(fmt.Sprintf("received %d bytes from %s", len(cameraBuffer[:n]), addr.String()))

			// maybe add new camera
			cameras = addCamera(addr, cameras)

			// write data to corresponding camera buffer
			cameras = addDataToCameraBuffer(cameraBuffer[:n], addr, cameras)

			// find longest camera buffer and write to monitor feed
			longestBufferIndex = findLongestCameraBufferIndex(cameras)
			if longestBufferIndex > -1 {
				log.Debug("writing data to monitor feed: ", cameras[longestBufferIndex].Buffer)
				n, err = monitorFeed.Write(cameras[longestBufferIndex].Buffer)
				if err != nil {
					log.Fatalf("failed to send data: %s", err)
				}
				log.Debug(fmt.Sprintf("sent %d bytes to monitor feed: %s", n, &monitorAddr))
				cameras[longestBufferIndex].Buffer = []byte{}
				longestBufferIndex = -1
			}
		}
	}
}
