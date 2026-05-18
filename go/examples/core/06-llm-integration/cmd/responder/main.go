// Command responder is a diagnostic for the LLM watcher: it queries the
// local Ollama instance's /api/tags endpoint and prints the installed
// model names so you can verify Ollama is reachable before booting the
// watcher.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultOllamaURL = "http://localhost:11434"

type tagResponse struct {
	Models []struct {
		Name       string `json:"name"`
		ModifiedAt string `json:"modified_at"`
		Size       int64  `json:"size"`
	} `json:"models"`
}

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("responder failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	base := strings.TrimRight(envDefault("OLLAMA_URL", defaultOllamaURL), "/")
	url := base + "/api/tags"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, url, body)
	}

	var tags tagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	slog.Info("ollama reachable", "url", base, "model_count", len(tags.Models))
	if len(tags.Models) == 0 {
		fmt.Println("(no models installed — run `ollama pull llama3.2` first)")
		return nil
	}
	for _, m := range tags.Models {
		fmt.Printf("%-40s  %12d bytes  %s\n", m.Name, m.Size, m.ModifiedAt)
	}
	return nil
}

func envDefault(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

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
