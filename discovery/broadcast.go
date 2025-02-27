package discovery

import (
	"fmt"
	"go-distributed-system/config"

	"net"
	"time"
)

func BroadcastPresence(cfg *config.Config) {
	// Use the Docker network's broadcast address
	addr := net.UDPAddr{
		Port: cfg.DiscoveryPort,
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
		_, err := conn.Write([]byte(cfg.NodeID))
		if err != nil {
			fmt.Printf("⚠️ Failed to broadcast presence: %v\n", err)
		} else {
			fmt.Println("📢 Broadcasted presence to network")
		}

		// Wait before next broadcast
		time.Sleep(30 * time.Second)
	}
}
