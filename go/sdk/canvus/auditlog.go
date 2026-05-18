package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// AuditEvent represents an audit log entry in the Canvus system.
type AuditEvent struct {
	ID         int     `json:"id"`
	Action     string  `json:"action,omitempty"`
	AuthorID   *int    `json:"author_id,omitempty"`
	CreatedAt  string  `json:"created_at,omitempty"`
	Details    string  `json:"details,omitempty"`
	IPAddress  string  `json:"ip_address,omitempty"`
	TargetID   *string `json:"target_id,omitempty"`
	TargetType string  `json:"target_type,omitempty"`
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

// ListAuditEvents retrieves audit log events as a flat array.
func (s *Session) ListAuditEvents(ctx context.Context, opts *AuditLogOptions) ([]AuditEvent, error) {
	var events []AuditEvent
	if err := s.doRequest(ctx, http.MethodGet, "audit-log", nil, &events, auditQueryFromOpts(opts), false); err != nil {
		return nil, fmt.Errorf("ListAuditEvents: %w", err)
	}
	return events, nil
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
