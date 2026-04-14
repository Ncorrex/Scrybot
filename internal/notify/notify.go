package notify

import "context"

// Card is a library-agnostic representation of a card for notifications.
type Card struct {
	ID          string
	Name        string
	Rarity      string
	ImageURL    string
	ScryfallURI string
	Prices	  	Prices
	Colors      []string
}

type Prices struct {
	USD	string
	EUR	string
}

// Notifier sends card notifications to an external service.
type Notifier interface {
	Notify(ctx context.Context, card Card) error
}

