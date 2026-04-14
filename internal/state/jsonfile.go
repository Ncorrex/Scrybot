package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const stateFileName = "card_state.json"

// JSONFileStore persists CardState as a JSON file on disk.
type JSONFileStore struct {
	dir string
}

// NewJSONFileStore creates a JSONFileStore, ensuring the directory exists.
func NewJSONFileStore(dataDir string) (*JSONFileStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &JSONFileStore{dir: dataDir}, nil
}

// Load reads the state from disk. Returns an empty state if the file does not
// yet exist.
func (s *JSONFileStore) Load() (*CardState, error) {
	path := filepath.Join(s.dir, stateFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewCardState(), nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var st CardState
	if err := json.Unmarshal(data, &st); err != nil {
		log.Printf("Warning: corrupt state file, starting fresh: %v", err)
		return NewCardState(), nil
	}
	if st.SeenIDs == nil {
		st.SeenIDs = make(map[string]struct{})
	}
	return &st, nil
}

// Save writes the state to disk atomically via a temp-file + rename.
func (s *JSONFileStore) Save(st *CardState) error {
	path := filepath.Join(s.dir, stateFileName)
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write temp state file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename state file: %w", err)
	}
	return nil
}

