package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BlueMonday/go-scryfall"

	"Scrybot/internal/config"
	"Scrybot/internal/notify"
	"Scrybot/internal/poller"
	"Scrybot/internal/server"
	"Scrybot/internal/state"
)

//go:embed ui/dist
var uiFiles embed.FS

func main() {
	cfg := config.LoadFromEnv()

	hub := server.NewHub()

	// Intercept log output so every line is also broadcast over WebSocket.
	log.SetOutput(server.NewLogWriter(hub, os.Stderr))
	log.SetFlags(0) // timestamps are added by LogWriter's ts field

	log.Printf("Starting Scryfall Alert Bot")
	log.Printf("  Poll Interval: %v", cfg.PollInterval)
	log.Printf("  Search Query:  %s", cfg.SearchQuery)
	log.Printf("  Data Directory: %s", cfg.DataDir)

	// --- Dependencies ---

	client, err := scryfall.NewClient(
		scryfall.WithUserAgent("Scrybot/1.0 +https://github.com/Ncorrex/scrybot"),
	)
	if err != nil {
		log.Fatalf("Failed to create Scryfall client: %v", err)
	}

	notifier := notify.NewDiscordNotifier(cfg.WebhookURL)

	store, err := state.NewJSONFileStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialise state store: %v", err)
	}

	p := poller.NewPoller(client, notifier, store, cfg.SearchQuery)

	// --- HTTP / WebSocket server ---

	uiPort := os.Getenv("UI_PORT")
	if uiPort == "" {
		uiPort = "8080"
	}

	sub, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Printf("Warning: could not sub-FS for UI: %v", err)
	}

	srv := server.New(hub, cfg, store, sub, uiPort)
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// --- Run loop ---

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	// Initial check
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p.Poll(ctx)
	}()

	for {
		select {
		case <-ticker.C:
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				p.Poll(ctx)
			}()
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Shutting down gracefully...", sig)
			return
		}
	}
}
