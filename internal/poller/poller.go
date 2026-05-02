package poller

import (
	"context"
	"log"
	"time"

	scryfall "github.com/BlueMonday/go-scryfall"

	"Scrybot/internal/notify"
	"Scrybot/internal/state"
)

// cardSearcher is the subset of the Scryfall API used by Poller.
// *scryfall.Client satisfies this interface automatically.
type cardSearcher interface {
	SearchCards(ctx context.Context, query string, opts scryfall.SearchCardsOptions) (scryfall.CardListResponse, error)
}

// Poller performs the poll-diff-notify cycle against Scryfall.
type Poller struct {
	client         cardSearcher
	notifier       notify.Notifier
	store          state.Store
	query          string
	rateLimitDelay time.Duration
}

// NewPoller creates a new Poller with the given dependencies.
func NewPoller(client cardSearcher, notifier notify.Notifier, store state.Store, query string) *Poller {
	return &Poller{
		client:         client,
		notifier:       notifier,
		store:          store,
		query:          query,
		rateLimitDelay: 500 * time.Millisecond,
	}
}

// Poll runs a single polling cycle: fetch cards, diff against saved state,
// notify for new cards, then persist the updated state.
func (p *Poller) Poll(ctx context.Context) {
	log.Printf("[%s] Starting polling cycle...", time.Now().Format("15:04:05"))

	st, err := p.store.Load()
	if err != nil {
		log.Printf("Error loading state: %v", err)
		return
	}

	opts := scryfall.SearchCardsOptions{
		Unique: scryfall.UniqueModeArt,
		Order:  scryfall.OrderSet,
		Dir:    scryfall.DirDesc,
		Page:   1,
	}

	resp, err := p.client.SearchCards(ctx, p.query, opts)
	if err != nil {
		log.Printf("Error searching Scryfall: %v", err)
		return
	}

	if len(resp.Cards) == 0 {
		log.Printf("No cards found matching query: %s", p.query)
		return
	}

	hasMore := resp.HasMore
	log.Printf("Fetched %d card(s) from Scryfall (has_more=%v).", len(resp.Cards), hasMore)
	cards := resp.Cards
	while hasMore {
		opts.Page++
		nextResp, err := p.client.SearchCards(ctx, p.query, opts)
		if err != nil {
			log.Printf("Error fetching page %d: %v", opts.Page, err)
			break
		}
		cards = append(cards, nextResp.Cards...)
		hasMore = nextResp.HasMore
	}


	firstRun := len(st.SeenIDs) == 0

	// Collect every card ID from this fetch so we can persist it afterwards.
	currentIDs := make(map[string]struct{}, len(cards))
	for _, c := range cards {
		currentIDs[c.ID] = struct{}{}
	}

	if firstRun {
		// On the very first run we record the current set without notifying,
		// so we don't spam the channel with the entire back-catalogue.
		log.Printf("First run: recording %d cards without notifying.", len(cards))
	} else {
		// Walk in reverse (oldest first) for chronological notification order.
		newCount := 0
		for i := len(cards) - 1; i >= 0; i-- {
			c := cards[i]
			if _, seen := st.SeenIDs[c.ID]; seen {
				continue
			}

			card := toNotifyCard(&c)
			if err := p.notifier.Notify(ctx, card); err != nil {
				log.Printf("WARNING: Failed to notify for %s — card will be retried in the next iteration: %v", c.Name, err)
				delete(currentIDs, c.ID) // Don't mark as seen if notification failed
				continue
			}
			newCount++
			time.Sleep(p.rateLimitDelay)
		}
		log.Printf("Notified %d new card(s) this cycle.", newCount)
	}

	// Replace the seen set with the current fetch and save.
	st.SeenIDs = currentIDs
	st.LastCheck = time.Now()

	if err := p.store.Save(st); err != nil {
		log.Printf("Error saving state: %v", err)
	}

	log.Println("Polling cycle complete.")
}

// toNotifyCard maps a scryfall.Card to the generic notify.Card.
func toNotifyCard(c *scryfall.Card) notify.Card {
	colors := make([]string, len(c.Colors))
	for i, clr := range c.Colors {
		colors[i] = string(clr)
	}
	// Double-faced cards (DFCs) have an empty (or absent) top-level ImageURIs;
	// fall back to the front face image instead.
	var imageURL string
	if c.ImageURIs != nil {
		imageURL = c.ImageURIs.Normal
	}
	if imageURL == "" && len(c.CardFaces) > 0 {
		imageURL = c.CardFaces[0].ImageURIs.Normal
	}

	return notify.Card{
		ID:          c.ID,
		Name:        c.Name,
		ImageURL:    imageURL,
		ScryfallURI: c.ScryfallURI,
		Colors:      colors,
		Prices: notify.Prices{
			USD: c.Prices.USD,
			EUR: c.Prices.EUR,
		},
	}
}
