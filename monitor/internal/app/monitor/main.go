package main

import (
	"flag"
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

func addDataToCameraBuffer(motion gocv.Mat, addr net.Addr, cameras []Camera) []Camera {
	for index, camera := range cameras {
		if camera.Address == addr {
			cameras[index].Buffer = append(cameras[index].Buffer, motion)
			log.Debug("added data to camera buffer: ", addr)
		}
	}
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
					log.Debug("flushing camera buffers...")
					// flush camera buffers to monitor feed window
					for cIndex, camera := range cameras {
						log.Debug("flushing camera ", camera.Address)
						for mIndex, motion := range camera.Buffer {
							log.Debug("flushing motion event ", mIndex)
							window.IMShow(motion)
							window.WaitKey(1000)
							// write to mounted USB disk here
						}
						cameras[cIndex].Buffer = []gocv.Mat{}
					}
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	// perpetual loop for accepting camera connections
	for {
		c, err := cameraListener.Accept()
		if err != nil && err != io.EOF {
			log.Fatal("failure accepting connection on camera listener: ", err)
		} else {
			cameraRoutines <- +1
			go func(conn net.Conn) {
				log.Debug("serving camera connection ", conn.RemoteAddr())
				defer conn.Close()

				// get motion event from camera connection
				buffer := make([]byte, 0)
				for {
					_, err := conn.Read(buffer)
					if err != nil {
						log.Debug("error reading data:", err)
						break
					}
				}

				log.Debug("received data from camera: ", len(buffer), " bytes")

				motion := gocv.Mat{}
				err := gocv.IMDecodeIntoMat(buffer, gocv.IMReadUnchanged, &motion)
				if err != nil {
					log.Error("failure while decoding bytes: ", err)
				} else {
					cameraMutex.Lock()
					// maybe add new camera
					cameras = addCamera(conn.RemoteAddr(), cameras)
					// write data to corresponding camera buffer
					cameras = addDataToCameraBuffer(motion, conn.RemoteAddr(), cameras)
					cameraMutex.Unlock()
				}
				cameraRoutines <- -1
			}(c)
		}
	}
}
