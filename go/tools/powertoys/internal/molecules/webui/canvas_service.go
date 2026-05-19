package webui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
	webuiatoms "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/webui"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/services"
)

// CanvasService tracks the active canvas ID by subscribing to workspace events.
type CanvasService struct {
	session            *canvus.Session
	clientResolver     *webuiatoms.ClientResolver
	canvasTracker      *webuiatoms.CanvasTracker
	ctx                context.Context
	cancel             context.CancelFunc
	clientID           string
	clientName         string
	installationName   string
	overrideClientName string
	mu                 sync.RWMutex
	wg                 sync.WaitGroup
	hasReceivedEvents  bool
	lastEventTime      time.Time
	subscriptionStart  time.Time
}

// NewCanvasService creates a canvas service.
func NewCanvasService(fileService *services.FileService, session *canvus.Session) *CanvasService {
	ctx, cancel := context.WithCancel(context.Background())
	return &CanvasService{
		session:        session,
		clientResolver: webuiatoms.NewClientResolver(fileService),
		canvasTracker:  webuiatoms.NewCanvasTracker(),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start resolves the client ID and begins workspace subscription.
func (cs *CanvasService) Start() error {
	installationName, err := cs.clientResolver.GetInstallationName()
	if err != nil {
		return fmt.Errorf("CanvasService.Start: %w", err)
	}
	cs.mu.Lock()
	cs.installationName = installationName
	cs.mu.Unlock()

	cs.mu.RLock()
	ctx := cs.ctx
	cs.mu.RUnlock()

	clientID, err := cs.clientResolver.ResolveClientID(ctx, cs.session, installationName)
	if err != nil {
		// Not fatal — canvas ID can be set via manual override.
		logger.Logf("CanvasService: could not resolve client ID for %q: %v — canvas ID must be set manually", installationName, err)
		return nil
	}

	if clientID == "" {
		return nil
	}

	cs.mu.Lock()
	cs.clientID = clientID
	cs.subscriptionStart = time.Now()
	cs.mu.Unlock()

	// Fetch client name asynchronously.
	go cs.refreshClientName()

	cs.startSubscription(clientID)
	return nil
}

// startSubscription subscribes to workspace events for the given clientID.
func (cs *CanvasService) startSubscription(clientID string) {
	cs.mu.RLock()
	ctx := cs.ctx
	cs.mu.RUnlock()

	subscriber := webuiatoms.NewWorkspaceSubscriber(cs.session, clientID)
	eventChan, errChan := subscriber.Subscribe(ctx)

	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		for range errChan {
			// errors are surfaced by the subscriber; context cancellation stops the subscription
		}
	}()

	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-eventChan:
				if !ok {
					return
				}
				cs.canvasTracker.UpdateCanvas(ev.CanvasID, ev.CanvasName)
				cs.mu.Lock()
				cs.hasReceivedEvents = true
				cs.lastEventTime = ev.Timestamp
				cs.mu.Unlock()
			}
		}
	}()
}

// refreshClientName fetches the client name from the server by matching the stored clientID.
func (cs *CanvasService) refreshClientName() {
	cs.mu.RLock()
	clientID := cs.clientID
	ctx := cs.ctx
	cs.mu.RUnlock()

	if clientID == "" {
		return
	}

	clients, err := cs.session.ListClients(ctx)
	if err != nil {
		return
	}

	for _, c := range clients {
		if c.ID == clientID {
			cs.mu.Lock()
			cs.clientName = c.InstallationName
			cs.mu.Unlock()
			return
		}
	}
}

// stopAndWait cancels the workspace subscription and waits for all goroutines to exit.
func (cs *CanvasService) stopAndWait() {
	cs.mu.Lock()
	cancel := cs.cancel
	cs.mu.Unlock()
	cancel()
	cs.wg.Wait()
}

// Stop cancels the workspace subscription.
func (cs *CanvasService) Stop() {
	cs.mu.Lock()
	cancel := cs.cancel
	cs.mu.Unlock()
	cancel()
}

// Restart restarts the canvas service subscription, re-resolving or using override client name.
func (cs *CanvasService) Restart() error {
	cs.stopAndWait()

	cs.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	cs.ctx = ctx
	cs.cancel = cancel
	cs.hasReceivedEvents = false
	cs.lastEventTime = time.Time{}
	override := cs.overrideClientName
	installationName := cs.installationName
	cs.mu.Unlock()

	name := override
	if name == "" {
		name = installationName
	}
	if name == "" {
		return fmt.Errorf("Restart: no client name available")
	}
	return cs.restartWithClientName(name)
}

// OverrideClient manually sets a client name to monitor.
func (cs *CanvasService) OverrideClient(clientName string) error {
	cs.mu.Lock()
	if clientName == "" {
		cs.overrideClientName = ""
	} else {
		cs.overrideClientName = clientName
	}
	cs.mu.Unlock()
	return cs.restartWithClientName(clientName)
}

// restartWithClientName restarts the workspace subscription using a specific client name.
func (cs *CanvasService) restartWithClientName(clientName string) error {
	cs.stopAndWait()

	cs.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	cs.ctx = ctx
	cs.cancel = cancel
	cs.mu.Unlock()

	clients, err := cs.session.ListClients(ctx)
	if err != nil {
		return fmt.Errorf("restartWithClientName: list clients: %w", err)
	}

	clientNameLower := strings.ToLower(clientName)
	var clientID string
	var foundName string
	for _, c := range clients {
		if strings.ToLower(c.InstallationName) == clientNameLower || strings.ToLower(c.Name) == clientNameLower {
			clientID = c.ID
			foundName = c.InstallationName
			break
		}
	}

	if clientID == "" {
		available := make([]string, 0, len(clients))
		for _, c := range clients {
			available = append(available, c.InstallationName)
		}
		return fmt.Errorf("restartWithClientName: no client with name %q; available: %v", clientName, available)
	}

	cs.mu.Lock()
	cs.clientID = clientID
	cs.clientName = foundName
	cs.hasReceivedEvents = false
	cs.lastEventTime = time.Time{}
	cs.subscriptionStart = time.Now()
	cs.mu.Unlock()

	cs.startSubscription(clientID)
	return nil
}

// GetCanvasID returns the currently active canvas ID.
func (cs *CanvasService) GetCanvasID() string { return cs.canvasTracker.GetCanvasID() }

// GetCanvasName returns the currently active canvas name.
func (cs *CanvasService) GetCanvasName() string { return cs.canvasTracker.GetCanvasName() }

// SetClientID manually overrides the client ID used for workspace subscription.
func (cs *CanvasService) SetClientID(id string) {
	cs.mu.Lock()
	cs.clientID = id
	cs.mu.Unlock()
}

// GetClientID returns the resolved or overridden client ID.
func (cs *CanvasService) GetClientID() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.clientID
}

// GetClientName returns the client's installation name from the server.
func (cs *CanvasService) GetClientName() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.clientName
}

// GetInstallationName returns the local installation name read from mt-canvus.ini.
func (cs *CanvasService) GetInstallationName() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.installationName
}

// IsConnected returns whether the service has a client ID and is subscribed.
func (cs *CanvasService) IsConnected() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	if cs.clientID == "" {
		return false
	}
	if cs.hasReceivedEvents {
		return true
	}
	if !cs.subscriptionStart.IsZero() {
		return time.Since(cs.subscriptionStart) < 30*time.Second
	}
	return true
}
