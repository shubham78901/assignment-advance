package peers

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	mu    sync.Mutex
	Peers = make(map[string]bool) // Stores active peers
)

// AddPeer adds a new peer to the list.
func AddPeer(peer string) {
	mu.Lock()
	defer mu.Unlock()
	Peers[peer] = true
}

// RemovePeer removes a peer from the list.
func RemovePeer(peer string) {
	mu.Lock()
	defer mu.Unlock()
	delete(Peers, peer)
}

// GetPeers returns the list of current peers.
func GetPeers() []string {
	mu.Lock()
	defer mu.Unlock()
	var peerList []string
	for peer := range Peers {
		peerList = append(peerList, peer)
	}
	return peerList
}

// HealthCheckPeers periodically checks peer health.
func HealthCheckPeers() {
	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		for peer := range Peers {
			url := fmt.Sprintf("http://%s/health", peer)
			resp, err := http.Get(url)
			if err != nil || resp.StatusCode != http.StatusOK {
				delete(Peers, peer)
				fmt.Printf("\U0001F6AB Removed dead peer: %s\n", peer)
			}
		}
		mu.Unlock()
	}
}
