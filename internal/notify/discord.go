package notify

import (
	"context"
	"fmt"

	"github.com/gtuk/discordwebhook"
)

// DiscordNotifier sends card notifications via a Discord webhook.
type DiscordNotifier struct {
	webhookURL string
}

// NewDiscordNotifier creates a new DiscordNotifier.
func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{webhookURL: webhookURL}
}

// Notify sends a card embed to the configured Discord webhook.
func (d *DiscordNotifier) Notify(_ context.Context, card Card) error {
	title := card.Name
	description := ""
	switch {
		case card.Prices.USD != "":
			description = fmt.Sprintf("**Price (USD):** %s", card.Prices.USD)
		case card.Prices.EUR != "":
			description = fmt.Sprintf("**Price (EUR):** %s", card.Prices.EUR)
	}
	colorStr := fmt.Sprintf("%d", colorForCard(card))
	urlStr := card.ScryfallURI

	embed := discordwebhook.Embed{
		Title:       &title,
		Description: &description,
		Color:       &colorStr,
		Image:       &discordwebhook.Image{Url: &card.ImageURL},
		Url:         &urlStr,
	}

	content := fmt.Sprintf("New Card: **%s**", card.Name)
	embeds := []discordwebhook.Embed{embed}
	msg := discordwebhook.Message{
		Content: &content,
		Embeds:  &embeds,
	}

	if err := discordwebhook.SendMessage(d.webhookURL, msg); err != nil {
		return fmt.Errorf("discord webhook: %w", err)
	}
	return nil
}

// colorForCard returns a Discord embed colour based on the card's colour identity.
func colorForCard(card Card) int {
	if len(card.Colors) > 1 {
		return 0xFFD700 // Gold for multi-colour
	}
	colorMap := map[string]int{
		"W": 0xF0E68C, // White
		"U": 0x4169E1, // Blue
		"B": 0x1C1C1C, // Black
		"R": 0xFF4500, // Red
		"G": 0x228B22, // Green
	}
	if len(card.Colors) == 1 {
		if c, ok := colorMap[card.Colors[0]]; ok {
			return c
		}
	}
	return 0x808080 // Colourless
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func boolPtr(b bool) *bool { return &b }

