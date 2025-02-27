package handlers

import (
	"assignmet/advance/internal/counter"
	"assignmet/advance/internal/peers"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Handler processes HTTP requests
type Handler struct {
	nodeID         string
	counterService *counter.Counter
	peersManager   *peers.PeersManager
}

// New creates a new Handler instance
func New(nodeID string, counterService *counter.Counter, peersManager *peers.PeersManager) *Handler {
	return &Handler{
		nodeID:         nodeID,
		counterService: counterService,
		peersManager:   peersManager,
	}
}

// RegisterPeer adds a new peer
func (h *Handler) RegisterPeer(w http.ResponseWriter, r *http.Request) {
	var peer struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Don't add ourselves as a peer
	if peer.ID == h.nodeID {
		w.WriteHeader(http.StatusOK)
		return
	}

	if h.peersManager.AddPeer(peer.ID) {
		fmt.Printf("✅ Registered new peer: %s\n", peer.ID)
	}

	w.WriteHeader(http.StatusOK)
}

// GetPeers returns the list of known peers
func (h *Handler) GetPeers(w http.ResponseWriter, r *http.Request) {
	peerList := h.peersManager.GetPeerList()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peerList)
}

// RemovePeer removes a peer based on the provided ID
func (h *Handler) RemovePeer(w http.ResponseWriter, r *http.Request) {
	var peer struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	h.peersManager.RemovePeer(peer.ID)
	fmt.Printf("🔌 Removed peer: %s\n", peer.ID)
	w.WriteHeader(http.StatusOK)
}

// HealthCheck is a simple endpoint to check node health
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"node_id": h.nodeID,
	})
}

// CountHandler returns the current counter value
func (h *Handler) CountHandler(w http.ResponseWriter, r *http.Request) {
	localCount := h.counterService.GetValue()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   localCount,
		"node_id": h.nodeID,
	})
}

// IncrementHandler increments the counter locally and propagates the new value
func (h *Handler) IncrementHandler(w http.ResponseWriter, r *http.Request) {
	newValue := h.counterService.Increment()
	fmt.Printf("✅ Counter incremented: %d\n", newValue)

	// Propagate the new counter value to all peers concurrently
	var wg sync.WaitGroup
	peerList := h.peersManager.GetPeerList()

	for _, peer := range peerList {
		wg.Add(1)
		go h.propagateCounter(peer, newValue, &wg)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   newValue,
		"node_id": h.nodeID,
	})
}

// SyncHandler updates the local counter if the received value is greater
func (h *Handler) SyncHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shouldPropagate := h.counterService.Sync(payload.Count)
	syncedValue := h.counterService.GetValue()

	fmt.Printf("🔄 Counter synced, new value: %d\n", syncedValue)

	// Propagate only if there was an update
	if shouldPropagate {
		var wg sync.WaitGroup
		peerList := h.peersManager.GetPeerList()

		for _, peer := range peerList {
			wg.Add(1)
			go h.propagateCounter(peer, syncedValue, &wg)
		}
		wg.Wait()
	}

	w.WriteHeader(http.StatusOK)
}

// DiscoveryHandler handles HTTP requests for peer discovery
func (h *Handler) DiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		peerList := h.peersManager.GetPeerList()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_id": h.nodeID,
			"peers":   peerList,
		})
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// propagateCounter sends the new counter value to a peer
func (h *Handler) propagateCounter(peer string, value int, wg *sync.WaitGroup) {
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
	h.peersManager.RemovePeer(peer)
}
