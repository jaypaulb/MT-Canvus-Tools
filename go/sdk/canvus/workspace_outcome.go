package canvus

import "fmt"

// OpenCanvasError preserves the outcome of a non-atomic open/readiness/camera
// operation. Acknowledgment is not proof of native readiness. A false value is
// not proof that nothing happened. Never automatically replay this operation.
type OpenCanvasError struct {
	Stage               string // "open", "readiness", or "camera"
	CommandAcknowledged bool
	Err                 error
}

// Error describes the failing stage and known HTTP acknowledgment.
func (e *OpenCanvasError) Error() string {
	return fmt.Sprintf("open canvas %s (command acknowledged=%t): %v", e.Stage, e.CommandAcknowledged, e.Err)
}

// Unwrap retains API status, camera partial outcomes and cancellation causes.
func (e *OpenCanvasError) Unwrap() error { return e.Err }
