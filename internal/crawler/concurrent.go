package crawler

import (
	"sync"
)

type ConcurrentCounter struct {
	mutex sync.Mutex
	count int
}

func NewConcurrentCounter() (c *ConcurrentCounter) {
	return &ConcurrentCounter{
		count: 0,
	}
}

func (c *ConcurrentCounter) GetCount() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.count
}

func (c *ConcurrentCounter) Add(amount int) {
	c.modify(amount)
}

func (c *ConcurrentCounter) Subtract(amount int) {
	c.modify(-amount)
}

func (c *ConcurrentCounter) modify(amount int) {
	c.mutex.Lock()
	c.count += amount
	c.mutex.Unlock()
}

type ConcurrentMap struct {
	mutex sync.Mutex
	data  map[string]int
}

func NewConcurrentMap() (c *ConcurrentMap) {
	return &ConcurrentMap{
		data: make(map[string]int),
	}
}

func (c *ConcurrentMap) Add(key string, val int) {
	c.mutex.Lock()
	if _, ok := c.data[key]; !ok {
		c.data[key] = val
	}
	c.mutex.Unlock()
}

func (c *ConcurrentMap) KeyExists(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.data[key]
	return ok
}
