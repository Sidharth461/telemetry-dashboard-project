package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var proc *Processor

func main() {
	brokerHost := getenv("BROKER_HOST", "localhost")
	brokerPort := getenvInt("BROKER_PORT", 1883)

	// create the processor — it keeps state for all vehicles
	proc = NewProcessor(60*time.Second, 15*time.Second)
	hub = NewHub(proc) // send data in 1s to browser
	go hub.Run()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", brokerHost, brokerPort))
	opts.SetClientID("ingest")
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)

	// send every incoming message to the processor — it runs the fault checks
	opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		proc.Handle(m.Payload()) // whatever message bytes give it to the processor
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

	// check every second — which vehicle has been silent for 15s → OFFLINE
	go func() {
		for {
			proc.Sweep()
			time.Sleep(1 * time.Second)
		}
	}()
	log.Println("HTTP server on :8080")
	err := http.ListenAndServe(":8080", http.HandlerFunc(handleHTTP))
	if err != nil {
		log.Fatalf("http server failed: %v", err)
	}

	// graceful shutdown (SIGTERM = docker stop, SIGINT = ctrl+c)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
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
