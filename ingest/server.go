package main

import (
	"encoding/json"
	"net/http"
)

// handleHTTP
func handleHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/devices":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(proc.SnapshotAll())
	case "/api/stream":
		http.Error(w, "coming soon", http.StatusNotImplemented)
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
