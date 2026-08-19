package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	brokerHost := getenv("BROKER_HOST", "localhost")
	brokerPort := getenvInt("BROKER_PORT", 1883)

	// notebook banao — ye saare vehicles ka state sambhalega
	proc := NewProcessor()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", brokerHost, brokerPort))
	opts.SetClientID("ingest")
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)

	// jo bhi message aaye, processor ko bhejo — wahi 4 checks honge
	opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		proc.Handle(m.Payload())
	})

	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		log.Fatalf("mqtt connect failed: %v", t.Error())
	}
	defer client.Disconnect(250)
	log.Println("connected to MQTT broker")

	// wildcard: fleet/{region}/{device}/telemetry — listen all the devices

	if t := client.Subscribe("fleet/+/+/telemetry", 0, nil); t.Wait() && t.Error() != nil {
		log.Fatalf("mqtt subscribe failed: %v", t.Error())
	}
	log.Println("subscribed to fleet/+/+/telemetry")

	// graceful shutdown
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
