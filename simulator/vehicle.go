package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	StateMoving   = "MOVING"
	StateIdle     = "IDLE"
	StateCharging = "CHARGING"
	StateOff      = "OFF"
)

type Telemetry struct {
	DeviceID       string   `json:"device-Id"`
	TS             string   `json:"ts"`
	State          string   `json:"state"`
	SpeedKph       float64  `json:"speedKph"`
	BatteryPct     float64  `json:"batteryPct"`
	OdometerMeters float64  `json:"odometerMeters"`
	RouteID        string   `json:"routeId"`
	Faults         []string `json:"faults"`
}

type Vehicle struct {
	id             string
	region         string
	state          string
	dwellUntil     int64
	speedKph       float64
	batteryPct     float64
	odometerMeters float64
	delayed        *Telemetry
	rng            *rand.Rand
}

func NewVehicle(region, id string) *Vehicle {
	now := time.Now().UnixMilli()
	// Each Vehicle has its own random seed - And each Behaves Different
	seed := time.Now().UnixNano() + int64(id[4])
	return &Vehicle{
		id:             id,
		region:         region,
		state:          StateIdle,
		dwellUntil:     now,
		batteryPct:     40 + rand.Float64()*60,
		odometerMeters: 100000 + rand.Float64()*200000,
		rng:            rand.New(rand.NewSource(seed)),
	}
}

func (v *Vehicle) step() {
	switch v.state {
	case StateMoving:
		v.speedKph = 30 + v.rng.Float64()*90
		v.batteryPct -= 0.05 + v.rng.Float64()*0.2
		v.odometerMeters += v.speedKph * (1000.0 / 3600.0) * 0.2
	case StateIdle:
		v.speedKph = 0
		v.batteryPct -= 0.01
	case StateCharging:
		v.speedKph = 0
		v.batteryPct += 0.4 + v.rng.Float64()
	case StateOff:
		v.speedKph = 0
	}
	v.batteryPct = max(0, min(100, v.batteryPct))
	v.maybeTransition()
}

func (v *Vehicle) maybeTransition() {
	if time.Now().UnixMilli() < v.dwellUntil {
		return
	}

	r := v.rng.Float64()
	switch v.state {
	case StateMoving:
		if r < 0.2 {
			v.state = StateIdle
			v.dwellUntil = time.Now().UnixMilli() + 3000 // 3 sec idle
		}
	case StateIdle:
		if v.batteryPct < 20 && r < 0.5 {
			v.state = StateCharging
			v.dwellUntil = time.Now().UnixMilli() + 5000
		} else if r < 0.2 {
			v.state = StateMoving
			v.dwellUntil = time.Now().UnixMilli() + 8000 // 8 sec move
		}
	case StateCharging:
		if v.batteryPct > 95 || r < 0.1 {
			v.state = StateMoving
			v.dwellUntil = time.Now().UnixMilli() + 8000
		}
	case StateOff:
		if r < 0.1 {
			v.state = StateMoving
			v.dwellUntil = time.Now().UnixMilli() + 8000
		}
	}
}

func (v *Vehicle) publish(client mqtt.Client, m *Telemetry) {
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	topic := fmt.Sprintf("fleet/%s/%s/telemetry", v.region, v.id)
	client.Publish(topic, 0, false, data)
}

func (v *Vehicle) Run(client mqtt.Client, interval time.Duration) {
	for {
		time.Sleep(interval) //200ms ruko (interval = 1s/5Hz)

		v.step() // state machine update

		msg := &Telemetry{
			DeviceID:       v.id,
			TS:             time.Now().Format("2006-01-02T15:04:05.000Z"),
			State:          v.state,
			SpeedKph:       v.speedKph,
			BatteryPct:     v.batteryPct,
			OdometerMeters: v.odometerMeters,
			RouteID:        fmt.Sprintf("R-%03d", v.rng.Intn(300)),
			Faults:         []string{},
		}
		if v.delayed != nil {
			v.publish(client, v.delayed)
			v.delayed = nil
		}
		v.publish(client, msg)

		// FAULT: duplicate delivery (1%) — same message published twice
		if v.rng.Float64() < 0.01 {
			v.publish(client, msg)
		}
		if v.rng.Float64() < 0.02 {
			v.publish(client, msg)
		}
	}
}
