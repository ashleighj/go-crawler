package crawler

import (
	"sync"
)

// ConcurrentCounter uses a mutex to manage changes to a 
// counter that is potentially updated by multiple goroutines
type ConcurrentCounter struct {
	mutex sync.Mutex
	count int
}

// NewConcurrentCounter creates, inits and returns a new ConcurrentCounter struct
func NewConcurrentCounter() (c *ConcurrentCounter) {
	return &ConcurrentCounter{
		count: 0,
	}
}

// GetCount retrieves the current counter value before another goroutine can update it
func (c *ConcurrentCounter) GetCount() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.count
}

// Add increments the counter
func (c *ConcurrentCounter) Add(amount int) {
	c.modify(amount)
}

// Subtract decrements the counter
func (c *ConcurrentCounter) Subtract(amount int) {
	c.modify(-amount)
}

func (c *ConcurrentCounter) modify(amount int) {
	c.mutex.Lock()
	c.count += amount
	c.mutex.Unlock()
}

// ConcurrentMap uses a mutex to manage changes to a 
// map that is potentially updated by multiple goroutines
type ConcurrentMap struct {
	mutex sync.Mutex
	data  map[string]int
}

// NewConcurrentMap creates, inits and returns a new ConcurrentMap struct
func NewConcurrentMap() (c *ConcurrentMap) {
	return &ConcurrentMap{
		data: make(map[string]int),
	}
}

// Add adds an entry to the ConcurrentMap's map
func (c *ConcurrentMap) Add(key string, val int) {
	c.mutex.Lock()
	if _, ok := c.data[key]; !ok {
		c.data[key] = val
	}
	c.mutex.Unlock()
}

// KeyExists checks whether a given key exists in the 
// ConcurrentMap's map before another goroutine can update it
func (c *ConcurrentMap) KeyExists(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.data[key]
	return ok
}
