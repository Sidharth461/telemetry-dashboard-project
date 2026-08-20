# Fleet Telemetry Dashboard

A live telemetry pipeline for a fleet of vehicles: a Go simulator publishes MQTT messages, the ingest service processes them (with fault handling), and a React dashboard shows everything live via SSE.

## Architecture

```
+-------------+   MQTT (1883)   +-----------+   MQTT   +-----------+   HTTP/SSE   +-------------+
|  Simulator  | ---------------> |   EMQX    | ------> |  Ingest   | ----------> |  React UI   |
|    (Go)     |  fleet/+/+/telemetry  | (broker)  |          |   (Go)    |             |   (Vite)    |
+-------------+                  +-----------+          +-----------+             +-------------+
```

## Services

| Service  | Tech              | What it does                                                                                             |
|----------|-------------------|---------------------------------------------------------------------------------------------------------|
| simulator| Go                | Publishes vehicle telemetry to MQTT - 20 vehicles, 5 msgs/sec each. Injects realistic faults: duplicate messages, out-of-order messages, odometer counter resets, and silent devices (30-60s no messages). |
| ingest   | Go                | Subscribes to `fleet/+/+/telemetry`, stores per-vehicle state, detects faults (duplicates, out-of-order, counter reset, offline after 15s), computes 60s rolling averages, and exposes a REST + SSE API. |
| ui       | React + Vite + Tailwind | Live dashboard - device cards with speed, battery, moving %, 60s distance, last message time, and a messages/sec counter. Updates via SSE, no polling. |

## API

| Endpoint        | Description                                                        |
|-----------------|--------------------------------------------------------------------|
| `GET /api/devices` | JSON snapshot of all devices (sorted)                            |
| `GET /api/stream` | SSE stream - full snapshot + stats every second                  |
| `GET /healthz`  | Health check                                                       |
| `GET /stats`    | Ingestion counters (received, duplicates skipped, out-of-order skipped) |

## Configuration

All config comes from environment variables (`.env` for local dev, `-e` flags for Docker). See `.env.example`.

## Run with Docker

EMQX (the MQTT broker) is started with docker-compose:

```bash
docker compose up -d
```

The three services each have a Dockerfile:

```bash
# build images
docker build -t fleet-simulator ./simulator
docker build -t fleet-ingest ./ingest
docker build -t fleet-ui ./Telementry-Ui

# run services
docker run -d --name sim --network host -e BROKER_HOST=localhost fleet-simulator
docker run -d --name ingest -p 8080:8080 -e BROKER_HOST=host.docker.internal fleet-ingest
docker run -d --name ui -p 8081:80 fleet-ui
```

Open `http://localhost:8081` for the dashboard.

> **Note:** `host.docker.internal` lets the ingest container reach EMQX published on the host. On Linux you can use `localhost` with `--network host` instead.

## Run locally (development)

```bash
# terminal 1 - broker
docker compose up -d

# terminal 2 - simulator
cd simulator && go run .

# terminal 3 - ingest
cd ingest && go run .

# terminal 4 - UI
cd Telementry-Ui && npm install && npm run dev
```

Open `http://localhost:5173`.

## Tests

```bash
cd ingest
go test ./...          # unit tests (fault handling)
go test -race ./...    # race detector
```

## Project structure

```
+-- simulator/          # Go MQTT publisher with fault injection
+-- ingest/             # Go MQTT consumer + REST/SSE API
+-- Telementry-Ui/      # React dashboard (Vite + Tailwind)
+-- docker-compose.yml  # EMQX broker
+-- .env.example        # config template
+-- postman/            # Postman collection for the API
```