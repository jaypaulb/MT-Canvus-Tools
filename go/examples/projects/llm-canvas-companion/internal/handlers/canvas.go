package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/sashabaranov/go-openai"
)

// HandleCanvasPrecis processes an AI_Icon_Canvus_Precis icon and generates a
// summary of all visible content on the canvas.
func HandleCanvasPrecis(ctx context.Context, s *canvus.Session, cfg *config.Config, canvasID string, iconWidget canvus.Widget) {
	start := time.Now()

	opCtx, cancel := context.WithTimeout(ctx, cfg.ProcessingTimeout)
	defer cancel()

	// Mark the icon as processing by renaming it.
	_, _ = s.UpdateImage(ctx, canvasID, iconWidget.ID, map[string]any{
		"title": "!! AI Processing !! Canvas Precis",
	})

	// Fetch all widgets.
	widgets, err := listWidgetsWithRetry(opCtx, s, cfg, canvasID)
	if err != nil {
		slog.Error("HandleCanvasPrecis: list widgets failed", "err", err)
		return
	}

	// Create processing note near the icon.
	processingID, err := createProcessingNote(ctx, s, canvasID, iconWidget)
	if err != nil {
		slog.Error("HandleCanvasPrecis: createProcessingNote failed", "err", err)
		return
	}

	// Exclude the icon itself from the payload.
	var filtered []canvus.Widget
	for _, w := range widgets {
		if w.ID != iconWidget.ID {
			filtered = append(filtered, w)
		}
	}

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
		"Analysing canvas content...\nProcessing "+strconv.Itoa(len(filtered))+" widgets",
		1, 0)

	widgetsJSON, err := json.Marshal(filtered)
	if err != nil {
		slog.Error("HandleCanvasPrecis: marshal widgets failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}

	systemMsg := `You are an assistant analyzing a collaborative workspace.
Describe the content and relationships between items in a natural, narrative way.
Focus on the story the workspace is telling and how items relate to each other.
Avoid mentioning technical details like IDs or coordinates.
Format your response as JSON: {"type":"text","content":"..."} where content is Markdown with:
# Overview
Describe the main themes and content.
# Insights
Share observations about relationships.
# Recommendations
Provide actionable recommendations.`

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
		"Generating canvas analysis... This may take a moment.", 1, 0)

	clientCfg := openai.DefaultConfig(cfg.OpenAIAPIKey)
	if cfg.TextLLMURL != "" {
		clientCfg.BaseURL = cfg.TextLLMURL
	} else if cfg.BaseLLMURL != "" {
		clientCfg.BaseURL = cfg.BaseLLMURL
	}
	aiClient := openai.NewClientWithConfig(clientCfg)

	resp, err := aiClient.CreateChatCompletion(opCtx, openai.ChatCompletionRequest{
		Model: cfg.OpenAICanvasModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemMsg},
			{Role: openai.ChatMessageRoleUser, Content: string(widgetsJSON)},
		},
		MaxTokens:   int(cfg.CanvasPrecisTokens),
		Temperature: 0.3,
	})
	if err != nil {
		slog.Error("HandleCanvasPrecis: AI call failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}
	if len(resp.Choices) == 0 {
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}

	content, err := extractJSONContent(resp.Choices[0].Message.Content)
	if err != nil {
		slog.Error("HandleCanvasPrecis: extract JSON failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}
	content = strings.ReplaceAll(content, "\\n", "\n")

	if err := createResponseNote(ctx, s, canvasID, iconWidget.ID, iconWidget, content, "", false); err != nil {
		slog.Error("HandleCanvasPrecis: createResponseNote failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}

	deleteWidget(ctx, s, canvasID, "note", processingID)
	deleteWidget(ctx, s, canvasID, "image", iconWidget.ID)

	addDuration(time.Since(start))
	slog.Info("HandleCanvasPrecis: completed", "elapsed", time.Since(start))
}

// listWidgetsWithRetry fetches all widgets with up to cfg.MaxRetries attempts.
func listWidgetsWithRetry(ctx context.Context, s *canvus.Session, cfg *config.Config, canvasID string) ([]canvus.Widget, error) {
	var lastErr error
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		widgets, err := s.ListWidgets(ctx, canvasID, nil)
		if err == nil {
			return widgets, nil
		}
		lastErr = err
		slog.Warn("listWidgetsWithRetry: attempt failed",
			"attempt", attempt, "max", cfg.MaxRetries, "err", err)
		time.Sleep(cfg.RetryDelay)
	}
	return nil, fmt.Errorf("listWidgetsWithRetry: %w", lastErr)
}
