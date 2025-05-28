package pkg

import (
	"fmt"
	"sync"
	"time"
)

// MemoryStore implements APIKeyStore using in-memory storage
type MemoryStore struct {
	data          map[string]*APIKeyEntry
	expiry        map[string]time.Time
	mu            sync.RWMutex
	defaultExpiry time.Duration
}

// NewMemoryStore creates a new memory-based API key store
func NewMemoryStore(defaultExpiry time.Duration) *MemoryStore {
	return &MemoryStore{
		data:          make(map[string]*APIKeyEntry),
		expiry:        make(map[string]time.Time),
		defaultExpiry: defaultExpiry,
	}
}

func (m *MemoryStore) SaveAPIKey(email, apiKey string, expiry ...time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiryDuration := m.defaultExpiry
	if len(expiry) > 0 {
		expiryDuration = expiry[0]
	}

	entry := &APIKeyEntry{
		APIKey:    apiKey,
		CreatedAt: time.Now(),
	}

	m.data[email] = entry
	m.expiry[email] = time.Now().Add(expiryDuration)
	return nil
}

func (m *MemoryStore) GetAPIKey(email string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[email]
	if !exists {
		return "", fmt.Errorf("no API key found for email: %s", email)
	}

	if time.Now().After(m.expiry[email]) {
		delete(m.data, email)
		delete(m.expiry, email)
		return "", fmt.Errorf("API key expired for email: %s", email)
	}

	return entry.APIKey, nil
}

func (m *MemoryStore) DeleteAPIKey(email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, email)
	delete(m.expiry, email)
	return nil
}

func (m *MemoryStore) HasAPIKey(email string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[email]
	if !exists {
		return false, nil
	}

	if time.Now().After(m.expiry[email]) {
		delete(m.data, email)
		delete(m.expiry, email)
		return false, nil
	}

	return entry != nil, nil
}

func (m *MemoryStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = nil
	m.expiry = nil
	return nil
}
