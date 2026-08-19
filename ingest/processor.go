package main

import (
	"encoding/json"
	"sync"
	"time"
)

type TelemetryMessage struct {
	DeviceID       string   `json:"deviceId"`
	Timestamp      string   `json:"ts"`    // message kab bheja gaya (ISO)
	State          string   `json:"state"` // MOVING/IDLE/CHARGING/OFF
	SpeedKph       float64  `json:"speedKph"`
	BatteryPct     float64  `json:"batteryPct"`
	OdometerMeters float64  `json:"odometerMeters"`
	RouteID        string   `json:"routeId"`
	Faults         []string `json:"faults"`
}

type VehicleState struct {
	DeviceID              string
	CurrentState          string
	CurrentSpeedKph       float64
	CurrentBatteryPct     float64
	CurrentOdometerMeters float64
	TotalDistanceMeters   float64   // service start se total distance
	LastTimestamp         string    // out-of-order check ke liye
	LastMessageTime       time.Time // silence/offline check ke liye
	IsOnline              bool
}
type Processor struct {
	mu                     sync.Mutex               // race prevention
	devicePages            map[string]*VehicleState // key = deviceId
	TotalMessagesReceived  int64
	TotalDuplicatesSkipped int64
	TotalOutOfOrderSkipped int64
}

func NewProcessor() *Processor {
	return &Processor{
		devicePages: make(map[string]*VehicleState),
	}
}

func (p *Processor) Handle(payload []byte) {
	var msg TelemetryMessage
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		return
	}
	if msg.DeviceID == "" {
		return
	}
	p.TotalMessagesReceived++ // how much msg did we got
	p.mu.Lock()
	vehicle := p.devicePages[msg.DeviceID]
	if vehicle == nil {
		vehicle = &VehicleState{DeviceID: msg.DeviceID}
		p.devicePages[msg.DeviceID] = vehicle
	}
	p.mu.Unlock()
	if vehicle.LastTimestamp == msg.Timestamp && vehicle.CurrentOdometerMeters == msg.OdometerMeters {
		p.TotalDuplicatesSkipped++
		return
	}
	if msg.Timestamp < vehicle.LastTimestamp {
		p.TotalOutOfOrderSkipped++
		return
	}
	if vehicle.CurrentOdometerMeters > 0 && msg.OdometerMeters < vehicle.CurrentOdometerMeters {
		vehicle.CurrentOdometerMeters = msg.OdometerMeters // naya odometer note karo
		vehicle.LastTimestamp = msg.Timestamp
		vehicle.LastMessageTime = time.Now()
		vehicle.IsOnline = true
		return // distance add mat karo — reboot hai
	}
	// distance = naya odometer - purana odometer
	// sirf tab count karo jab pehle se reading ho (pehla message baseline hai)
	if vehicle.CurrentOdometerMeters > 0 {
		delta := msg.OdometerMeters - vehicle.CurrentOdometerMeters
		if delta >= 0 {
			vehicle.TotalDistanceMeters += delta
		}
	}

	vehicle.CurrentOdometerMeters = msg.OdometerMeters
	vehicle.CurrentSpeedKph = msg.SpeedKph
	vehicle.CurrentBatteryPct = msg.BatteryPct
	vehicle.CurrentState = msg.State
	vehicle.LastTimestamp = msg.Timestamp
	vehicle.LastMessageTime = time.Now()
	vehicle.IsOnline = true
}
