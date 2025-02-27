package counter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	Counter int
)

// GetCounter returns the current counter value.
func GetCounter() int {
	mu.Lock()
	defer mu.Unlock()
	return Counter
}

// IncrementCounter increments the counter.
func IncrementCounter() int {
	mu.Lock()
	defer mu.Unlock()
	Counter++
	fmt.Printf("✅ Counter incremented: %d\n", Counter)
	return Counter
}

// SyncCounter updates the counter with a higher value if received from a peer.
func SyncCounter(value int) {
	mu.Lock()
	defer mu.Unlock()
	if value > Counter {
		Counter = value
		fmt.Printf("🔄 Counter synced, new value: %d\n", Counter)
	}
}

// PropagateCounter sends the new counter value to a peer.
func PropagateCounter(peer string, value int, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://%s/sync", peer)
	client := http.Client{Timeout: 2 * time.Second}
	payload := map[string]int{"count": value}
	body, _ := json.Marshal(payload)

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("❌ Failed to propagate counter to %s: %v\n", peer, err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✅ Successfully propagated counter to %s\n", peer)
}
