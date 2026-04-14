package state

import "time"

// CardState tracks which cards have already been notified about.
// SeenIDs holds every card ID from the most recent poll so that new cards
// are detected by set-difference rather than a single cursor.  This is
// robust when multiple sets are being spoiled at the same time and the
// result order is unstable.
type CardState struct {
	SeenIDs   map[string]struct{} `json:"seen_ids"`
	LastCheck time.Time           `json:"last_check"`
}

// NewCardState returns an initialised, empty CardState.
func NewCardState() *CardState {
	return &CardState{SeenIDs: make(map[string]struct{})}
}

// Store abstracts how card state is persisted.
type Store interface {
	Load() (*CardState, error)
	Save(s *CardState) error
}

