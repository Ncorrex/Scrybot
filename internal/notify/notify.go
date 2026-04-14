package notify

import "context"

// Card is a library-agnostic representation of a card for notifications.
type Card struct {
	ID          string
	Name        string
	SetName     string
	TypeLine    string
	ManaCost    string
	Rarity      string
	OracleText  string
	ImageURL    string
	ScryfallURI string
	Colors      []string
}

// Notifier sends card notifications to an external service.
type Notifier interface {
	Notify(ctx context.Context, card Card) error
}

