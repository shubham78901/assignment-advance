package counter

import "sync"

// Counter manages the distributed counter state
type Counter struct {
	value int
	mu    sync.Mutex
}

// New creates a new counter instance
func New() *Counter {
	return &Counter{
		value: 0,
	}
}

// Increment increases the counter value and returns the new value
func (c *Counter) Increment() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
	return c.value
}

// GetValue returns the current counter value
func (c *Counter) GetValue() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Sync updates the counter value if the provided value is greater
// Returns true if the value was updated, false otherwise
func (c *Counter) Sync(newValue int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if newValue > c.value {
		c.value = newValue
		return true
	}
	return false
}
