package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestParseDuration covers the unexported helper directly.
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback time.Duration
		want     time.Duration
	}{
		{"empty string returns fallback", "", time.Hour, time.Hour},
		{"valid 30m", "30m", time.Hour, 30 * time.Minute},
		{"valid 5s", "5s", time.Minute, 5 * time.Second},
		{"zero duration", "0s", time.Hour, 0},
		{"invalid string returns fallback", "not-a-duration", time.Hour, time.Hour},
		{"bare number (missing unit) returns fallback", "1234", 5 * time.Minute, 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDuration(tt.input, tt.fallback)
			if got != tt.want {
				t.Errorf("parseDuration(%q, %v) = %v, want %v", tt.input, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestLoadFromEnv_Defaults(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "https://discord.com/api/webhooks/test")
	t.Setenv("POLL_INTERVAL", "")
	t.Setenv("DATA_DIR", "")
	t.Setenv("SEARCH_QUERY", "")

	cfg := LoadFromEnv()

	if cfg.PollInterval != time.Hour {
		t.Errorf("PollInterval default = %v, want 1h", cfg.PollInterval)
	}
	if cfg.DataDir != "/root/data" {
		t.Errorf("DataDir default = %q, want /root/data", cfg.DataDir)
	}
	wantQuery := fmt.Sprintf("year=%d", time.Now().UTC().Year())
	if cfg.SearchQuery != wantQuery {
		t.Errorf("SearchQuery default = %q, want %q", cfg.SearchQuery, wantQuery)
	}
}

func TestLoadFromEnv_ExplicitValues(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "https://discord.com/api/webhooks/custom")
	t.Setenv("POLL_INTERVAL", "15m")
	t.Setenv("DATA_DIR", "/tmp/testdata")
	t.Setenv("SEARCH_QUERY", "is:preview")

	cfg := LoadFromEnv()

	if cfg.WebhookURL != "https://discord.com/api/webhooks/custom" {
		t.Errorf("WebhookURL = %q, want custom webhook URL", cfg.WebhookURL)
	}
	if cfg.PollInterval != 15*time.Minute {
		t.Errorf("PollInterval = %v, want 15m", cfg.PollInterval)
	}
	if cfg.DataDir != "/tmp/testdata" {
		t.Errorf("DataDir = %q, want /tmp/testdata", cfg.DataDir)
	}
	if cfg.SearchQuery != "is:preview" {
		t.Errorf("SearchQuery = %q, want is:preview", cfg.SearchQuery)
	}
}

func TestLoadFromEnv_InvalidPollIntervalFallsBackToDefault(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "https://discord.com/api/webhooks/test")
	t.Setenv("POLL_INTERVAL", "not-a-valid-duration")

	cfg := LoadFromEnv()

	if cfg.PollInterval != time.Hour {
		t.Errorf("PollInterval with invalid input = %v, want 1h fallback", cfg.PollInterval)
	}
}

// TestLoadFromEnv_MissingWebhookURL_IsFatal verifies that the app terminates
// when WEBHOOK_URL is absent.  It uses the subprocess pattern to safely test
// log.Fatal without aborting the test suite.
func TestLoadFromEnv_MissingWebhookURL_IsFatal(t *testing.T) {
	const sentinelEnv = "TEST_FATAL_SUBPROCESS"
	if os.Getenv(sentinelEnv) == "1" {
		// Running inside the subprocess: trigger the fatal path.
		os.Unsetenv("WEBHOOK_URL")
		LoadFromEnv()
		return
	}

	// Build the subprocess environment: everything except WEBHOOK_URL + sentinel.
	env := []string{sentinelEnv + "=1"}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "WEBHOOK_URL=") && !strings.HasPrefix(kv, sentinelEnv+"=") {
			env = append(env, kv)
		}
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoadFromEnv_MissingWebhookURL_IsFatal")
	cmd.Env = env
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.Success() {
		t.Fatal("expected non-zero exit when WEBHOOK_URL is missing, got success")
	}
}