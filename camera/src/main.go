package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/dhowden/raspicam"
	"gopkg.in/yaml.v2"
)

func logError(error error) {
	fmt.Println("osss: camera: error:", error)
}

type Config struct {
	Camera struct {
	}
}

func getConfig() Config {
	c := Config{}
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logError(err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logError(err)
	}
	return c
}

func main() {
	fmt.Println("osss: camera: info: started")
	config := getConfig()
	fmt.Println("osss: camera: info: loaded config.yaml:")
	spew.Dump(config)

	// scan local network for monitor listening on port 7777
	client, err := net.DialTCP("tcp", "0.0.0.0:7777")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v", err)
		return
	}
	log.Println("Listening on 0.0.0.0:7777")

	for {
		// wait for IR to detect movement
		// take video for configured time
		// upload video to monitor
		dummy_video_stream := []byte("videooo")
		status, err := client.Write(dummy_video_stream)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v", err)
			return
		}
		go func() {
			s := raspicam.Vid()
			errCh := make(chan error)
			go func() {
				for x := range errCh {
					fmt.Fprintf(os.Stderr, "%v\n", x)
				}
			}()
			log.Println("Capturing video...")
			raspicam.Capture(s, conn, errCh)
			log.Println("Done")
			conn.Close()
		}()
	}
}
