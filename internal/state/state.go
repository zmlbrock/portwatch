// Package state manages persistent port scan state between runs,
// enabling portwatch to detect changes across daemon restarts.
package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// PortState represents the last known state of a single port.
type PortState struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Open     bool   `json:"open"`
	Process  string `json:"process,omitempty"`
}

// Snapshot holds the complete port state captured at a point in time.
type Snapshot struct {
	Timestamp time.Time   `json:"timestamp"`
	Ports     []PortState `json:"ports"`
}

// Store manages reading and writing port state snapshots to disk.
type Store struct {
	mu       sync.RWMutex
	filePath string
	current  *Snapshot
}

// New creates a new Store backed by the given file path.
// If the file does not exist, the store starts with an empty snapshot.
func New(filePath string) (*Store, error) {
	s := &Store{filePath: filePath}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// load reads the snapshot from disk into memory.
func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	s.current = &snap
	return nil
}

// Save persists the given snapshot to disk and updates the in-memory state.
func (s *Store) Save(snap *Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return err
	}
	s.current = snap
	return nil
}

// Current returns the most recently loaded or saved snapshot.
// Returns nil if no snapshot has been loaded yet.
func (s *Store) Current() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Diff compares a new snapshot against the previous one and returns
// lists of ports that were opened or closed between the two snapshots.
func Diff(prev, next *Snapshot) (opened, closed []PortState) {
	if prev == nil {
		return next.Ports, nil
	}

	prevIndex := indexByKey(prev.Ports)
	nextIndex := indexByKey(next.Ports)

	for key, ps := range nextIndex {
		if _, exists := prevIndex[key]; !exists {
			opened = append(opened, ps)
		}
	}

	for key, ps := range prevIndex {
		if _, exists := nextIndex[key]; !exists {
			closed = append(closed, ps)
		}
	}

	return opened, closed
}

// indexByKey builds a lookup map keyed by "protocol:port" for fast diffing.
func indexByKey(ports []PortState) map[string]PortState {
	m := make(map[string]PortState, len(ports))
	for _, p := range ports {
		key := p.Protocol + ":" + itoa(p.Port)
		m[key] = p
	}
	return m
}

// itoa is a minimal integer-to-string helper to avoid importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
