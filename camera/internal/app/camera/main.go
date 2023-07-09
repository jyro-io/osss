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
	"time"
)

type Config struct {
	Debug             bool    `yaml:"debug"`
	LogLevel          string  `yaml:"logLevel"`
	MonitorAddress    string  `yaml:"monitorAddress"`
	Port              int     `yaml:"port"`
	CameraName        string  `yaml:"cameraName"`
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

func detectMotion(config Config, webcam *gocv.VideoCapture) []byte {
	img := gocv.NewMat()
	defer img.Close()
	imgDelta := gocv.NewMat()
	defer imgDelta.Close()
	imgThresh := gocv.NewMat()
	defer imgThresh.Close()
	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()
	foundMotion := false

	for {
		if ok := webcam.Read(&img); !ok {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if img.Empty() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		statusColor := color.RGBA{0, 255, 0, 0}

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, &imgDelta)

		// remaining cleanup of the image to use for finding contours.
		// first use threshold
		gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

		// then dilate
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.Dilate(imgThresh, &imgThresh, kernel)
		kernel.Close()

		// now find contours
		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

		for i := 0; i < contours.Size(); i++ {
			area := gocv.ContourArea(contours.At(i))
			if area >= config.MinimumMotionArea {
				foundMotion = true
				statusColor = color.RGBA{255, 0, 0, 0}
				gocv.DrawContours(&img, contours, i, statusColor, 2)
				rect := gocv.BoundingRect(contours.At(i))
				gocv.Rectangle(&img, rect, color.RGBA{0, 0, 255, 0}, 2)
			}
		}
		contours.Close()

		gocv.PutText(&img, config.CameraName, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		if foundMotion {
			break
		} else {
			img = gocv.NewMat()
			imgDelta = gocv.NewMat()
			imgThresh = gocv.NewMat()
			mog2 = gocv.NewBackgroundSubtractorMOG2()
			continue
		}
	}

	if config.Debug {
		window := gocv.NewWindow("Camera Debug Monitor")
		window.IMShow(img)
		window.WaitKey(1000)
		err := window.Close()
		if err != nil {
			log.Fatal("failed to close debug window: ", err)
		}
	}

	return img.ToBytes()
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
		log.Fatal("invalid log level: ", err)
	}
	log.SetLevel(level)

	// open connection to monitor server
	monitorAddr := net.TCPAddr{
		IP:   net.ParseIP(config.MonitorAddress),
		Port: config.Port,
	}
	monitor, err := net.Dial("tcp", monitorAddr.String())
	if err != nil {
		log.Fatal("failed to dial TCP: ", err)
	}
	defer monitor.Close()
	log.Info("connected to monitor on ", monitorAddr.String())

	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		log.Fatal("failed to open first video capture device: ", err)
	}
	defer webcam.Close()

	for {
		motion := detectMotion(config, webcam)
		log.Debug("sending motion event data to monitor: ", len(motion)/1024, "KB")
		_, err := monitor.Write(motion)
		if err != nil && err != io.EOF {
			log.Fatal("failure while sending motion event data: ", err)
		}
	}
}
