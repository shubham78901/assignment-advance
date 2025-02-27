package handlers

import (
	"assignmet/advance/config"
	"assignmet/advance/discovery"
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"
	"sync"
	"time"
)

var (
	mu sync.Mutex
)

func IncrementHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cfg.Counter++ // local increment
		newValue := cfg.Counter
		mu.Unlock()

		fmt.Printf("✅ Counter incremented: %d\n", newValue)

		// Propagate the new counter value to all peers concurrently
		var wg sync.WaitGroup

		discovery.PeersMu.RLock()
		currentPeers := make([]string, 0, len(cfg.Peers))
		for peer := range cfg.Peers {
			currentPeers = append(currentPeers, peer)
		}
		discovery.PeersMu.RUnlock()

		for _, peer := range currentPeers {
			wg.Add(1)
			go propagateCounter(cfg, peer, newValue, &wg)
		}
		wg.Wait()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   newValue,
			"node_id": cfg.NodeID,
		})
	}
}

func CountHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		localCount := cfg.Counter
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   localCount,
			"node_id": cfg.NodeID,
		})
	}
}

func SyncHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		shouldPropagate := false

		mu.Lock()
		if payload.Count > cfg.Counter {
			cfg.Counter = payload.Count
			shouldPropagate = true
		}
		syncedValue := cfg.Counter
		mu.Unlock()

		fmt.Printf("🔄 Counter synced, new value: %d\n", syncedValue)

		// Propagate only if there was an update
		if shouldPropagate {
			var wg sync.WaitGroup

			discovery.PeersMu.RLock()
			currentPeers := make([]string, 0, len(cfg.Peers))
			for peer := range cfg.Peers {
				currentPeers = append(currentPeers, peer)
			}
			discovery.PeersMu.RUnlock()

			for _, peer := range currentPeers {
				wg.Add(1)
				go propagateCounter(cfg, peer, syncedValue, &wg)
			}
			wg.Wait()
		}

		w.WriteHeader(http.StatusOK)
	}
}

func propagateCounter(cfg *config.Config, peer string, value int, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://%s/sync", peer)
	client := http.Client{Timeout: 2 * time.Second}
	payload := map[string]int{"count": value}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal payload for %s: %v\n", peer, err)
		return
	}

	for i := 0; i < 3; i++ {
		fmt.Printf("🔄 Propagating counter %d to %s (Attempt %d)\n", value, peer, i+1)
		resp, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Printf("✅ Counter synced to %s\n", peer)
				return
			}
		}
		fmt.Printf("⚠️ Failed to propagate to %s: %v\n", peer, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	fmt.Printf("❌ Final failure: Could not propagate counter to %s\n", peer)

	// Remove failed peer after multiple retries
	discovery.PeersMu.Lock()
	delete(cfg.Peers, peer)
	discovery.PeersMu.Unlock()
}
