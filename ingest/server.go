package main

import (
	"encoding/json"
	"net/http"
)

var hub *Hub

// handleHTTP
func handleHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/devices":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(proc.SnapshotAll())
	case "/api/stream":
		hub.ServeStream(w, r)
	case "/healthz":
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	case "/stats":
		json.NewEncoder(w).Encode(map[string]int64{
			"messagesReceived":  proc.TotalMessagesReceived,
			"duplicatesSkipped": proc.TotalDuplicatesSkipped,
			"outOfOrderSkipped": proc.TotalOutOfOrderSkipped,
		})
	default:
		http.NotFound(w, r)
	}
}
