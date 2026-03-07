package bannersstorage

import "sync"

// mutexRef wraps a Mutex with a count of how many goroutines are using or waiting for it.
type mutexRef struct {
	mu    sync.Mutex
	count int
}

// keyedMutex is a mutex that locks only for the same keys, without leaking memory.
type keyedMutex struct {
	mus map[string]*mutexRef
	mu  sync.Mutex
}

func newKeyedMutex() *keyedMutex {
	return &keyedMutex{
		mus: make(map[string]*mutexRef),
	}
}

func (m *keyedMutex) Lock(key string) {
	m.mu.Lock()
	ref, ok := m.mus[key]
	if !ok {
		ref = &mutexRef{}
		m.mus[key] = ref
	}
	// Increment the counter to register that this goroutine is in line for the lock
	ref.count++
	m.mu.Unlock()

	// Block until we actually get the lock for this specific key
	ref.mu.Lock()
}

func (m *keyedMutex) Unlock(key string) {
	m.mu.Lock()
	ref, ok := m.mus[key]
	if !ok {
		m.mu.Unlock() // Always release the global lock before a panic!
		panic("unlocking non-existing key")
	}

	// Decrement the counter because this goroutine is done with it
	ref.count--

	// If the count reaches 0, NO other goroutines are waiting for this key.
	// It is now 100% safe to delete it from the map.
	if ref.count == 0 {
		delete(m.mus, key)
	}
	m.mu.Unlock()

	// Finally, release the actual lock for this key
	ref.mu.Unlock()
}
