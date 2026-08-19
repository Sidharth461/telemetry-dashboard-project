package main

import (
	"encoding/json"
	"sort"
	"sync"
	"time"
)

type TelemetryMessage struct {
	DeviceID       string   `json:"deviceId"`
	Timestamp      string   `json:"ts"`
	State          string   `json:"state"`
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
	TotalDistanceMeters   float64   // service start to total distance
	LastTimestamp         string    // out-of-order check
	LastMessageTime       time.Time // silence/offline check for that purpose
	IsOnline              bool
	Window                []Sample
	WindowDistance        float64 // last 60s  distance
	AvgSpeedKph           float64 // last 60s  avg speed
	AvgBatteryPct         float64 // last 60s  avg battery
	MovingPercent         float64 // last 60s may how much % MOVING
}
type Sample struct {
	At      time.Time
	Odo     float64
	Speed   float64
	Battery float64
	State   string
}
type Processor struct {
	mu                     sync.Mutex               // race prevention
	devicePages            map[string]*VehicleState // key = deviceId
	window                 time.Duration            // rolling window size (60s)
	offlineAfter           time.Duration            // how may much time it has been silence = OFFLINE
	TotalMessagesReceived  int64
	TotalDuplicatesSkipped int64
	TotalOutOfOrderSkipped int64
}

func NewProcessor(window, offlineAfter time.Duration) *Processor {
	return &Processor{
		devicePages:  make(map[string]*VehicleState),
		window:       window,
		offlineAfter: offlineAfter,
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

	now := time.Now()

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
		vehicle.CurrentOdometerMeters = msg.OdometerMeters //note the new odometer
		vehicle.LastTimestamp = msg.Timestamp
		vehicle.LastMessageTime = time.Now()
		vehicle.IsOnline = true
		return //  reboot
	}

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
	vehicle.LastMessageTime = now
	vehicle.IsOnline = true

	vehicle.Window = append(vehicle.Window, Sample{
		At: now, Odo: msg.OdometerMeters, Speed: msg.SpeedKph, Battery: msg.BatteryPct, State: msg.State,
	})

	cutoff := now.Add(-p.window)
	newWindow := vehicle.Window[:0]
	for _, s := range vehicle.Window {
		if !s.At.Before(cutoff) {
			newWindow = append(newWindow, s)
		}
	}
	vehicle.Window = newWindow

	if len(vehicle.Window) >= 2 {
		windowDistance := vehicle.Window[len(vehicle.Window)-1].Odo - vehicle.Window[0].Odo
		if windowDistance < 0 {
			windowDistance = 0
		}
		vehicle.WindowDistance = windowDistance
	}

	var speedSum, batterySum float64
	for _, s := range vehicle.Window {
		speedSum += s.Speed
		batterySum += s.Battery
	}
	if n := len(vehicle.Window); n > 0 {
		vehicle.AvgSpeedKph = speedSum / float64(n)
		vehicle.AvgBatteryPct = batterySum / float64(n)
	}

	movingCount := 0
	for _, s := range vehicle.Window {
		if s.State == "MOVING" {
			movingCount++
		}
	}
	if len(vehicle.Window) > 0 {
		vehicle.MovingPercent = float64(movingCount) / float64(len(vehicle.Window)) * 100
	}
}

// Snapshot = Data from one device that can be sent to the browser (clean copy)
type DeviceSnapshot struct {
	DeviceID            string  `json:"deviceId"`
	State               string  `json:"state"`
	SpeedKph            float64 `json:"speedKph"`
	BatteryPct          float64 `json:"batteryPct"`
	OdometerMeters      float64 `json:"odometerMeters"`
	DistanceMeters      float64 `json:"distanceMeters"`
	TotalDistanceMeters float64 `json:"totalDistanceMeters"`
	AvgSpeedKph         float64 `json:"avgSpeedKph"`
	AvgBatteryPct       float64 `json:"avgBatteryPct"`
	MovingPercent       float64 `json:"movingPercent"`
	LastMessageAt       string  `json:"lastMessageAt"`
	IsOnline            bool    `json:"isOnline"`
}

// SnapshotAll = create all clean copies of all vehicle (for the list)
func (p *Processor) SnapshotAll() []DeviceSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]DeviceSnapshot, 0, len(p.devicePages))
	for _, v := range p.devicePages {
		out = append(out, snapshotOf(v))
	}

	//sort it according to deviceId  (map gives random order )
	sort.Slice(out, func(i, j int) bool {
		return out[i].DeviceID < out[j].DeviceID
	})
	return out
}

// snapshotOf = one device the internal state → clean copy
func snapshotOf(v *VehicleState) DeviceSnapshot {
	return DeviceSnapshot{
		DeviceID:            v.DeviceID,
		State:               v.CurrentState,
		SpeedKph:            v.CurrentSpeedKph,
		BatteryPct:          v.CurrentBatteryPct,
		OdometerMeters:      v.CurrentOdometerMeters,
		DistanceMeters:      v.WindowDistance,
		TotalDistanceMeters: v.TotalDistanceMeters,
		AvgSpeedKph:         v.AvgSpeedKph,
		AvgBatteryPct:       v.AvgBatteryPct,
		MovingPercent:       v.MovingPercent,
		LastMessageAt:       v.LastMessageTime.Format("2006-01-02T15:04:05.000Z"),
		IsOnline:            v.IsOnline,
	}
}

// Sweep = call every second — 15s still silent device mark it as OFFLINE
func (p *Processor) Sweep() {
	cutoff := time.Now().Add(-p.offlineAfter)

	p.mu.Lock()
	for _, v := range p.devicePages {
		if v.LastMessageTime.Before(cutoff) {
			v.IsOnline = false
		}
	}
	p.mu.Unlock()
}
