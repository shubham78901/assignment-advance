// File: discovery/service.go
package discovery

import (
	"fmt"
	"go-distributed-system/config"
	"net"
	"sync"
)

var (
	// PeersMu is a read-write mutex to protect access to the peers map.
	PeersMu sync.RWMutex
)

// GetPeerList returns the list of peer IDs as a slice.
func GetPeerList(cfg *config.Config) []string {
	PeersMu.RLock()
	defer PeersMu.RUnlock()

	peerList := make([]string, 0, len(cfg.Peers))
	for p := range cfg.Peers {
		peerList = append(peerList, p)
	}
	return peerList
}

func SetupServiceDiscovery(cfg *config.Config) {
	if !cfg.DiscoveryEnabled {
		fmt.Println("ℹ️ Service discovery is disabled")
		return
	}

	// Start listening for multicast discovery on a separate goroutine
	go listenForDiscovery(cfg)

	// Register with any peers provided in the initial configuration
	for peer := range cfg.Peers {
		go RegisterPeer(cfg, peer)
	}

	fmt.Printf("🔍 Service discovery enabled on port %d\n", cfg.DiscoveryPort)
}

func listenForDiscovery(cfg *config.Config) {
	addr := net.UDPAddr{
		Port: cfg.DiscoveryPort,
		IP:   net.ParseIP("0.0.0.0"), // Listen on all interfaces
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
		if peerID == cfg.NodeID {
			// Ignore our own broadcasts
			continue
		}

		// Add the new peer
		PeersMu.Lock()
		if _, exists := cfg.Peers[peerID]; !exists {
			cfg.Peers[peerID] = true
			fmt.Printf("🔍 Discovered new peer via multicast: %s (from %s)\n", peerID, remoteAddr)

			// Register with the new peer to establish two-way connection
			go RegisterPeer(cfg, peerID)
		}
		PeersMu.Unlock()
	}
}

// RegisterPeer attempts to register with a known peer
func RegisterPeer(cfg *config.Config, peerID string) {
	// Implementation similar to original registerWithPeer function
	// This would make HTTP requests to the peer's /register endpoint
	fmt.Printf("🔄 Registering with peer: %s\n", peerID)
	// Implementation details omitted for brevity
}
