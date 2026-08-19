package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type Hub struct {
	proc      *Processor
	broadcast chan []byte
}

func NewHub(proc *Processor) *Hub {
	return &Hub{
		proc:      proc,
		broadcast: make(chan []byte, 16),
	}
}

func (h *Hub) Run() {
	for {
		time.Sleep(1 * time.Second)

		snapshots := h.proc.SnapshotAll()

		data, err := json.Marshal(snapshots)
		if err != nil {
			continue
		}

		select {
		case h.broadcast <- data:
		default:
		}
	}
}

func (h *Hub) ServeStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	client := make(chan []byte, 16)

	go func() {
		for {
			data := <-h.broadcast
			select {
			case client <- data:
			default:
			}
		}
	}()

	for {
		data := <-client
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()
	}
}