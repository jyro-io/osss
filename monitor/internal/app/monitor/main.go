package main

import (
	"bufio"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"os"
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

type Camera struct {
	Address net.Addr
	Buffer  []byte
}

func addCamera(addr net.Addr, cameras []Camera) []Camera {
	for _, camera := range cameras {
		if camera.Address == addr {
			return cameras
		}
	}
	cameras = append(cameras, Camera{Address: addr, Buffer: make([]byte, 1024)})
	log.Debug("added camera: ", addr)
	return cameras
}

func addDataToCameraBuffer(data []byte, addr net.Addr, cameras []Camera) []Camera {
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

func handleConnection(connection net.Conn, cameras []Camera, monitorFeed *net.UDPConn) {
	log.Info(fmt.Sprintf("serving %s", connection.RemoteAddr().String()))
	longestBufferIndex := -1
	for {
		netData, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			log.Trace("failure while reading data: ", err)
		}
		if len(netData) > 0 {
			log.Debug("received from client: ", string(netData))

			// maybe add new camera
			cameras = addCamera(connection.RemoteAddr(), cameras)

			// write data to corresponding camera buffer
			cameras = addDataToCameraBuffer([]byte(netData), connection.RemoteAddr(), cameras)

			// find the longest camera buffer and write to monitor feed
			longestBufferIndex = findLongestCameraBufferIndex(cameras)
			if longestBufferIndex > -1 {
				log.Debug("writing data to monitor feed: ", cameras[longestBufferIndex].Buffer)
				n, err := monitorFeed.Write(cameras[longestBufferIndex].Buffer)
				if err != nil {
					log.Fatalf("failed to send data: %s", err)
				}
				log.Debug(fmt.Sprintf("sent %d bytes to monitor feed", n))
				cameras[longestBufferIndex].Buffer = []byte{}
				longestBufferIndex = -1
			}
		}
	}
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

	// set up monitor feed
	monitorAddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: config.MonitorPort,
	}
	monitorFeed, err := net.DialUDP("udp", nil, &monitorAddr)
	if err != nil {
		log.Fatalf("failed to dial TCP: %s", err)
	}
	log.Info("started monitor feed on ", monitorAddr.String())

	cameras := []Camera{}
	for {
		c, err := cameraListener.Accept()
		if err != nil {
			log.Trace("failure while accepting camera connection: ", err)
			continue
		}
		go handleConnection(c, cameras, monitorFeed)
	}
}
