package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gopkg.in/yaml.v3"
	"image"
	"image/color"
	"io"
	"net"
	"os"
)

type Config struct {
	LogLevel       string `yaml:"logLevel"`
	MonitorAddress string `yaml:"monitorAddress"`
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

func detectMotion() gocv.Mat {
	webcam, _ := gocv.VideoCaptureDevice(0) // Open the first camera device
	defer webcam.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	status := "Ready"

	for {
		img := gocv.NewMat()
		defer img.Close()

		if ok := webcam.Read(&img); !ok {
			continue
		}
		if img.Empty() {
			continue
		}

		status = "Ready"
		statusColor := color.RGBA{0, 255, 0, 0}

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, &imgDelta)

		// remaining cleanup of the image to use for finding contours.
		// first use threshold
		gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

		// then dilate
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.Dilate(imgThresh, &imgThresh, kernel)

		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		for i := 0; i < contours.Size(); i++ {
			area := gocv.ContourArea(contours.At(i))
			if area < 3000.0 {
				continue
			}

			status = "Motion detected"
			statusColor = color.RGBA{255, 0, 0, 0}
			rect := gocv.BoundingRect(contours.At(i))
			gocv.Rectangle(&img, rect, color.RGBA{255, 0, 0, 0}, 2)
		}

		gocv.PutText(&img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		return img
	}
}

type Message struct {
	Name string
	Body string
	Time int64
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

	for {
		motion := detectMotion()
		_, err := monitor.Write(motion.ToBytes())
		if err != nil {
			log.Fatalf("failure while sending motion data: %s", err)
		}
	}
}
