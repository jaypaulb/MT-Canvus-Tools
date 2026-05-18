package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/ledongthuc/pdf"
	"github.com/sashabaranov/go-openai"
)

// HandlePDFPrecis processes an AI_Icon_PDF_Precis icon placed on a PDF widget.
// The icon's parent_id identifies the PDF to summarise.
// iconWidget is the AI icon; parentID is the PDF widget's ID.
func HandlePDFPrecis(ctx context.Context, s *canvus.Session, cfg *config.Config, canvasID, parentID string, iconWidget canvus.Widget) {
	start := time.Now()

	// Fetch the parent PDF widget to get its geometry.
	pdfWidget, err := s.GetWidget(ctx, canvasID, parentID)
	if err != nil {
		slog.Error("HandlePDFPrecis: GetWidget(parent PDF) failed", "parent_id", parentID, "err", err)
		return
	}
	if !strings.EqualFold(pdfWidget.WidgetType, "pdf") {
		slog.Error("HandlePDFPrecis: parent is not a PDF", "widget_type", pdfWidget.WidgetType)
		return
	}

	// Create processing note.
	processingID, err := createProcessingNote(ctx, s, canvasID, *pdfWidget)
	if err != nil {
		slog.Error("HandlePDFPrecis: createProcessingNote failed", "err", err)
		return
	}

	opCtx, cancel := context.WithTimeout(ctx, cfg.ProcessingTimeout)
	defer cancel()

	// Download PDF bytes.
	_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Downloading PDF...", 1, 0)
	pdfData, err := s.DownloadPDF(opCtx, canvasID, parentID)
	if err != nil {
		slog.Error("HandlePDFPrecis: DownloadPDF failed", "pdf_id", parentID, "err", err)
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Failed to download PDF.", 1, 0)
		return
	}

	// Write to temp file for the ledongthuc/pdf parser.
	tmpPath := filepath.Join(cfg.DownloadsDir, fmt.Sprintf("temp_pdf_%s.pdf", parentID))
	if err := os.MkdirAll(cfg.DownloadsDir, 0o755); err != nil {
		slog.Error("HandlePDFPrecis: mkdir failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}
	if err := os.WriteFile(tmpPath, pdfData, 0o644); err != nil {
		slog.Error("HandlePDFPrecis: write temp PDF failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}
	defer removeFile(tmpPath)

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Extracting text from PDF...", 1, 0)
	pdfText, err := extractPDFText(tmpPath)
	if err != nil {
		slog.Error("HandlePDFPrecis: text extraction failed", "err", err)
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Failed to extract text from PDF.", 1, 0)
		return
	}

	chunks := splitIntoChunks(pdfText, int(cfg.PDFChunkSizeTokens))
	totalChunks := len(chunks)

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID,
		fmt.Sprintf("Preparing %d PDF sections for analysis...", totalChunks), 1, 0)

	// Build multi-chunk message history.
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("You will receive %d chunks of a document. Do not respond until you receive the final chunk. After the last chunk, I will prompt you for your analysis of the entire document.", totalChunks),
		},
	}
	for i, chunk := range chunks {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("#--- chunk %d of %d ---#\n%s\n#--- end of chunk %d ---#", i+1, totalChunks, chunk, i+1),
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		Content: `You have now received all chunks. Please analyze the entire document and provide a summary in the following JSON format:
{"type": "text", "content": "..."}
The content field must be a Markdown-formatted summary with:
# Overview
# Key Points
# Details
# Conclusions
Respond ONLY with valid JSON as shown above, and ensure the content is Markdown.`,
	})

	_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Generating final PDF analysis...", 1, 0)

	clientCfg := openai.DefaultConfig(cfg.OpenAIAPIKey)
	if cfg.TextLLMURL != "" {
		clientCfg.BaseURL = cfg.TextLLMURL
	} else if cfg.BaseLLMURL != "" {
		clientCfg.BaseURL = cfg.BaseLLMURL
	}
	aiClient := openai.NewClientWithConfig(clientCfg)

	resp, err := aiClient.CreateChatCompletion(opCtx, openai.ChatCompletionRequest{
		Model:       cfg.OpenAIPDFModel,
		Messages:    messages,
		MaxTokens:   int(cfg.PDFPrecisTokens),
		Temperature: 0.3,
	})
	if err != nil {
		slog.Error("HandlePDFPrecis: AI call failed", "err", err)
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "Failed to generate PDF analysis.", 1, 0)
		return
	}
	if len(resp.Choices) == 0 {
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "No response from AI.", 1, 0)
		return
	}

	content, err := extractJSONContent(resp.Choices[0].Message.Content)
	if err != nil {
		slog.Error("HandlePDFPrecis: extract JSON failed", "err", err)
		_ = updateNoteWithRetry(ctx, s, canvasID, processingID, "AI response was not valid JSON.", 1, 0)
		return
	}

	content = strings.ReplaceAll(content, "\\n", "\n")
	if err := createResponseNote(ctx, s, canvasID, parentID, *pdfWidget, content, "", false); err != nil {
		slog.Error("HandlePDFPrecis: createResponseNote failed", "err", err)
		deleteWidget(ctx, s, canvasID, "note", processingID)
		return
	}

	deleteWidget(ctx, s, canvasID, "note", processingID)
	deleteWidget(ctx, s, canvasID, "image", iconWidget.ID)

	atomic.AddInt64(&Metrics.ProcessedPDFs, 1)
	addDuration(time.Since(start))
	slog.Info("HandlePDFPrecis: completed", "pdf_id", parentID, "elapsed", time.Since(start))
}

// extractPDFText extracts plain text from a PDF file path.
func extractPDFText(pdfPath string) (string, error) {
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("extractPDFText: open: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			slog.Warn("extractPDFText: page skipped", "page", i, "err", err)
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n\n")
	}

	result := sb.String()
	if result == "" {
		return "", fmt.Errorf("extractPDFText: no text content found")
	}
	return result, nil
}

// splitIntoChunks splits text into chunks of at most maxChunkSize characters,
// splitting on paragraph boundaries.
func splitIntoChunks(text string, maxChunkSize int) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var chunks []string
	paragraphs := strings.Split(text, "\n\n")

	var sb strings.Builder
	currentSize := 0

	for _, para := range paragraphs {
		paraSize := len(para)
		if currentSize+paraSize > maxChunkSize && sb.Len() > 0 {
			chunks = append(chunks, sb.String())
			sb.Reset()
			currentSize = 0
		}
		sb.WriteString(para)
		sb.WriteString("\n\n")
		currentSize += paraSize + 2
	}
	if sb.Len() > 0 {
		chunks = append(chunks, sb.String())
	}
	return chunks
}

// extractJSONContent extracts the "content" field from a JSON response of the
// form {"type":"text","content":"..."}.
func extractJSONContent(raw string) (string, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || start > end {
		return "", fmt.Errorf("extractJSONContent: no JSON object in response")
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(raw[start:end+1]), &m); err != nil {
		return "", fmt.Errorf("extractJSONContent: parse: %w", err)
	}
	content, ok := m["content"].(string)
	if !ok {
		return "", fmt.Errorf("extractJSONContent: missing 'content' field")
	}
	return content, nil
}
