package main

import (
	"testing"
	"time"
)

// Test 1: pehli baar message aaya → notebook mein page bana
func TestFirstMessageCreatesDevicePage(t *testing.T) {
	p := NewProcessor(60*time.Second, 15*time.Second)

	payload := `{"deviceId":"VEH-0001","ts":"2026-08-19T10:00:00.000Z","state":"MOVING","speedKph":45,"batteryPct":80,"odometerMeters":1000}`
	p.Handle([]byte(payload))

	p.mu.Lock()
	v := p.devicePages["VEH-0001"]
	p.mu.Unlock()

	if v == nil {
		t.Fatal("VEH-0001 ka page nahi bana") // page hona chahiye
	}
	if v.CurrentSpeedKph != 45 {
		t.Fatalf("speed = %v, expected 45", v.CurrentSpeedKph) // speed 45 honi chahiye
	}
}

// Test 2: counter reset — odometer 1000 se 0 → negative distance nahi
func TestCounterResetNoNegativeDistance(t *testing.T) {
	p := NewProcessor(60*time.Second, 15*time.Second)

	// msg 1: odometer 1000
	p.Handle([]byte(`{"deviceId":"VEH-0001","ts":"2026-08-19T10:00:00.000Z","odometerMeters":1000}`))
	// msg 2: odometer 0 (tracker reboot)
	p.Handle([]byte(`{"deviceId":"VEH-0001","ts":"2026-08-19T10:00:01.000Z","odometerMeters":0}`))

	p.mu.Lock()
	v := p.devicePages["VEH-0001"]
	p.mu.Unlock()

	if v.TotalDistanceMeters != 0 {
		t.Fatalf("distance = %v, expected 0 (no negative)", v.TotalDistanceMeters)
	}
	if v.CurrentOdometerMeters != 0 {
		t.Fatalf("odometer = %v, expected 0", v.CurrentOdometerMeters)
	}
}

// Test 3: out-of-order — purana timestamp wala message latest ko overwrite nahi karega
func TestOutOfOrderNotOverwritten(t *testing.T) {
	p := NewProcessor(60*time.Second, 15*time.Second)

	// msg 1: speed 50 at 10:00:02 (latest)
	p.Handle([]byte(`{"deviceId":"VEH-0001","ts":"2026-08-19T10:00:02.000Z","speedKph":50,"odometerMeters":1000}`))
	// msg 2: speed 45 at 10:00:01 (PURANA) → skip
	p.Handle([]byte(`{"deviceId":"VEH-0001","ts":"2026-08-19T10:00:01.000Z","speedKph":45,"odometerMeters":999}`))

	p.mu.Lock()
	v := p.devicePages["VEH-0001"]
	p.mu.Unlock()

	if v.CurrentSpeedKph != 50 {
		t.Fatalf("speed = %v, expected 50 (old msg overwrote!)", v.CurrentSpeedKph)
	}
}

// Test 4: duplicate — same ts + same odometer → same message 2 → skip
func TestDuplicateSkipped(t *testing.T) {
	p := NewProcessor(60*time.Second, 15*time.Second)

	// same message 2 baar bheja (simulator ka 1% fault)
	payload := `{"deviceId":"VEH-0001","ts":"2026-08-19T10:00:00.000Z","speedKph":45,"odometerMeters":1000}`
	p.Handle([]byte(payload))
	p.Handle([]byte(payload))

	if p.TotalDuplicatesSkipped != 1 {
		t.Fatalf("duplicates skipped = %v, expected 1", p.TotalDuplicatesSkipped)
	}
}

