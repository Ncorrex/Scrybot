package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewJSONFileStore_CreatesNestedDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "subdir")
	store, err := NewJSONFileStore(dir)
	if err != nil {
		t.Fatalf("NewJSONFileStore(%q): %v", dir, err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected directory %q to be created", dir)
	}
}

func TestLoad_EmptyStateWhenFileAbsent(t *testing.T) {
	store, _ := NewJSONFileStore(t.TempDir())

	st, err := store.Load()
	if err != nil {
		t.Fatalf("Load on fresh store: %v", err)
	}
	if len(st.SeenIDs) != 0 {
		t.Errorf("SeenIDs len = %d, want 0", len(st.SeenIDs))
	}
	if st.SeenIDs == nil {
		t.Error("SeenIDs should be an initialised map, not nil")
	}
}

func TestLoad_FreshStateWhenFileIsCorrupt(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONFileStore(dir)

	if err := os.WriteFile(filepath.Join(dir, stateFileName), []byte("{ not valid json "), 0644); err != nil {
		t.Fatal(err)
	}

	st, err := store.Load()
	if err != nil {
		t.Fatalf("Load on corrupt file should not return error: %v", err)
	}
	if len(st.SeenIDs) != 0 {
		t.Errorf("expected empty SeenIDs after corrupt file recovery, got %d entries", len(st.SeenIDs))
	}
}

func TestLoad_InitializesNilSeenIDs(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONFileStore(dir)

	// Write JSON that deserialises to a nil map.
	data := []byte(`{"seen_ids":null,"last_check":"2026-01-01T00:00:00Z"}`)
	if err := os.WriteFile(filepath.Join(dir, stateFileName), data, 0644); err != nil {
		t.Fatal(err)
	}

	st, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if st.SeenIDs == nil {
		t.Error("Load should upgrade nil SeenIDs to an empty map")
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONFileStore(dir)

	ts := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)
	original := &CardState{
		SeenIDs: map[string]struct{}{
			"abc-111": {},
			"def-222": {},
		},
		LastCheck: ts,
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}

	if len(loaded.SeenIDs) != 2 {
		t.Errorf("SeenIDs len = %d, want 2", len(loaded.SeenIDs))
	}
	for _, id := range []string{"abc-111", "def-222"} {
		if _, ok := loaded.SeenIDs[id]; !ok {
			t.Errorf("SeenIDs missing %q after round-trip", id)
		}
	}
	if !loaded.LastCheck.Equal(ts) {
		t.Errorf("LastCheck = %v, want %v", loaded.LastCheck, ts)
	}
}

func TestSaveLoad_OverwritesPreviousState(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONFileStore(dir)

	first := &CardState{SeenIDs: map[string]struct{}{"old": {}}}
	if err := store.Save(first); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	second := &CardState{SeenIDs: map[string]struct{}{"new-1": {}, "new-2": {}}}
	if err := store.Save(second); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := loaded.SeenIDs["old"]; ok {
		t.Error("overwritten state should not contain old IDs")
	}
	if len(loaded.SeenIDs) != 2 {
		t.Errorf("SeenIDs len = %d after overwrite, want 2", len(loaded.SeenIDs))
	}
}

func TestSave_AtomicWrite_NoTempFileAfterSuccess(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONFileStore(dir)

	if err := store.Save(NewCardState()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	tmpPath := filepath.Join(dir, stateFileName+".tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should be removed after a successful atomic Save")
	}

	realPath := filepath.Join(dir, stateFileName)
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		t.Error("state file should exist after Save")
	}
}

func TestSave_WritesValidJSON(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONFileStore(dir)

	st := &CardState{
		SeenIDs:   map[string]struct{}{"x": {}},
		LastCheck: time.Now().UTC(),
	}
	if err := store.Save(st); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, stateFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var check CardState
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("state file contains invalid JSON: %v\ncontents: %s", err, data)
	}
}