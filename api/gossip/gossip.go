package gossip

import (
	"assignmet/advance/api/peers"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GossipNewPeer informs all known peers about a newly discovered peer.
func GossipNewPeer(newPeer string) {
	peersList := peers.GetPeers() // Get a snapshot of peers
	for _, peer := range peersList {
		if peer == newPeer {
			continue
		}

		url := fmt.Sprintf("http://%s/register", peer)
		client := http.Client{Timeout: 2 * time.Second}
		payload := map[string]string{"id": newPeer}
		body, _ := json.Marshal(payload)

		for i := 0; i < 3; i++ {
			resp, err := client.Post(url, "application/json", bytes.NewReader(body))
			if err == nil && resp.StatusCode == http.StatusOK {
				fmt.Printf("🔄 Peer %s notified about %s\n", peer, newPeer)
				resp.Body.Close()
				return
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
}

// RegisterWithPeers registers this node with existing peers.
func RegisterWithPeers(port string) {
	self := fmt.Sprintf("localhost:%s", port)
	client := http.Client{Timeout: 2 * time.Second}
	payload := map[string]string{"id": self}
	body, _ := json.Marshal(payload)

	peersList := peers.GetPeers() // Get snapshot of peers
	for _, peer := range peersList {
		url := fmt.Sprintf("http://%s/register", peer)
		resp, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err == nil && resp.StatusCode == http.StatusOK {
			var peerList []string
			json.NewDecoder(resp.Body).Decode(&peerList)
			resp.Body.Close()

			for _, p := range peerList {
				peers.AddPeer(p)
			}

			fmt.Printf("🔗 Registered with peer %s and received peers: %v\n", peer, peerList)
		}
	}
}
