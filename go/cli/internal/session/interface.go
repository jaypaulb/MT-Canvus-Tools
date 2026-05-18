package session

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// SessionProvider is an interface that wraps the canvus.Session methods we use.
// This allows tests to inject mock sessions.
// Note: Method signatures must match canvus.Session exactly so the real
// session value satisfies the interface without an adapter.
type SessionProvider interface {
	// Canvas operations
	CreateCanvas(ctx context.Context, req any) (*canvus.Canvas, error)
	ListCanvases(ctx context.Context, filter *canvus.Filter) ([]canvus.Canvas, error)
	GetCanvas(ctx context.Context, id string) (*canvus.Canvas, error)
	UpdateCanvas(ctx context.Context, id string, req any) (*canvus.Canvas, error)
	DeleteCanvas(ctx context.Context, id string) error
	MoveCanvas(ctx context.Context, id string, req canvus.MoveOrCopyCanvasRequest) (*canvus.Canvas, error)
	CopyCanvas(ctx context.Context, id string, req canvus.MoveOrCopyCanvasRequest) (*canvus.Canvas, error)

	// Widget operations
	CreateWidget(ctx context.Context, canvasID string, req any, contentType ...string) (*canvus.Widget, error)
	ListWidgets(ctx context.Context, canvasID string, filter *canvus.Filter, includeAnnotations ...bool) ([]canvus.Widget, error)
	GetWidget(ctx context.Context, canvasID, widgetID string) (*canvus.Widget, error)
	UpdateWidget(ctx context.Context, canvasID, widgetID string, req map[string]any) (*canvus.Widget, error)
	DeleteWidget(ctx context.Context, canvasID, widgetID, widgetType string) error

	// User operations
	CreateUser(ctx context.Context, req any) (*canvus.User, error)
	ListUsers(ctx context.Context) ([]canvus.User, error)
	GetUser(ctx context.Context, id int64) (*canvus.User, error)
	UpdateUser(ctx context.Context, id int64, req any) (*canvus.User, error)
	DeleteUser(ctx context.Context, id int64) error
	CreateAccessToken(ctx context.Context, userID int64, req any) (*canvus.AccessToken, error)
	ListAccessTokens(ctx context.Context, userID int64) ([]canvus.AccessToken, error)
	DeleteAccessToken(ctx context.Context, userID int64, tokenID string) error

	// Group operations
	CreateGroup(ctx context.Context, req any) (*canvus.Group, error)
	ListGroups(ctx context.Context) ([]canvus.Group, error)
	GetGroup(ctx context.Context, id int) (*canvus.Group, error)
	DeleteGroup(ctx context.Context, id int) error
	AddUserToGroup(ctx context.Context, groupID int, userID int64) error
	RemoveUserFromGroup(ctx context.Context, groupID int, userID int64) error

	// System operations
	GetLicenseInfo(ctx context.Context) (*canvus.LicenseInfo, error)

	// Authentication operations
	Login(ctx context.Context, username, password string) error
}

// sessionProviderKey is the context key for SessionProvider (used in tests)
type sessionProviderKey int

const (
	providerKey sessionProviderKey = iota + 1000 // Offset to avoid collision with sessionKey
)

// WithSessionProvider returns a new context with the session provider attached.
// This is used in tests to inject mock sessions.
func WithSessionProvider(ctx context.Context, provider SessionProvider) context.Context {
	return context.WithValue(ctx, providerKey, provider)
}

// GetSessionProvider retrieves the session provider from the context.
// If not found, it tries to get the real session and wraps it.
func GetSessionProvider(ctx context.Context) (SessionProvider, error) {
	// First try to get test provider
	if provider, ok := ctx.Value(providerKey).(SessionProvider); ok {
		return provider, nil
	}

	// Fall back to real session
	return GetSession(ctx)
}
