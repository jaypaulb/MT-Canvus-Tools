package webui

import (
	"context"
	"fmt"
	"sync"
	"time"

	canvus "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
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
	overrideClientName string
	mu                 sync.RWMutex
	hasReceivedEvents  bool
	lastEventTime      time.Time
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

	clientID, err := cs.clientResolver.ResolveClientID(cs.ctx, cs.session, installationName)
	if err != nil {
		// Not fatal — canvas ID can be set via manual override.
		return nil
	}

	cs.clientID = clientID
	subscriber := webuiatoms.NewWorkspaceSubscriber(cs.session, clientID)
	eventChan, _ := subscriber.Subscribe(cs.ctx)

	go func() {
		for ev := range eventChan {
			cs.canvasTracker.UpdateCanvas(ev.CanvasID, ev.CanvasName)
			cs.mu.Lock()
			cs.hasReceivedEvents = true
			cs.lastEventTime = ev.Timestamp
			cs.mu.Unlock()
		}
	}()

	return nil
}

// Stop cancels the workspace subscription.
func (cs *CanvasService) Stop() { cs.cancel() }

// GetCanvasID returns the currently active canvas ID.
func (cs *CanvasService) GetCanvasID() string { return cs.canvasTracker.GetCanvasID() }

// SetClientID manually overrides the client ID used for workspace subscription.
func (cs *CanvasService) SetClientID(id string) { cs.clientID = id }

// GetClientID returns the resolved or overridden client ID.
func (cs *CanvasService) GetClientID() string { return cs.clientID }
