package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"

	"github.com/jessfraz/netscan/pkg/scanner"
	"gopkg.in/yaml.v2"
)

func logError(error error) {
	log.Println("osss: monitor: error: ", error)
}

func logLine(message string) {
	log.Println(fmt.Sprintf("osss: monitor: %s", message))
}

type Config struct {
	Monitor struct {
		address string `yaml:"address"`
		port    int    `yaml:"port"`
	} `yaml:"monitor"`
}

func getConfig(file string) Config {
	c := Config{}
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		logError(err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logError(err)
	}
	return c
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func main() {
	configFile := flag.String("config-file", "configs/config.yaml", "config file location")
	flag.Parse()
	logLine(*configFile)

	logLine("started")
	config := getConfig(*configFile)
	logLine("loaded config.yaml")

	networkScanner := scanner.NewScanner(
		scanner.WithPorts([]int{config.Monitor.port}),
	)
	cameraAddresses := networkScanner.Scan()

	cameraAddress := fmt.Sprintf(":%s", strconv.Itoa(config.Monitor.port))
	serverAddr, err := net.ResolveUDPAddr("udp", cameraAddress)
	if err != nil {
		logError(err)
	}
	cameraConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		logError(err)
	}
	cameraBuffer := make([]byte, 1024)
	logLine(fmt.Sprintf("started camera listener on %s", cameraAddress))

	// start localhost camera stream monitoring server
	monitorAddress := fmt.Sprintf("0.0.0.0:%s", strconv.Itoa(config.Monitor.port))
	monitorListener, err := net.Listen("udp", monitorAddress)
	if err != nil {
		logError(err)
	}
	logLine(fmt.Sprintf("started monitor listener on %s", monitorAddress))

	for {
		n, addr, err := cameraConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Error reading from UDP: %s", err.Error())
			continue
		}
		fmt.Printf("Received %d bytes from %s: %s\n", n, addr.String(), string(buf[:n]))
		_, err = cameraConn.WriteToUDP(buf[:n], addr)
		if err != nil {
			fmt.Printf("Error writing to UDP: %s", err.Error())
			continue
		}
		// watch camera streams for data
		// switch localhost:7777 stream to most recently active camera
	}
}
