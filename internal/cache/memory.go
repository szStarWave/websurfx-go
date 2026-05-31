package cache

import (
	"sync"
	"time"

	"github.com/szStarWave/websurfx-go/internal/search"
)

type Memory struct {
	ttl   time.Duration
	mu    sync.RWMutex
	items map[string]entry
}

type entry struct {
	expires time.Time
	value   search.Response
}

func NewMemory(ttl time.Duration) *Memory {
	return &Memory{
		ttl:   ttl,
		items: make(map[string]entry),
	}
}

func (m *Memory) Get(key string) (search.Response, bool) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()
	if !ok || time.Now().After(item.expires) {
		if ok {
			m.mu.Lock()
			delete(m.items, key)
			m.mu.Unlock()
		}
		return search.Response{}, false
	}
	item.value.Cached = true
	return item.value, true
}

func (m *Memory) Set(key string, value search.Response) {
	m.mu.Lock()
	m.items[key] = entry{
		expires: time.Now().Add(m.ttl),
		value:   value,
	}
	m.mu.Unlock()
}
