package poller

import (
	"context"
	"errors"
	"testing"
	"time"

	scryfall "github.com/BlueMonday/go-scryfall"

	"Scrybot/internal/notify"
	"Scrybot/internal/state"
)

// --------------------------------------------------------------------------
// Test doubles
// --------------------------------------------------------------------------

type mockSearcher struct {
	cards []scryfall.Card
	err   error
}

func (m *mockSearcher) SearchCards(_ context.Context, _ string, _ scryfall.SearchCardsOptions) (scryfall.CardListResponse, error) {
	if m.err != nil {
		return scryfall.CardListResponse{}, m.err
	}
	return scryfall.CardListResponse{Cards: m.cards}, nil
}

type mockNotifier struct {
	notified []notify.Card
	err      error
}

func (m *mockNotifier) Notify(_ context.Context, card notify.Card) error {
	if m.err != nil {
		return m.err
	}
	m.notified = append(m.notified, card)
	return nil
}

type mockStore struct {
	state   *state.CardState
	saved   *state.CardState
	loadErr error
	saveErr error
}

func (m *mockStore) Load() (*state.CardState, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	if m.state == nil {
		return state.NewCardState(), nil
	}
	return m.state, nil
}

func (m *mockStore) Save(s *state.CardState) error {
	m.saved = s
	return m.saveErr
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func makeCard(id, name string) scryfall.Card {
	return scryfall.Card{
		ID:          id,
		Name:        name,
		SetName:     "Test Set",
		TypeLine:    "Creature",
		ManaCost:    "{1}",
		Rarity:      "common",
		ScryfallURI: "https://scryfall.com/card/test/" + id,
		ImageURIs:   &scryfall.ImageURIs{Normal: "https://example.com/" + id + ".jpg"},
	}
}

// newFastPoller creates a Poller with zero rate-limit delay so tests are not
// slowed down by the 500 ms real-world throttle.
func newFastPoller(searcher cardSearcher, notifier notify.Notifier, store state.Store) *Poller {
	p := NewPoller(searcher, notifier, store, "test-query")
	p.rateLimitDelay = 0
	return p
}

// seenState builds a CardState with the given IDs already seen.
func seenState(ids ...string) *state.CardState {
	st := state.NewCardState()
	for _, id := range ids {
		st.SeenIDs[id] = struct{}{}
	}
	return st
}

// --------------------------------------------------------------------------
// Poll — lifecycle scenarios
// --------------------------------------------------------------------------

func TestPoll_FirstRun_NoNotificationsAllCardsSaved(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{
		makeCard("id-1", "Alpha"),
		makeCard("id-2", "Beta"),
	}}
	notifier := &mockNotifier{}
	store := &mockStore{state: state.NewCardState()} // empty SeenIDs triggers first-run path

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 0 {
		t.Errorf("first run: expected 0 notifications, got %d", len(notifier.notified))
	}
	if store.saved == nil {
		t.Fatal("first run: expected state to be saved")
	}
	if len(store.saved.SeenIDs) != 2 {
		t.Errorf("first run: expected 2 seen IDs saved, got %d", len(store.saved.SeenIDs))
	}
	if _, ok := store.saved.SeenIDs["id-1"]; !ok {
		t.Error("first run: id-1 missing from saved SeenIDs")
	}
}

func TestPoll_NewCard_IsNotified(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{
		makeCard("old-id", "OldCard"),
		makeCard("new-id", "NewCard"),
	}}
	notifier := &mockNotifier{}
	store := &mockStore{state: seenState("old-id")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.notified))
	}
	if notifier.notified[0].Name != "NewCard" {
		t.Errorf("notified card = %q, want NewCard", notifier.notified[0].Name)
	}
}

func TestPoll_NoNewCards_NoNotifications(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{
		makeCard("a", "A"),
		makeCard("b", "B"),
	}}
	notifier := &mockNotifier{}
	store := &mockStore{state: seenState("a", "b")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 0 {
		t.Errorf("expected 0 notifications for all-known cards, got %d", len(notifier.notified))
	}
}

func TestPoll_MultipleNewCards_AllNotified(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{
		makeCard("known", "Known"),
		makeCard("new-1", "New1"),
		makeCard("new-2", "New2"),
		makeCard("new-3", "New3"),
	}}
	notifier := &mockNotifier{}
	store := &mockStore{state: seenState("known")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(notifier.notified))
	}
}

// TestPoll_NewCardsNotifiedChronologically verifies that the poller walks the
// card list in reverse (oldest first) before sending notifications.
func TestPoll_NewCardsNotifiedChronologically(t *testing.T) {
	// API response is newest-first (index 0 = newest).
	// The sentinel ("known") keeps us out of first-run mode.
	searcher := &mockSearcher{cards: []scryfall.Card{
		makeCard("known", "Known"),   // index 0 — already seen
		makeCard("new-newest", "Newest"), // index 1
		makeCard("new-middle", "Middle"), // index 2
		makeCard("new-oldest", "Oldest"), // index 3 — will be notified first
	}}
	notifier := &mockNotifier{}
	store := &mockStore{state: seenState("known")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(notifier.notified))
	}
	order := []string{"Oldest", "Middle", "Newest"}
	for i, want := range order {
		if notifier.notified[i].Name != want {
			t.Errorf("notification[%d] = %q, want %q (oldest-first order)", i, notifier.notified[i].Name, want)
		}
	}
}

func TestPoll_StateSavedWithUpdatedSeenIDsAndTimestamp(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{makeCard("x", "X")}}
	notifier := &mockNotifier{}
	store := &mockStore{state: state.NewCardState()}

	before := time.Now().Add(-time.Second)
	newFastPoller(searcher, notifier, store).Poll(context.Background())
	after := time.Now().Add(time.Second)

	if store.saved == nil {
		t.Fatal("expected state to be saved")
	}
	if _, ok := store.saved.SeenIDs["x"]; !ok {
		t.Error("saved state should include card 'x' in SeenIDs")
	}
	if store.saved.LastCheck.Before(before) || store.saved.LastCheck.After(after) {
		t.Errorf("LastCheck = %v, want between %v and %v", store.saved.LastCheck, before, after)
	}
}

// --------------------------------------------------------------------------
// Poll — error-resilience scenarios
// --------------------------------------------------------------------------

func TestPoll_StateLoadError_ReturnsEarlyWithoutNotifyingOrSaving(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{makeCard("x", "X")}}
	notifier := &mockNotifier{}
	store := &mockStore{loadErr: errors.New("disk read failure")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 0 {
		t.Error("expected no notifications when state load fails")
	}
	if store.saved != nil {
		t.Error("expected no state save when load fails")
	}
}

func TestPoll_SearchError_ReturnsEarlyWithoutSaving(t *testing.T) {
	searcher := &mockSearcher{err: errors.New("Scryfall API unavailable")}
	notifier := &mockNotifier{}
	store := &mockStore{state: seenState("x")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if len(notifier.notified) != 0 {
		t.Error("expected no notifications when search fails")
	}
	if store.saved != nil {
		t.Error("expected no state save when search fails")
	}
}

func TestPoll_EmptySearchResults_ReturnsEarlyWithoutSaving(t *testing.T) {
	searcher := &mockSearcher{cards: []scryfall.Card{}}
	notifier := &mockNotifier{}
	store := &mockStore{state: seenState("x")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if store.saved != nil {
		t.Error("expected no state save when search returns no cards")
	}
}

func TestPoll_NotifyError_ContinuesAndSavesState(t *testing.T) {
	// Even when every notification fails, the poller should persist the new
	// seen-ID set so the next cycle doesn't re-attempt the same cards.
	searcher := &mockSearcher{cards: []scryfall.Card{
		makeCard("old", "Old"),
		makeCard("new-1", "New1"),
		makeCard("new-2", "New2"),
	}}
	notifier := &mockNotifier{err: errors.New("webhook unreachable")}
	store := &mockStore{state: seenState("old")}

	newFastPoller(searcher, notifier, store).Poll(context.Background())

	if store.saved == nil {
		t.Fatal("expected state to be saved despite notification errors")
	}
	if len(store.saved.SeenIDs) != 3 {
		t.Errorf("expected 3 seen IDs in saved state, got %d", len(store.saved.SeenIDs))
	}
}

// --------------------------------------------------------------------------
// toNotifyCard
// --------------------------------------------------------------------------

func TestToNotifyCard_BasicFieldMapping(t *testing.T) {
	c := scryfall.Card{
		ID:          "uuid-123",
		Name:        "Lightning Bolt",
		SetName:     "Alpha",
		TypeLine:    "Instant",
		ManaCost:    "{R}",
		Rarity:      "common",
		OracleText:  "Lightning Bolt deals 3 damage to any target.",
		ScryfallURI: "https://scryfall.com/card/alpha/lb",
		ImageURIs:   &scryfall.ImageURIs{Normal: "https://example.com/bolt.jpg"},
		Colors:      []scryfall.Color{scryfall.ColorRed},
	}

	got := toNotifyCard(&c)

	checks := map[string][2]string{
		"ID":          {"uuid-123", got.ID},
		"Name":        {"Lightning Bolt", got.Name},
		"SetName":     {"Alpha", got.SetName},
		"TypeLine":    {"Instant", got.TypeLine},
		"ManaCost":    {"{R}", got.ManaCost},
		"Rarity":      {"common", got.Rarity},
		"OracleText":  {"Lightning Bolt deals 3 damage to any target.", got.OracleText},
		"ScryfallURI": {"https://scryfall.com/card/alpha/lb", got.ScryfallURI},
		"ImageURL":    {"https://example.com/bolt.jpg", got.ImageURL},
	}
	for field, pair := range checks {
		if pair[0] != pair[1] {
			t.Errorf("%s = %q, want %q", field, pair[1], pair[0])
		}
	}
	if len(got.Colors) != 1 || got.Colors[0] != "R" {
		t.Errorf("Colors = %v, want [R]", got.Colors)
	}
}

func TestToNotifyCard_DoubleFacedCard_UsesFrontFaceImage(t *testing.T) {
	c := scryfall.Card{
		ID:        "dfc-uuid",
		Name:      "Delver of Secrets // Insectile Aberration",
		ImageURIs: &scryfall.ImageURIs{}, // empty top-level — typical for DFCs
		CardFaces: []scryfall.CardFace{
			{
				Name:      "Delver of Secrets",
				ImageURIs: scryfall.ImageURIs{Normal: "https://example.com/delver-front.jpg"},
			},
			{
				Name:      "Insectile Aberration",
				ImageURIs: scryfall.ImageURIs{Normal: "https://example.com/delver-back.jpg"},
			},
		},
	}

	got := toNotifyCard(&c)

	if got.ImageURL != "https://example.com/delver-front.jpg" {
		t.Errorf("DFC ImageURL = %q, want front-face URL", got.ImageURL)
	}
}

func TestToNotifyCard_MulticolorCard_ColorsPreserved(t *testing.T) {
	c := scryfall.Card{
		Colors: []scryfall.Color{scryfall.ColorWhite, scryfall.ColorBlue},
	}

	got := toNotifyCard(&c)

	if len(got.Colors) != 2 {
		t.Fatalf("Colors len = %d, want 2", len(got.Colors))
	}
	if got.Colors[0] != "W" || got.Colors[1] != "U" {
		t.Errorf("Colors = %v, want [W U]", got.Colors)
	}
}

func TestToNotifyCard_ColorlessCard_EmptyColorsSlice(t *testing.T) {
	c := scryfall.Card{
		Colors:    []scryfall.Color{},
		ImageURIs: &scryfall.ImageURIs{Normal: "https://example.com/wastes.jpg"},
	}

	got := toNotifyCard(&c)

	if len(got.Colors) != 0 {
		t.Errorf("Colors = %v, want empty slice for colourless card", got.Colors)
	}
}