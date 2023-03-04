package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/dhowden/raspicam"
	"gopkg.in/yaml.v2"
)

func logError(error error) {
	log.Println("osss: camera: error:", error)
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
	log.Println("osss: camera: info: started")
	config := getConfig()
	log.Println("osss: camera: info: loaded config.yaml:")
	spew.Dump(config)

	// scan local network for monitor listening on port 7777
	conn, err := net.Dial("tcp", "192.168.1.100:7777")
	if err != nil {
		logError(err)
		return
	}

	for {
		// wait for IR to detect movement

		v := raspicam.NewVid()
		errCh := make(chan error)
		go func() {
			for x := range errCh {
				logError(x)
			}
		}()

		log.Println("osss: camera: capturing video...")
		raspicam.Capture(v, conn, errCh)
		log.Println("osss: camera: done")
	}

	conn.Close()
}
