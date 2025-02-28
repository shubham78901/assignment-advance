package peers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DiscoveryService interface for any service that can provide node ID
type DiscoveryService interface {
	GetNodeID() string
}

// PeersManager manages peer connections
type PeersManager struct {
	nodeID  string
	peers   map[string]bool
	peersMu sync.RWMutex
}

// New creates a new PeersManager
func New(nodeID string) *PeersManager {
	return &PeersManager{
		nodeID: nodeID,
		peers:  make(map[string]bool),
	}
}
func (pm *PeersManager) SetupInitialPeers(peerList string) {
	if peerList != "" {
		for _, peer := range strings.Split(peerList, ",") {
			if peer != pm.nodeID {
				pm.AddPeer(peer)
			}
		}
	}
}

// AddPeer adds a new peer if it doesn't already exist
// Returns true if the peer was added, false if it already existed
func (pm *PeersManager) AddPeer(peerID string) bool {
	pm.peersMu.Lock()
	defer pm.peersMu.Unlock()

	if _, exists := pm.peers[peerID]; !exists {
		pm.peers[peerID] = true
		return true
	}
	return false
}

// RemovePeer removes a peer
func (pm *PeersManager) RemovePeer(peerID string) {
	pm.peersMu.Lock()
	defer pm.peersMu.Unlock()

	delete(pm.peers, peerID)
}

// GetPeerList returns the list of peer IDs as a slice
func (pm *PeersManager) GetPeerList() []string {
	pm.peersMu.RLock()
	defer pm.peersMu.RUnlock()

	peerList := make([]string, 0, len(pm.peers))
	for p := range pm.peers {
		peerList = append(peerList, p)
	}
	return peerList
}

// HealthCheckPeers periodically checks if peers are alive
func (pm *PeersManager) HealthCheckPeers() {
	for {
		time.Sleep(5 * time.Second)

		pm.peersMu.RLock()
		currentPeers := make([]string, 0, len(pm.peers))
		for peer := range pm.peers {
			currentPeers = append(currentPeers, peer)
		}
		pm.peersMu.RUnlock()

		for _, peer := range currentPeers {
			url := fmt.Sprintf("http://%s/health", peer)
			_, err := http.Get(url)
			if err != nil {
				pm.RemovePeer(peer)
				fmt.Printf("❌ Removed dead peer: %s\n", peer)
			}
		}
	}
}

// RegisterWithPeer attempts to register with a known peer
func (pm *PeersManager) RegisterWithPeer(peerID string, service DiscoveryService) {
	url := fmt.Sprintf("http://%s/register", peerID)
	payload := map[string]string{"id": service.GetNodeID()}
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
		go pm.FetchPeersFromPeer(peerID, service)
	} else {
		fmt.Printf("⚠️ Peer %s returned status %d during registration\n", peerID, resp.StatusCode)
	}
}

// FetchPeersFromPeer gets the peer list from a known peer
func (pm *PeersManager) FetchPeersFromPeer(peerID string, service DiscoveryService) {
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

	// Add new peers to our list and register with them
	for _, newPeer := range peerList {
		if newPeer != service.GetNodeID() && pm.AddPeer(newPeer) {
			fmt.Printf("🔍 Discovered new peer via %s: %s\n", peerID, newPeer)

			// Register with the new peer
			go pm.RegisterWithPeer(newPeer, service)
		}
	}
}
