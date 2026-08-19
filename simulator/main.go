package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

type SimulatorConfig struct {
	BrokerHost       string
	BrokerPort       int
	Region           string
	NumberDevices    int
	PublishHz        int
	DuplicateRate    float64
	OutOfOrderRate   float64
	CounterResetRate float64
	SilentRate       float64
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using defaults")
	}
	brokerHost := getenv("BROKER_HOST", "localhost")
	brokerPort := getenvInt("BROKER_PORT", 1883)

	region := getenv("REGION", "west-01")
	numDevices := getenvInt("NUM_DEVICES", 20)
	publishHz := getenvInt("PUBLISH_HZ", 5)

	dupRate := getenvFloat("DUP_RATE", 0.01)
	outofRate := getenvFloat("OUT_OF_ORDER_RATE", 0.02)
	resetRate := getenvFloat("RESET_RATE", 0.001)
	silentRate := getenvFloat("SILENT_RATE", 0.005)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", brokerHost, brokerPort))
	opts.SetClientID("simulator")
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)

	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		log.Fatalf("mqtt connect failed: %v", t.Error())
	}
	defer client.Disconnect(250)
	log.Println("connected to MQTT broker")

	log.Printf("starting %d vehicles at %d Hz (region %s)", numDevices, publishHz, region)

	interval := time.Second / time.Duration(publishHz)
	for i := 0; i < numDevices; i++ {
		v := NewVehicle(region, fmt.Sprintf("VEH-%04d", i+1), dupRate, outofRate, resetRate, silentRate)
		go v.Run(client, interval)
	}

	// select {}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("shutting down...")
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func getenvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	}
	return def
}
