package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"os"
	"sync"
	"time"
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
	Buffer  []gocv.Mat
}

func addCamera(addr net.Addr, cameras []Camera) []Camera {
	for _, camera := range cameras {
		if camera.Address == addr {
			return cameras
		}
	}
	cameras = append(cameras, Camera{Address: addr, Buffer: []gocv.Mat{}})
	log.Debug("added camera: ", addr)
	return cameras
}

func addDataToCameraBuffer(data gocv.Mat, addr net.Addr, cameras []Camera) []Camera {
	for index, camera := range cameras {
		if camera.Address == addr {
			cameras[index].Buffer = append(cameras[index].Buffer, data)
			log.Debug("added data to camera buffer: ", addr)
		}
	}
	return cameras
}

func receiveMotion(conn net.Conn) (gocv.Mat, error) {
	decoder := json.NewDecoder(conn)
	var m gocv.Mat
	err := decoder.Decode(&m)
	return m, err
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

	// set up monitor live feed
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

	var cameras []Camera
	var cameraMutex sync.Mutex
	var cameraRoutines = make(chan int)
	window := gocv.NewWindow("Camera Monitor")
	defer window.Close()
	// call perpetual goroutine that flushes camera buffers
	go func() {
		for {
			// wait for all goroutines to finish,
			// then flush camera buffers
			numGoroutines := 0
			for diff := range cameraRoutines {
				numGoroutines += diff
				if numGoroutines == 0 {
					log.Debug("flushing camera buffers: ", cameras)
					// flush camera buffers to monitorFeed
					for index, camera := range cameras {
						if len(camera.Buffer) > 0 {
							for _, motion := range camera.Buffer {
								defer motion.Close()
								window.IMShow(motion)
								// write to mounted USB disk here
							}
						}
						cameras[index].Buffer = []gocv.Mat{}
					}
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	// perpetual loop for accepting camera connections
	for {
		c, err := cameraListener.Accept()
		if err != nil && err != io.EOF {
			log.Fatal("failure accepting connection on camera listener: ", err)
		} else {
			cameraRoutines <- +1
			go func(connection net.Conn) {
				log.Info(fmt.Sprintf("serving %s", connection.RemoteAddr().String()))
				defer connection.Close()
				img, err := receiveMotion(connection)
				netData, err := bufio.NewReader(connection).ReadString('\n')
				if err != nil && err != io.EOF {
					log.Fatal("failure while reading data: ", err)
				} else {
					if len(netData) > 0 {
						log.Debug("received from client: ", string(netData))
						cameraMutex.Lock()
						// maybe add new camera
						cameras = addCamera(connection.RemoteAddr(), cameras)
						// write data to corresponding camera buffer
						cameras = addDataToCameraBuffer(img, connection.RemoteAddr(), cameras)
						cameraMutex.Unlock()
					}
				}
				cameraRoutines <- -1
			}(c)
		}
	}
}
