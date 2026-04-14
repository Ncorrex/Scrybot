package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- colorForCard ---

func TestColorForCard(t *testing.T) {
	tests := []struct {
		name   string
		colors []string
		want   int
	}{
		{"colorless (nil)", nil, 0x808080},
		{"colorless (empty)", []string{}, 0x808080},
		{"white", []string{"W"}, 0xF0E68C},
		{"blue", []string{"U"}, 0x4169E1},
		{"black", []string{"B"}, 0x1C1C1C},
		{"red", []string{"R"}, 0xFF4500},
		{"green", []string{"G"}, 0x228B22},
		{"two-color gets gold", []string{"W", "U"}, 0xFFD700},
		{"five-color gets gold", []string{"W", "U", "B", "R", "G"}, 0xFFD700},
		{"unknown single color falls back to colourless", []string{"X"}, 0x808080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorForCard(Card{Colors: tt.colors})
			if got != tt.want {
				t.Errorf("colorForCard(%v) = 0x%06X, want 0x%06X", tt.colors, got, tt.want)
			}
		})
	}
}

// --- truncate ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"short string unchanged", "hello", 10, "hello"},
		{"exact max length unchanged", "hello", 5, "hello"},
		{"one over limit appends ellipsis", "abcdef", 5, "ab..."},
		{"empty string unchanged", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
			if len(got) > tt.max {
				t.Errorf("result length %d exceeds max %d", len(got), tt.max)
			}
		})
	}
}

func TestTruncate_Discord1024Limit(t *testing.T) {
	// Simulate a long oracle text hitting Discord's embed field cap.
	input := strings.Repeat("a", 2000)
	got := truncate(input, 1024)
	if len(got) != 1024 {
		t.Errorf("truncated length = %d, want 1024", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("truncated result should end with '...'")
	}
}

// --- Notify (via httptest server) ---

func webhookServer(t *testing.T, body *[]byte, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body != nil {
			*body, _ = io.ReadAll(r.Body)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNotify_SendsPayloadToWebhook(t *testing.T) {
	var captured []byte
	srv := webhookServer(t, &captured, http.StatusNoContent)
	defer srv.Close()

	card := Card{
		Name:        "Lightning Bolt",
		SetName:     "Alpha",
		TypeLine:    "Instant",
		ManaCost:    "{R}",
		Rarity:      "common",
		OracleText:  "Lightning Bolt deals 3 damage to any target.",
		ImageURL:    "https://example.com/bolt.jpg",
		ScryfallURI: "https://scryfall.com/card/alpha/lb",
		Colors:      []string{"R"},
	}

	if err := NewDiscordNotifier(srv.URL).Notify(context.Background(), card); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if len(captured) == 0 {
		t.Fatal("expected non-empty webhook payload")
	}
	if !bytes.Contains(captured, []byte("Lightning Bolt")) {
		t.Errorf("payload does not contain card name; body: %s", captured)
	}
	// Red = 0xFF4500 = 16729344 decimal — embedded as a string in the JSON.
	if !bytes.Contains(captured, []byte("16729344")) {
		t.Errorf("payload does not contain red colour code (16729344); body: %s", captured)
	}
}

func TestNotify_EmptyOracleText_NoOracleField(t *testing.T) {
	var captured []byte
	srv := webhookServer(t, &captured, http.StatusNoContent)
	defer srv.Close()

	card := Card{Name: "Blank Card", OracleText: ""}

	if err := NewDiscordNotifier(srv.URL).Notify(context.Background(), card); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if bytes.Contains(captured, []byte("Oracle Text")) {
		t.Errorf("expected no Oracle Text field when OracleText is empty; body: %s", captured)
	}
}

func TestNotify_LongOracleText_TruncatedInPayload(t *testing.T) {
	var captured []byte
	srv := webhookServer(t, &captured, http.StatusNoContent)
	defer srv.Close()

	card := Card{
		Name:       "Wordy Card",
		OracleText: strings.Repeat("A", 2000),
	}

	if err := NewDiscordNotifier(srv.URL).Notify(context.Background(), card); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	// The raw 2000-char string should NOT appear; the truncated form ending
	// with "..." should.
	if bytes.Contains(captured, []byte(strings.Repeat("A", 2000))) {
		t.Error("expected oracle text to be truncated, but full 2000-char string found in payload")
	}
	if !bytes.Contains(captured, []byte("...")) {
		t.Errorf("expected truncation marker '...' in payload; body: %s", captured)
	}
}

func TestNotify_MulticolorCard_GoldColor(t *testing.T) {
	var captured []byte
	srv := webhookServer(t, &captured, http.StatusNoContent)
	defer srv.Close()

	card := Card{
		Name:   "Azorius Sphinx",
		Colors: []string{"W", "U"},
	}

	if err := NewDiscordNotifier(srv.URL).Notify(context.Background(), card); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	// Gold = 0xFFD700 = 16766720 decimal.
	if !bytes.Contains(captured, []byte("16766720")) {
		t.Errorf("expected gold colour code (16766720) for multi-colour card; body: %s", captured)
	}
}

func TestNotify_WebhookError_ReturnsWrappedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	err := NewDiscordNotifier(srv.URL).Notify(context.Background(), Card{Name: "Test"})
	if err == nil {
		t.Error("expected error when webhook returns 500, got nil")
	}
}

func TestNotify_PayloadIsValidJSON(t *testing.T) {
	var captured []byte
	srv := webhookServer(t, &captured, http.StatusNoContent)
	defer srv.Close()

	card := Card{
		Name:       "Serra Angel",
		SetName:    "Alpha",
		TypeLine:   "Creature — Angel",
		ManaCost:   "{3}{W}{W}",
		Rarity:     "uncommon",
		OracleText: "Flying, vigilance",
		Colors:     []string{"W"},
	}

	if err := NewDiscordNotifier(srv.URL).Notify(context.Background(), card); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	var msg map[string]json.RawMessage
	if err := json.Unmarshal(captured, &msg); err != nil {
		t.Errorf("webhook payload is not valid JSON: %v\nbody: %s", err, captured)
	}
}