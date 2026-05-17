package canvus

import (
	"context"
	"fmt"
	"net/http"
)

// ServerConfig represents the server configuration as a nested struct.
//
// Phase 3 Go work item #10 flags that the spec describes a flat
// `[]ConfigElement{Key, Value, Type}` shape; the legacy SDK decodes into
// this nested struct. Until Phase 4 verifies the live server, we retain the
// nested struct AND expose a raw `ServerConfigRaw` map for flexibility.
type ServerConfig struct {
	Access         string                `json:"access,omitempty"`
	Authentication *AuthenticationConfig `json:"authentication,omitempty"`
	Email          *EmailConfig          `json:"email,omitempty"`
	ExternalURL    string                `json:"external_url,omitempty"`
	ServerName     string                `json:"server_name,omitempty"`
}

// AuthenticationConfig represents authentication settings for the server.
type AuthenticationConfig struct {
	DomainAllowList      []string        `json:"domain_allow_list,omitempty"`
	Password             *PasswordConfig `json:"password,omitempty"`
	QRCode               *QRCodeConfig   `json:"qr_code,omitempty"`
	RequireAdminApproval bool            `json:"require_admin_approval,omitempty"`
	SAML                 *SAMLConfig     `json:"saml,omitempty"`
}

// PasswordConfig represents password authentication settings.
type PasswordConfig struct {
	Enabled       bool `json:"enabled,omitempty"`
	SignUpEnabled bool `json:"sign_up_enabled,omitempty"`
}

// QRCodeConfig represents QR code authentication settings.
type QRCodeConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// SAMLConfig represents SAML authentication settings.
type SAMLConfig struct {
	ACSURL             string `json:"acs_url,omitempty"`
	Enabled            bool   `json:"enabled,omitempty"`
	IDPCertFingerPrint string `json:"idp_cert_finger_print,omitempty"`
	IDPEntityID        string `json:"idp_entity_id,omitempty"`
	IDPTargetURL       string `json:"idp_target_url,omitempty"`
	NameIDFormat       string `json:"name_id_format,omitempty"`
	SignUpEnabled      bool   `json:"sign_up_enabled,omitempty"`
	SPEntityID         string `json:"sp_entity_id,omitempty"`
}

// EmailConfig represents email server settings.
type EmailConfig struct {
	MailReplyToAddress              string `json:"mail_reply_to_address,omitempty"`
	MailReplyToName                 string `json:"mail_reply_to_name,omitempty"`
	MailSenderAddress               string `json:"mail_sender_address,omitempty"`
	MailSenderName                  string `json:"mail_sender_name,omitempty"`
	SMTPAllowSelfSignedCertificates bool   `json:"smtp_allow_self_signed_certificates,omitempty"`
	SMTPHost                        string `json:"smtp_host,omitempty"`
	SMTPPassword                    string `json:"smtp_password,omitempty"`
	SMTPPort                        int    `json:"smtp_port,omitempty"`
	SMTPSecurity                    string `json:"smtp_security,omitempty"`
	SMTPUsername                    string `json:"smtp_username,omitempty"`
}

// ConfigElement is the flat element type spec'd by the public docs. Provided
// for callers using GetServerConfigRaw.
type ConfigElement struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
	Type  string `json:"type,omitempty"`
}

// GetServerConfig retrieves the server configuration as a nested struct.
func (s *Session) GetServerConfig(ctx context.Context) (*ServerConfig, error) {
	var config ServerConfig
	if err := s.doRequest(ctx, http.MethodGet, "server-config", nil, &config, nil, false); err != nil {
		return nil, fmt.Errorf("GetServerConfig: %w", err)
	}
	return &config, nil
}

// GetServerConfigRaw retrieves the server configuration as the spec'd flat
// element array. Useful while the response-shape ambiguity (Phase 3 item #10)
// is still being investigated.
func (s *Session) GetServerConfigRaw(ctx context.Context) ([]ConfigElement, error) {
	var elems []ConfigElement
	if err := s.doRequest(ctx, http.MethodGet, "server-config", nil, &elems, nil, false); err != nil {
		return nil, fmt.Errorf("GetServerConfigRaw: %w", err)
	}
	return elems, nil
}

// UpdateServerConfig updates the server configuration.
func (s *Session) UpdateServerConfig(ctx context.Context, req any) (*ServerConfig, error) {
	var config ServerConfig
	if err := s.doRequest(ctx, http.MethodPatch, "server-config", req, &config, nil, false); err != nil {
		return nil, fmt.Errorf("UpdateServerConfig: %w", err)
	}
	return &config, nil
}

// SendTestEmail sends a test email. Per Phase 3 work item #9, the spec
// example body includes a `recipient-email` field. We default to the current
// session user's email when recipient is empty; otherwise we honour the arg.
func (s *Session) SendTestEmail(ctx context.Context, recipient string) error {
	body := map[string]string{}
	if recipient != "" {
		body["recipient-email"] = recipient
		body["recipient_email"] = recipient // defensive: server may have either casing
	}
	return s.doRequest(ctx, http.MethodPost, "server-config/send-test-email", body, nil, nil, false)
}

// ReloadCerts reloads the server's TLS certificates.
func (s *Session) ReloadCerts(ctx context.Context) error {
	return s.doRequest(ctx, http.MethodPost, "server-config/reload-certs", nil, nil, nil, false)
}
