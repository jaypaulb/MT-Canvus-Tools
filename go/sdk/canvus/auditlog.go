package canvus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuditEvent represents an audit log entry in the Canvus system.
type AuditEvent struct {
	ID        json.Number `json:"id"`
	Timestamp string      `json:"timestamp,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Action    string      `json:"action,omitempty"`
	Resource  string      `json:"resource,omitempty"`
	Details   string      `json:"details,omitempty"`
}

// AuditLogResponse is the envelope returned by GET /audit-log per the spec.
// Per Phase 3 work item #6, the SDK now decodes the envelope rather than the
// flat array the legacy SDK assumed.
type AuditLogResponse struct {
	Events     []AuditEvent `json:"events"`
	TotalCount int          `json:"total-count"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per-page"`
}

// auditQueryFromOpts builds the request query map from AuditLogOptions.
// Spec keys are hyphenated (`per-page`, `start-time`, `end-time`, `user-id`).
// The legacy SDK only supported `per_page`.
func auditQueryFromOpts(opts *AuditLogOptions) map[string]string {
	q := map[string]string{}
	if opts == nil {
		return q
	}
	if opts.Page > 0 {
		q["page"] = fmt.Sprintf("%d", opts.Page)
	}
	if opts.PerPage > 0 {
		q["per-page"] = fmt.Sprintf("%d", opts.PerPage)
		// Defensive: some older builds accept per_page only.
		q["per_page"] = fmt.Sprintf("%d", opts.PerPage)
	}
	if opts.Filter != "" {
		q["filter"] = opts.Filter
	}
	if opts.StartTime != "" {
		q["start-time"] = opts.StartTime
	}
	if opts.EndTime != "" {
		q["end-time"] = opts.EndTime
	}
	if opts.UserID != "" {
		q["user-id"] = opts.UserID
	}
	if opts.Action != "" {
		q["action"] = opts.Action
	}
	return q
}

// ListAuditEvents retrieves audit log events with the spec-defined envelope.
//
// Breaking change vs legacy SDK (returned []AuditEvent only): this now
// returns the full envelope so callers can paginate. See AuditLogResponse.
func (s *Session) ListAuditEvents(ctx context.Context, opts *AuditLogOptions) (*AuditLogResponse, error) {
	// Try the envelope decode first. If the server replies with a flat array
	// (older build), fall back to wrapping it.
	var raw json.RawMessage
	if err := s.doRequest(ctx, http.MethodGet, "audit-log", nil, &raw, auditQueryFromOpts(opts), false); err != nil {
		return nil, fmt.Errorf("ListAuditEvents: %w", err)
	}
	var env AuditLogResponse
	if err := json.Unmarshal(raw, &env); err == nil && (env.Events != nil || env.TotalCount > 0 || env.Page > 0) {
		return &env, nil
	}
	var arr []AuditEvent
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, fmt.Errorf("ListAuditEvents: decode failed: %w", err)
	}
	return &AuditLogResponse{Events: arr, TotalCount: len(arr), Page: 1, PerPage: len(arr)}, nil
}

// ExportAuditLog exports the audit log as a CSV file with the same filter set
// as ListAuditEvents.
func (s *Session) ExportAuditLog(ctx context.Context, opts *AuditLogOptions) ([]byte, error) {
	var data []byte
	if err := s.doRequest(ctx, http.MethodGet, "audit-log/export-csv", nil, &data, auditQueryFromOpts(opts), true); err != nil {
		return nil, fmt.Errorf("ExportAuditLog: %w", err)
	}
	return data, nil
}
