package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	counter          int
	mu               sync.Mutex
	peers            = make(map[string]bool)
	peersMu          sync.RWMutex
	discoveryPort    = 8089
	servicePort      string
	nodeID           string
	discoveryEnabled bool
)

func main() {
	// Get port from environment variables / command-line arguments
	servicePort = os.Getenv("PORT")
	if servicePort == "" {
		fmt.Println("❌ PORT not set, using default 8088")
		servicePort = "8088"
	}

	// Generate a unique node ID based on IP and port
	nodeID = getNodeID(servicePort)
	fmt.Printf("🆔 Node ID: %s\n", nodeID)

	// Setup initial peers from environment or command line
	setupInitialPeers()

	// Initialize service discovery
	setupServiceDiscovery()

	// Start regular peer health checks
	go healthCheckPeers()

	// Start broadcasting presence regularly
	if discoveryEnabled {
		go broadcastPresence()
	}

	fmt.Printf("📡 Node started on port %s, Peers: %v\n", servicePort, getPeerList())

	// Setup HTTP handlers
	http.HandleFunc("/register", registerPeer)
	http.HandleFunc("/peers", getPeers)
	http.HandleFunc("/remove-peer", removePeer)
	http.HandleFunc("/increment", incrementHandler)
	http.HandleFunc("/count", countHandler)
	http.HandleFunc("/sync", syncHandler)
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/discovery", discoveryHandler)

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%s", servicePort)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}

// getNodeID generates a unique ID for this node
func getNodeID(port string) string {
	// Get the host's preferred outbound IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback to local IP if cannot determine outbound IP
		return fmt.Sprintf("localhost:%s", port)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%s:%s", localAddr.IP.String(), port)
}

// setupInitialPeers configures initial peers from environment or command line
func setupInitialPeers() {
	peerList := os.Getenv("PEERS")
	if peerList != "" {
		for _, peer := range strings.Split(peerList, ",") {
			if peer != nodeID {
				peersMu.Lock()
				peers[peer] = true
				peersMu.Unlock()
			}
		}
	}

	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "--peers=") {
				peerList := strings.TrimPrefix(arg, "--peers=")
				for _, peer := range strings.Split(peerList, ",") {
					if peer != nodeID {
						peersMu.Lock()
						peers[peer] = true
						peersMu.Unlock()
					}
				}
			} else if strings.HasPrefix(arg, "--discovery-port=") {
				portStr := strings.TrimPrefix(arg, "--discovery-port=")
				fmt.Sscanf(portStr, "%d", &discoveryPort)
			} else if arg == "--disable-discovery" {
				discoveryEnabled = false
				return
			}
		}
	}

	// Enable discovery by default
	discoveryEnabled = true
}

// setupServiceDiscovery initializes the service discovery mechanism
func setupServiceDiscovery() {
	if !discoveryEnabled {
		fmt.Println("ℹ️ Service discovery is disabled")
		return
	}

	// Start listening for multicast discovery on a separate goroutine
	go listenForDiscovery()

	// Register with any peers provided in the initial configuration
	// They will help spread our presence to the rest of the network
	for peer := range peers {
		go registerWithPeer(peer)
	}

	fmt.Printf("🔍 Service discovery enabled on port %d\n", discoveryPort)
}

// listenForDiscovery listens for UDP broadcast/multicast discovery messages
func listenForDiscovery() {
	addr := net.UDPAddr{
		Port: discoveryPort,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("❌ Failed to start discovery listener: %v\n", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("⚠️ Error receiving discovery packet: %v\n", err)
			continue
		}

		// Parse the received peer information
		peerID := string(buffer[:n])
		if peerID == nodeID {
			// Ignore our own broadcasts
			continue
		}

		// Add the new peer
		peersMu.Lock()
		if _, exists := peers[peerID]; !exists {
			peers[peerID] = true
			fmt.Printf("🔍 Discovered new peer via multicast: %s (from %s)\n", peerID, remoteAddr)

			// Register with the new peer to establish two-way connection
			go registerWithPeer(peerID)
		}
		peersMu.Unlock()
	}
}

// broadcastPresence periodically announces this node's presence via UDP
// broadcastPresence periodically announces this node's presence via UDP
func broadcastPresence() {
	// Instead of using 255.255.255.255, try the Docker bridge network broadcast address
	addr := net.UDPAddr{
		Port: discoveryPort,
		IP:   net.ParseIP("172.18.255.255"), // Docker bridge network broadcast
	}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Printf("❌ Failed to create discovery broadcaster: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		// Broadcast our node ID
		_, err := conn.Write([]byte(nodeID))
		if err != nil {
			fmt.Printf("⚠️ Failed to broadcast presence: %v\n", err)
		} else {
			fmt.Println("📢 Broadcasted presence to network")
		}

		// Wait before next broadcast
		time.Sleep(30 * time.Second)
	}
}

// discoveryHandler handles HTTP requests for peer discovery
func discoveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return the list of known peers
		peersMu.RLock()
		peerList := getPeerList()
		peersMu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_id": nodeID,
			"peers":   peerList,
		})
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// registerWithPeer attempts to register with a known peer
func registerWithPeer(peerID string) {
	url := fmt.Sprintf("http://%s/register", peerID)
	payload := map[string]string{"id": nodeID}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal registration payload for %s: %v\n", peerID, err)
		return
	}

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("⚠️ Failed to register with peer %s: %v\n", peerID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Successfully registered with peer: %s\n", peerID)

		// Also fetch peers from this peer to expand our network
		go fetchPeersFromPeer(peerID)
	} else {
		fmt.Printf("⚠️ Peer %s returned status %d during registration\n", peerID, resp.StatusCode)
	}
}

// fetchPeersFromPeer gets the peer list from a known peer
func fetchPeersFromPeer(peerID string) {
	url := fmt.Sprintf("http://%s/peers", peerID)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch peers from %s: %v\n", peerID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("⚠️ Peer %s returned status %d when fetching peers\n", peerID, resp.StatusCode)
		return
	}

	var peerList []string
	if err := json.NewDecoder(resp.Body).Decode(&peerList); err != nil {
		fmt.Printf("❌ Failed to decode peer list from %s: %v\n", peerID, err)
		return
	}

	// Add new peers to our list
	peersMu.Lock()
	for _, newPeer := range peerList {
		if newPeer != nodeID && !peers[newPeer] {
			peers[newPeer] = true
			fmt.Printf("🔍 Discovered new peer via %s: %s\n", peerID, newPeer)

			// Register with the new peer
			go registerWithPeer(newPeer)
		}
	}
	peersMu.Unlock()
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
	peersMu.RLock()
	peerList := getPeerList()
	peersMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peerList)
}

// healthCheckPeers periodically checks if peers are alive.
func healthCheckPeers() {
	for {
		time.Sleep(5 * time.Second)

		peersMu.RLock()
		currentPeers := make([]string, 0, len(peers))
		for peer := range peers {
			currentPeers = append(currentPeers, peer)
		}
		peersMu.RUnlock()

		for _, peer := range currentPeers {
			url := fmt.Sprintf("http://%s/health", peer)
			_, err := http.Get(url)
			if err != nil {
				peersMu.Lock()
				delete(peers, peer)
				peersMu.Unlock()
				fmt.Printf("❌ Removed dead peer: %s\n", peer)
			}
		}
	}
}

// healthCheck is a simple endpoint to check node health.
func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"node_id": nodeID,
	})
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

	peersMu.Lock()
	delete(peers, peer.ID)
	peersMu.Unlock()

	fmt.Printf("🔌 Removed peer: %s\n", peer.ID)
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

	// Don't add ourselves as a peer
	if peer.ID == nodeID {
		w.WriteHeader(http.StatusOK)
		return
	}

	peersMu.Lock()
	wasNew := !peers[peer.ID]
	peers[peer.ID] = true
	peersMu.Unlock()

	if wasNew {
		fmt.Printf("✅ Registered new peer: %s\n", peer.ID)
	}

	w.WriteHeader(http.StatusOK)
}

// countHandler returns the current counter value.
func countHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	localCount := counter
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   localCount,
		"node_id": nodeID,
	})
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

	peersMu.RLock()
	currentPeers := make([]string, 0, len(peers))
	for peer := range peers {
		currentPeers = append(currentPeers, peer)
	}
	peersMu.RUnlock()

	for _, peer := range currentPeers {
		wg.Add(1)
		go propagateCounter(peer, newValue, &wg)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   newValue,
		"node_id": nodeID,
	})
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

	fmt.Printf("🔄 Counter synced, new value: %d\n", syncedValue)

	// Propagate only if there was an update
	if shouldPropagate {
		var wg sync.WaitGroup

		peersMu.RLock()
		currentPeers := make([]string, 0, len(peers))
		for peer := range peers {
			currentPeers = append(currentPeers, peer)
		}
		peersMu.RUnlock()

		for _, peer := range currentPeers {
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
	peersMu.Lock()
	delete(peers, peer)
	peersMu.Unlock()
}
