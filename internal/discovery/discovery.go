package discovery

import (
	"assignmet/advance/internal/peers"
	"fmt"
	"net"
	"time"
)

// Discovery handles peer discovery via UDP broadcast
type Discovery struct {
	nodeID        string
	discoveryPort int
	peersManager  *peers.PeersManager
}

// New creates a new Discovery instance
func New(nodeID string, discoveryPort int, peersManager *peers.PeersManager) *Discovery {
	return &Discovery{
		nodeID:        nodeID,
		discoveryPort: discoveryPort,
		peersManager:  peersManager,
	}
}

// Listen for UDP broadcast/multicast discovery messages
func (d *Discovery) Listen() {
	addr := net.UDPAddr{
		Port: d.discoveryPort,
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
		if peerID == d.nodeID {
			// Ignore our own broadcasts
			continue
		}

		// Add the new peer
		if d.peersManager.AddPeer(peerID) {
			fmt.Printf("🔍 Discovered new peer via multicast: %s (from %s)\n", peerID, remoteAddr)

			// Register with the new peer to establish two-way connection
			go d.peersManager.RegisterWithPeer(peerID, d)
		}
	}
}

// BroadcastPresence periodically announces this node's presence via UDP
func (d *Discovery) BroadcastPresence() {
	// Instead of using 255.255.255.255, try the Docker bridge network broadcast address
	addr := net.UDPAddr{
		Port: d.discoveryPort,
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
		_, err := conn.Write([]byte(d.nodeID))
		if err != nil {
			fmt.Printf("⚠️ Failed to broadcast presence: %v\n", err)
		} else {
			fmt.Println("📢 Broadcasted presence to network")
		}

		// Wait before next broadcast
		time.Sleep(30 * time.Second)
	}
}

// GetNodeID returns the ID of this node
func (d *Discovery) GetNodeID() string {
	return d.nodeID
}
