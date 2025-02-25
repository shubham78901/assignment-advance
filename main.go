package main

import (
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
	peers   []string
)

func main() {
	// Get port and peers from environment variables (or command-line arguments)
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("❌ PORT not set, using default 8088")
		port = "8088"
	}

	peerList := os.Getenv("PEERS")
	if peerList != "" {
		peers = strings.Split(peerList, ",")
	}

	fmt.Printf("📡 Node started on port %s, Peers: %v\n", port, peers)

	// Primary endpoints
	http.HandleFunc("/count", countHandler)
	http.HandleFunc("/increment", incrementHandler)
	http.HandleFunc("/sync", syncHandler)
	http.HandleFunc("/register", registerPeer)
	http.HandleFunc("/peers", getPeers)
	http.HandleFunc("/remove-peer", removePeer)

	// Duplicate endpoint registrations using wrapper functions.
	// (Because of duplicate registrations, these will override the above ones.)
	http.HandleFunc("/increment", incrementCounter)
	http.HandleFunc("/count", getCounter)
	http.HandleFunc("/sync", syncCounter)
	http.HandleFunc("/health", healthCheck)

	serverAddr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}

// countHandler returns the current counter.
func countHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	response := map[string]int{"count": counter}
	json.NewEncoder(w).Encode(response)
}

// incrementHandler increments the counter locally and propagates the update.
func incrementHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	counter++
	currentCount := counter
	mu.Unlock()

	fmt.Printf("✅ Counter incremented: %d\n", currentCount)

	// Propagate increment to all peers concurrently.
	var wg sync.WaitGroup
	for _, peer := range peers {
		wg.Add(1)
		go propagateIncrement(peer, &wg)
	}
	wg.Wait() // Wait for all propagations to complete

	w.WriteHeader(http.StatusOK)
}

// syncHandler updates the counter when receiving a sync request from a peer.
func syncHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	counter++
	currentCount := counter
	mu.Unlock()

	fmt.Printf("🔄 Counter synced from peer, new value: %d\n", currentCount)
	w.WriteHeader(http.StatusOK)
}

// propagateIncrement sends a sync request to a peer to update its counter.
func propagateIncrement(peer string, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://%s/sync", peer)
	client := http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 3; i++ {
		fmt.Printf("🔄 Propagating increment to %s (Attempt %d)\n", peer, i+1)
		resp, err := client.Post(url, "application/json", nil)

		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Printf("✅ Increment propagated to %s\n", peer)
			return
		}

		fmt.Printf("⚠️ Failed to propagate to %s: %v\n", peer, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	fmt.Printf("❌ Final failure: Could not propagate increment to %s\n", peer)
}

// registerPeer registers a new peer.
// It expects a POST request with JSON body: {"peer": "address:port"}.
func registerPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var data struct {
		Peer string `json:"peer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if peer already exists.
	for _, p := range peers {
		if p == data.Peer {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Peer already registered")
			return
		}
	}
	peers = append(peers, data.Peer)
	fmt.Printf("🔗 Peer registered: %s\n", data.Peer)
	w.WriteHeader(http.StatusCreated)
}

// getPeers returns the list of registered peers.
func getPeers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(peers)
}

// removePeer removes a peer from the registered list.
// It expects a POST request with JSON body: {"peer": "address:port"}.
func removePeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var data struct {
		Peer string `json:"peer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()

	for i, p := range peers {
		if p == data.Peer {
			peers = append(peers[:i], peers[i+1:]...)
			fmt.Printf("🗑️ Peer removed: %s\n", data.Peer)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Peer not found", http.StatusNotFound)
}

// healthCheck provides a simple health check endpoint.
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// The following wrapper functions call the original handlers.
// They are registered last and override the earlier registrations.

// incrementCounter wraps the incrementHandler.
func incrementCounter(w http.ResponseWriter, r *http.Request) {
	incrementHandler(w, r)
}

// getCounter wraps the countHandler.
func getCounter(w http.ResponseWriter, r *http.Request) {
	countHandler(w, r)
}

// syncCounter wraps the syncHandler.
func syncCounter(w http.ResponseWriter, r *http.Request) {
	syncHandler(w, r)
}
