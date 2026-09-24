// Package state maintains the in-memory observed and desired state cache for the SG Controller.
package state

import (
	"sync"
)

// Store tracks desired and observed cluster networking objects.
type Store struct {
	mu sync.RWMutex

	// Generations track object updates.
	generations map[string]int64
}

// NewStore creates a new controller state store.
func NewStore() *Store {
	return &Store{
		generations: make(map[string]int64),
	}
}

// SetGeneration records a generation for a given resource key.
func (s *Store) SetGeneration(key string, gen int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.generations[key] = gen
}

// GetGeneration returns the current generation for a resource key.
func (s *Store) GetGeneration(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	gen, exists := s.generations[key]
	return gen, exists
}
