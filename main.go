package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func intakeHandler(client *CaptureClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		var event MatterEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil || event.MatterID == "" || event.Action == "" {
			http.Error(w, "matter_id and action are required", http.StatusBadRequest)
			return
		}
		if err := client.Capture(event, fmt.Errorf("follow-up record queued for %s", groupKey(event))); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "captured", "group": groupKey(event)})
	}
}

func main() {
	client, err := NewCaptureClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/matter-events", intakeHandler(client))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
