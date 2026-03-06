package bannersstorage

import "sync"

// keyed mutex is a mutex that locks only for same keys
type keyedMutex struct {
	mus map[string]*sync.Mutex
	mu  sync.Mutex
}

func newKeyedMutex() *keyedMutex {
	return &keyedMutex{
		mus: make(map[string]*sync.Mutex),
	}
}

func (m *keyedMutex) Lock(key string) {
	m.mu.Lock()
	mu, ok := m.mus[key]
	if !ok {
		mu = &sync.Mutex{}
		m.mus[key] = mu
	}
	m.mu.Unlock()

	mu.Lock()
}

func (m *keyedMutex) Unlock(key string) {
	m.mu.Lock()
	mu, ok := m.mus[key]
	m.mu.Unlock()

	if !ok {
		panic("unlock not existing key")
	}

	mu.Unlock()
}
