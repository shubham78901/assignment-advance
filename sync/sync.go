package sync

import (
	"assignmet/advance/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func SyncCounter(cfg *config.Config, peer string, value int) error {
	url := fmt.Sprintf("http://%s/sync", peer)
	client := http.Client{Timeout: 2 * time.Second}
	payload := map[string]int{"count": value}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload for %s: %v", peer, err)
	}

	for i := 0; i < 3; i++ {
		fmt.Printf("🔄 Propagating counter %d to %s (Attempt %d)\n", value, peer, i+1)
		resp, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Printf("✅ Counter synced to %s\n", peer)
				return nil
			}
		}
		fmt.Printf("⚠️ Failed to propagate to %s: %v\n", peer, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("final failure: could not propagate counter to %s", peer)
}
