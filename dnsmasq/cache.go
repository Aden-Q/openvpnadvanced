package dnsmasq

import (
	"sync"
	"time"
)

type DNSRecord struct {
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

type Cache struct {
	data map[string]DNSRecord
	mu   sync.RWMutex
	ttl  time.Duration
}

func NewCacheWithTTL(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]DNSRecord),
		ttl:  ttl,
	}
}

func (c *Cache) Get(domain string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	record, ok := c.data[domain]
	if !ok {
		return "", false
	}
	if time.Since(record.Timestamp) > c.ttl {
		// expired
		return "", false
	}
	return record.IP, true
}

func (c *Cache) Set(domain, ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[domain] = DNSRecord{
		IP:        ip,
		Timestamp: time.Now(),
	}
}

func (c *Cache) Raw() map[string]DNSRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copied := make(map[string]DNSRecord)
	for k, v := range c.data {
		copied[k] = v
	}
	return copied
}
