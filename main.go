package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	counter int
	mu      sync.Mutex
	peers   = make(map[string]bool)
)

func main() {
	// Get port and peers from environment variables / command-line arguments
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("❌ PORT not set, using default 8088")
		port = "8088"
	}

	peerList := os.Getenv("PEERS")
	if peerList != "" {
		for _, peer := range strings.Split(peerList, ",") {
			peers[peer] = true
		}
	}

	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "--peers=") {
		peerList := strings.TrimPrefix(os.Args[1], "--peers=")
		for _, peer := range strings.Split(peerList, ",") {
			peers[peer] = true
		}
	}

	// Start background health check for peers
	go healthCheckPeers()
	fmt.Printf("📡 Node started on port %s, Peers: %v\n", port, getPeerList())

	http.HandleFunc("/register", registerPeer)
	http.HandleFunc("/peers", getPeers)
	http.HandleFunc("/remove-peer", removePeer)
	http.HandleFunc("/increment", incrementHandler)
	http.HandleFunc("/count", countHandler)
	http.HandleFunc("/sync", syncHandler)
	http.HandleFunc("/health", healthCheck)
	serverAddr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}

// getPeerList returns the list of peer IDs as a slice.
func getPeerList() []string {
	peerList := make([]string, 0, len(peers))
	for p := range peers {
		peerList = append(peerList, p)
	}
	return peerList
}

func getPeers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(getPeerList())
}

// healthCheckPeers periodically checks if peers are alive.
func healthCheckPeers() {
	for {
		time.Sleep(5 * time.Second)
		for peer := range peers {
			url := fmt.Sprintf("http://%s/health", peer)
			_, err := http.Get(url)
			if err != nil {
				delete(peers, peer)
				fmt.Printf("Removed dead peer: %s\n", peer)
			}
		}
	}
}

// healthCheck is a simple endpoint to check node health.
func healthCheck(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// removePeer removes a peer based on the provided ID.
func removePeer(w http.ResponseWriter, r *http.Request) {
	var peer struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	delete(peers, peer.ID)
	fmt.Printf("Removed peer: %s\n", peer.ID)
	w.WriteHeader(http.StatusOK)
}

// registerPeer adds a new peer.
func registerPeer(w http.ResponseWriter, r *http.Request) {
	var peer struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	peers[peer.ID] = true
	fmt.Printf("Registered peer: %s\n", peer.ID)
	w.WriteHeader(http.StatusOK)
}

// countHandler returns the current counter value.
func countHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	response := map[string]int{"count": counter}
	json.NewEncoder(w).Encode(response)
}

// incrementHandler increments the counter locally and propagates the new value.
func incrementHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	counter++ // local increment
	newValue := counter
	mu.Unlock()

	fmt.Printf("✅ Counter incremented: %d\n", newValue)

	// Propagate the new counter value to all peers concurrently
	var wg sync.WaitGroup
	for peer := range peers {
		wg.Add(1)
		go propagateCounter(peer, newValue, &wg)
	}
	wg.Wait()
	w.WriteHeader(http.StatusOK)
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shouldPropagate := false

	mu.Lock()
	if payload.Count > counter {
		counter = payload.Count
		shouldPropagate = true
	}
	syncedValue := counter
	mu.Unlock()

	fmt.Printf("🔄 Counter synced from peer, new value: %d\n", syncedValue)

	// Propagate only if there was an update
	if shouldPropagate {
		var wg sync.WaitGroup
		for peer := range peers {
			wg.Add(1)
			go propagateCounter(peer, syncedValue, &wg)
		}
		wg.Wait()
	}

	w.WriteHeader(http.StatusOK)
}

// propagateCounter sends the new counter value to a peer.
func propagateCounter(peer string, value int, wg *sync.WaitGroup) {
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
		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Printf("✅ Counter synced to %s\n", peer)
			return
		}
		fmt.Printf("⚠️ Failed to propagate to %s: %v\n", peer, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	fmt.Printf("❌ Final failure: Could not propagate counter to %s\n", peer)
}
