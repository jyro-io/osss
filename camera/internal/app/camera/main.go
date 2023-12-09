package main

import (
	"bytes"
	"flag"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gopkg.in/yaml.v3"
	"image"
	"image/color"
	"io"
	"net"
	"os"
	"time"
)

type Config struct {
	LogLevel          string  `yaml:"logLevel"`
	MonitorAddress    string  `yaml:"monitorAddress"`
	Port              int     `yaml:"port"`
	CameraName        string  `yaml:"cameraName"`
	Threshold         float32 `yaml:"threshold"`
	MinimumMotionArea float64 `yaml:"minimumMotionArea"`
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

func detectMotion(config Config, cameraDevice int, motionChannel chan []byte) {
	sleepDuration := (1000 / 30) * time.Millisecond // 30 frames per second

	webcam, err := gocv.VideoCaptureDevice(cameraDevice)
	defer webcam.Close()
	if err != nil {
		log.Fatal("failed to open first video capture device: ", err)
	}

	img := gocv.NewMat()
	defer img.Close()
	imgDelta := gocv.NewMat()
	defer imgDelta.Close()
	imgThresh := gocv.NewMat()
	defer imgThresh.Close()
	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			time.Sleep(sleepDuration)
			continue
		}
		if img.Empty() {
			time.Sleep(sleepDuration)
			continue
		}

		statusColor := color.RGBA{0, 255, 0, 0}

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, &imgDelta)
		gocv.Threshold(imgDelta, &imgThresh, config.Threshold, 255, gocv.ThresholdBinary)
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.Dilate(imgThresh, &imgThresh, kernel)
		kernel.Close()
		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

		for i := 0; i < contours.Size(); i++ {
			area := gocv.ContourArea(contours.At(i))
			if area < config.MinimumMotionArea {
				continue
			}
			statusColor = color.RGBA{255, 0, 0, 0}
			gocv.DrawContours(&img, contours, i, statusColor, 2)
			rect := gocv.BoundingRect(contours.At(i))
			gocv.Rectangle(&img, rect, color.RGBA{0, 0, 255, 0}, 2)
		}
		contours.Close()

		gocv.PutText(&img, config.CameraName, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		encodedImg, err := gocv.IMEncode(".jpg", img)
		if err != nil {
			log.Error("failed to encode image: ", err)
		}
		motionChannel <- encodedImg.GetBytes()
		time.Sleep(sleepDuration)
	}
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("started")

	configFile := flag.String("config-file", "configs/config.yaml", "config file location")
	cameraDevice := flag.Int("camera-device", 0, "camera device")
	flag.Parse()
	log.Debug(*configFile)
	config := getConfig(*configFile)
	log.Info(config)

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal("invalid log level: ", err)
	}
	log.SetLevel(level)

	// open connection to monitor server
	monitorAddr := net.TCPAddr{
		IP:   net.ParseIP(config.MonitorAddress),
		Port: config.Port,
	}
	conn, err := net.Dial("tcp", monitorAddr.String())
	if err != nil {
		log.Fatal("failed to dial TCP: ", err)
	}
	defer conn.Close()
	log.Trace("connected to monitor on ", monitorAddr.String())

	motionChannel := make(chan []byte)
	// perpetual goroutine for the camera feed
	go detectMotion(config, *cameraDevice, motionChannel)
	// perpetual loop for sending data to monitor
	for {
		for data := range motionChannel {
			log.Trace("sending motion data to monitor: ", len(data)/1024, "KB")
			reader := bytes.NewReader(data)
			_, err = io.Copy(conn, reader)
			if err != nil && err != io.EOF {
				log.Error("failure while sending motion data: ", err)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
