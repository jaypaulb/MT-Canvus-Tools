// Batch operations for efficient bulk work on Canvus resources.
package canvus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchOperationType enumerates supported batch operation types.
//
// Note: per coverage matrix and Phase 3 work item #4, non-spec widget
// operations (move/copy/pin/unpin via /widgets/{id}/...) have been removed.
// Cross-canvas widget cloning is now done via the CloneWidget helper using
// the standard create endpoints, so a `Clone` batch type would be a thin
// wrapper. Until a real caller appears we expose only the operations that
// have a clean SDK API: canvas-level move/copy, and resource delete.
type BatchOperationType string

// Supported batch operation types.
const (
	BatchOperationMove   BatchOperationType = "move"
	BatchOperationCopy   BatchOperationType = "copy"
	BatchOperationDelete BatchOperationType = "delete"
)

// BatchOperation describes one operation in a batch.
type BatchOperation struct {
	ID       string
	Type     BatchOperationType
	Resource any
	Target   any
	Metadata map[string]any
}

// BatchResult captures the outcome of a single batch operation.
type BatchResult struct {
	OperationID string
	Success     bool
	Error       error
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Retries     int
}

// BatchConfig configures BatchProcessor behavior.
type BatchConfig struct {
	MaxConcurrency   int
	Timeout          time.Duration
	RetryAttempts    int
	RetryDelay       time.Duration
	ContinueOnError  bool
	ProgressCallback func(completed, total int, results []*BatchResult)
}

// DefaultBatchConfig returns sane defaults for batch processing.
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		MaxConcurrency:  10,
		Timeout:         5 * time.Minute,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		ContinueOnError: true,
	}
}

// BatchProcessor executes BatchOperations with bounded concurrency.
type BatchProcessor struct {
	session *Session
	config  *BatchConfig
	sem     chan struct{}
}

// NewBatchProcessor creates a new processor.
func NewBatchProcessor(session *Session, config *BatchConfig) *BatchProcessor {
	if config == nil {
		config = DefaultBatchConfig()
	}
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 1
	}
	if config.MaxConcurrency > 100 {
		config.MaxConcurrency = 100
	}
	return &BatchProcessor{
		session: session,
		config:  config,
		sem:     make(chan struct{}, config.MaxConcurrency),
	}
}

// ExecuteBatch runs all operations concurrently subject to MaxConcurrency.
func (bp *BatchProcessor) ExecuteBatch(ctx context.Context, operations []*BatchOperation) ([]*BatchResult, error) {
	if len(operations) == 0 {
		return []*BatchResult{}, nil
	}
	if bp.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, bp.config.Timeout)
		defer cancel()
	}

	results := make([]*BatchResult, len(operations))
	resultsChan := make(chan *BatchResult, len(operations))
	var wg sync.WaitGroup

	for i, op := range operations {
		wg.Add(1)
		go func(idx int, operation *BatchOperation) {
			defer wg.Done()
			bp.sem <- struct{}{}
			defer func() { <-bp.sem }()
			result := bp.executeOperation(ctx, operation)
			results[idx] = result
			resultsChan <- result
		}(i, op)
	}
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var completed []*BatchResult
	for r := range resultsChan {
		completed = append(completed, r)
		if bp.config.ProgressCallback != nil {
			bp.config.ProgressCallback(len(completed), len(operations), completed)
		}
	}
	if ctx.Err() != nil {
		return results, fmt.Errorf("batch operation cancelled or timed out: %w", ctx.Err())
	}
	return results, nil
}

func (bp *BatchProcessor) executeOperation(ctx context.Context, op *BatchOperation) *BatchResult {
	result := &BatchResult{OperationID: op.ID, StartTime: time.Now()}

	for attempt := 0; attempt <= bp.config.RetryAttempts; attempt++ {
		result.Retries = attempt
		var err error
		switch op.Type {
		case BatchOperationMove:
			err = bp.executeMove(ctx, op)
		case BatchOperationCopy:
			err = bp.executeCopy(ctx, op)
		case BatchOperationDelete:
			err = bp.executeDelete(ctx, op)
		default:
			err = fmt.Errorf("unsupported operation type: %s", op.Type)
		}
		if err == nil {
			result.Success = true
			break
		}
		result.Error = err
		if attempt == bp.config.RetryAttempts || ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		case <-time.After(bp.config.RetryDelay):
		}
	}
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result
}

func (bp *BatchProcessor) executeMove(ctx context.Context, op *BatchOperation) error {
	canvas, ok := op.Resource.(*Canvas)
	if !ok {
		return fmt.Errorf("move operation only supports *Canvas resources")
	}
	target, ok := op.Target.(string)
	if !ok {
		return fmt.Errorf("move target must be a folder ID string")
	}
	_, err := bp.session.MoveCanvas(ctx, canvas.ID, MoveOrCopyCanvasRequest{FolderID: target})
	return err
}

func (bp *BatchProcessor) executeCopy(ctx context.Context, op *BatchOperation) error {
	canvas, ok := op.Resource.(*Canvas)
	if !ok {
		return fmt.Errorf("copy operation only supports *Canvas resources")
	}
	target, ok := op.Target.(string)
	if !ok {
		return fmt.Errorf("copy target must be a folder ID string")
	}
	_, err := bp.session.CopyCanvas(ctx, canvas.ID, MoveOrCopyCanvasRequest{FolderID: target})
	return err
}

func (bp *BatchProcessor) executeDelete(ctx context.Context, op *BatchOperation) error {
	switch r := op.Resource.(type) {
	case *Canvas:
		return bp.session.DeleteCanvas(ctx, r.ID)
	case *Widget:
		canvasID, hasCanvas := op.Metadata["canvas_id"].(string)
		widgetType, hasType := op.Metadata["widget_type"].(string)
		if !hasCanvas || !hasType {
			return fmt.Errorf("delete operation requires canvas_id and widget_type in metadata")
		}
		return bp.session.DeleteWidget(ctx, canvasID, r.ID, widgetType)
	case *User:
		return bp.session.DeleteUser(ctx, r.ID)
	default:
		return fmt.Errorf("unsupported resource type for delete operation")
	}
}

// BatchOperationBuilder helps build batch operations fluently.
type BatchOperationBuilder struct {
	operations []*BatchOperation
}

// NewBatchOperationBuilder creates a new builder.
func NewBatchOperationBuilder() *BatchOperationBuilder {
	return &BatchOperationBuilder{operations: make([]*BatchOperation, 0)}
}

// Move adds a canvas-move operation to the batch.
func (b *BatchOperationBuilder) Move(id string, canvas *Canvas, targetFolderID string) *BatchOperationBuilder {
	b.operations = append(b.operations, &BatchOperation{
		ID: id, Type: BatchOperationMove, Resource: canvas, Target: targetFolderID,
	})
	return b
}

// Copy adds a canvas-copy operation to the batch.
func (b *BatchOperationBuilder) Copy(id string, canvas *Canvas, targetFolderID string) *BatchOperationBuilder {
	b.operations = append(b.operations, &BatchOperation{
		ID: id, Type: BatchOperationCopy, Resource: canvas, Target: targetFolderID,
	})
	return b
}

// Delete adds a delete operation to the batch.
func (b *BatchOperationBuilder) Delete(id string, resource any, metadata ...map[string]any) *BatchOperationBuilder {
	op := &BatchOperation{ID: id, Type: BatchOperationDelete, Resource: resource}
	if len(metadata) > 0 {
		op.Metadata = metadata[0]
	}
	b.operations = append(b.operations, op)
	return b
}

// Build returns the accumulated operations slice.
func (b *BatchOperationBuilder) Build() []*BatchOperation { return b.operations }

// BatchSummary summarises a slice of batch results.
type BatchSummary struct {
	TotalOperations  int
	Successful       int
	Failed           int
	TotalDuration    time.Duration
	AverageDuration  time.Duration
	FailedOperations []*BatchResult
}

// Summarize aggregates per-op results.
func Summarize(results []*BatchResult) *BatchSummary {
	summary := &BatchSummary{TotalOperations: len(results), FailedOperations: []*BatchResult{}}
	var total time.Duration
	for _, r := range results {
		if r == nil {
			continue
		}
		if r.Success {
			summary.Successful++
		} else {
			summary.Failed++
			summary.FailedOperations = append(summary.FailedOperations, r)
		}
		total += r.Duration
	}
	summary.TotalDuration = total
	if summary.TotalOperations > 0 {
		summary.AverageDuration = total / time.Duration(summary.TotalOperations)
	}
	return summary
}
