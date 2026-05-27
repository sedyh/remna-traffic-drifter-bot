package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store struct {
	path string
	mu   sync.Mutex
	data map[string]bool
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, data: make(map[string]bool)}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, &s.data)
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func Key(username, kind string) string {
	return username + "|" + kind
}

func (s *Store) ShouldNotify(username, kind string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.data[Key(username, kind)]
}

func (s *Store) MarkNotified(username, kind string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[Key(username, kind)] = true
}

func (s *Store) ClearUser(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := username + "|"
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			delete(s.data, k)
		}
	}
}

func (s *Store) SyncActive(active map[string]struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.data {
		if _, ok := active[k]; !ok {
			delete(s.data, k)
		}
	}
}
