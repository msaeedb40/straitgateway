// Package ipam — store.go
// Implements crash-safe, atomic, disk-backed persistence for IP allocations.
package ipam

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// DefaultStorePath is the local filesystem store path.
const DefaultStorePath = "/var/lib/cni/straitgateway/allocations.json"

// PodAllocation tracks the IP ownership details for a single container/pod.
type PodAllocation struct {
	ContainerID  string `json:"containerID"`
	PodName      string `json:"podName,omitempty"`
	PodNamespace string `json:"podNamespace,omitempty"`
	IfName       string `json:"ifName,omitempty"`
	IPv4         string `json:"ipv4,omitempty"` // e.g. "10.244.1.5/24"
	IPv6         string `json:"ipv6,omitempty"` // e.g. "fd00:10:244::5/64"
	AllocatedAt  int64  `json:"allocatedAt"`
}

// State represents the complete serialized IPAM allocation table.
type State struct {
	Version     int                      `json:"version"`
	LastUpdated int64                    `json:"lastUpdated"`
	Allocations map[string]PodAllocation `json:"allocations"` // ContainerID -> PodAllocation
}

// Store handles crash-safe persistence of IP allocations.
// Thread-safe within a process via sync.Mutex, and cross-process safe via flock.
type Store struct {
	mu   sync.Mutex
	path string
}

// NewStore creates a new Store instance.
func NewStore(path string) *Store {
	if path == "" {
		path = DefaultStorePath
	}
	return &Store{path: path}
}

func (s *Store) acquireFileLock() (*os.File, error) {
	lockPath := s.path + ".lock"
	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create IPAM lock directory %s: %w", dir, err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open IPAM lock file %s: %w", lockPath, err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to acquire flock on %s: %w", lockPath, err)
	}
	return f, nil
}

func (s *Store) releaseFileLock(f *os.File) {
	if f != nil {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
	}
}

// Load reads allocation state from disk.
func (s *Store) Load() (*State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockFile, err := s.acquireFileLock()
	if err != nil {
		return nil, err
	}
	defer s.releaseFileLock(lockFile)

	return s.loadLocked()
}

func (s *Store) loadLocked() (*State, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &State{
			Version:     1,
			LastUpdated: time.Now().Unix(),
			Allocations: make(map[string]PodAllocation),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read IPAM store at %s: %w", s.path, err)
	}

	// Support backwards compatibility with legacy format: map[string]string
	var state State
	if err := json.Unmarshal(data, &state); err == nil && state.Allocations != nil {
		return &state, nil
	}

	// Try legacy unmarshal: {"allocations": {"pod-1": "10.244.1.5"}}
	var legacy struct {
		Allocations map[string]string `json:"allocations"`
	}
	if errLegacy := json.Unmarshal(data, &legacy); errLegacy == nil && legacy.Allocations != nil {
		upgraded := &State{
			Version:     1,
			LastUpdated: time.Now().Unix(),
			Allocations: make(map[string]PodAllocation),
		}
		for k, v := range legacy.Allocations {
			upgraded.Allocations[k] = PodAllocation{
				ContainerID: k,
				IPv4:        v,
				AllocatedAt: time.Now().Unix(),
			}
		}
		return upgraded, nil
	}

	return nil, fmt.Errorf("failed to parse IPAM store at %s: corrupt data", s.path)
}

// Save atomically writes allocation state to disk with full fsync and directory sync.
func (s *Store) Save(state *State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockFile, err := s.acquireFileLock()
	if err != nil {
		return err
	}
	defer s.releaseFileLock(lockFile)

	return s.saveLocked(state)
}

func (s *Store) saveLocked(state *State) error {
	state.LastUpdated = time.Now().Unix()
	if state.Version == 0 {
		state.Version = 1
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create IPAM state directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode IPAM state: %w", err)
	}

	// Write to temporary file with unique suffix
	tmpFile := fmt.Sprintf("%s.tmp.%d.%d", s.path, os.Getpid(), time.Now().UnixNano())
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temporary IPAM file: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to write IPAM temporary file: %w", err)
	}

	// fsync data to storage
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to fsync IPAM file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to close IPAM temporary file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, s.path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to atomically rename IPAM store: %w", err)
	}

	// fsync parent directory to guarantee directory entry persistence
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

// Get returns the allocation for a specific container/pod if present.
func (s *Store) Get(containerID string) (PodAllocation, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockFile, err := s.acquireFileLock()
	if err != nil {
		return PodAllocation{}, false, err
	}
	defer s.releaseFileLock(lockFile)

	state, err := s.loadLocked()
	if err != nil {
		return PodAllocation{}, false, err
	}
	alloc, exists := state.Allocations[containerID]
	return alloc, exists, nil
}

// Put records or updates an allocation for a container.
// Holds lock across the full read-modify-write cycle.
func (s *Store) Put(alloc PodAllocation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockFile, err := s.acquireFileLock()
	if err != nil {
		return err
	}
	defer s.releaseFileLock(lockFile)

	state, err := s.loadLocked()
	if err != nil {
		return err
	}
	if alloc.AllocatedAt == 0 {
		alloc.AllocatedAt = time.Now().Unix()
	}
	state.Allocations[alloc.ContainerID] = alloc
	return s.saveLocked(state)
}

// Delete removes an allocation for a container.
// Holds lock across the full read-modify-write cycle.
func (s *Store) Delete(containerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockFile, err := s.acquireFileLock()
	if err != nil {
		return err
	}
	defer s.releaseFileLock(lockFile)

	state, err := s.loadLocked()
	if err != nil {
		return err
	}
	delete(state.Allocations, containerID)
	return s.saveLocked(state)
}

// List returns all active pod allocations.
func (s *Store) List() ([]PodAllocation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lockFile, err := s.acquireFileLock()
	if err != nil {
		return nil, err
	}
	defer s.releaseFileLock(lockFile)

	state, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	res := make([]PodAllocation, 0, len(state.Allocations))
	for _, a := range state.Allocations {
		res = append(res, a)
	}
	return res, nil
}

