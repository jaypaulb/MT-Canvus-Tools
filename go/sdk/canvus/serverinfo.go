package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ServerInfo represents information about the Canvus Server instance.
type ServerInfo struct {
	API      []string `json:"api"`
	Go       string   `json:"go"`
	ServerID string   `json:"server_id"`
	Version  string   `json:"version"`
}

// GetServerInfo retrieves information about the Canvus Server instance.
func (s *Session) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	var info ServerInfo
	if err := s.doRequest(ctx, http.MethodGet, "server-info", nil, &info, nil, false); err != nil {
		return nil, fmt.Errorf("GetServerInfo: %w", err)
	}
	return &info, nil
}
