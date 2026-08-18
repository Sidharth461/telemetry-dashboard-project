# Live Vehicle Tracking System

## 3 Components

### 1. Simulator
- Job: sends fake vehicle data to the MQTT broker

### 2. Ingest Service
- Receives data, processes it, and exposes an API

### 3. Web UI
- Shows live vehicle status

Also: there is an EMQX (Bmax) broker that routes the MQTT messages.

**All components should run with a single `docker-compose` command.**

---

## 1. Simulator

1. It's a program that plays the role of 20–50 vehicles.
2. Each vehicle has its own state machine that follows: **moving, idle, charging, off**.
3. Sends a message every 5–10 seconds — speed, battery, odometer — all values should be realistic.

### Fault Injection (Important)
- Duplicate message (1%)
- Out-of-order message (2%)
- Odometer glitch/skip (<1%)
- Some devices will go silent

---

## 2. Ingest Service

This is the MQTT subscriber and the broadcaster of device snapshots.

1. **Odometer monotonic check** — if the odometer jumps/drops to 0, treat it as the vehicle having rebooted; never show negative distance.
2. **out of orders / timestamp pruning** — old/duplicate-timestamp messages should not be re-processed.
3. **Duplicate message should not be counted again.**
4. **Silence detection** — if no message arrives within 15 seconds, mark the device offline.
5. **Backpressure** — a slow subscriber should not block ingestion.

### In-memory storage / cache
- Rolling 60-second window for averages
- Last transition log, live subscribe
- Throttled update/broadcast for devices

---

## 3. Web UI (React)
- Rigid watchdog
- Vehicle card: id, state, speed, battery, distance
- "Time since last update"

---

## Decision
**SSE confirmed — not WebSocket, we'll use Server-Sent Events (current code already has SSE).**