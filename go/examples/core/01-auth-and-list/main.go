// Command 01-auth-and-list authenticates against a Canvus server using an
// API key, lists every accessible canvas, and prints a tab-aligned summary
// table to stdout.
//
// It is the canonical Canvus "hello world" — minimal happy-path code that
// proves the env vars are set, the server is reachable, and the API key is
// accepted.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("auth-and-list failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL, err := mustEnv("CANVUS_BASE_URL")
	if err != nil {
		return err
	}
	apiKey, err := mustEnv("CANVUS_API_KEY")
	if err != nil {
		return err
	}

	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))

	ctx := context.Background()
	canvases, err := s.ListCanvases(ctx, nil)
	if err != nil {
		return fmt.Errorf("list canvases: %w", err)
	}

	slog.Info("listed canvases", "count", len(canvases))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()

	if _, err := fmt.Fprintln(w, "ID\tNAME\tMODE\tACCESS\tSTATE"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, c := range canvases {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.ID, c.Name, c.Mode, c.Access, c.State); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	return nil
}

// mustEnv returns the value of name or an error if it is empty/unset.
func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

// setupLogging configures slog.Default with a text or JSON handler depending
// on LOG_FORMAT.
func setupLogging() {
	var h slog.Handler
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}
