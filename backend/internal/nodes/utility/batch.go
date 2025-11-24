package utility

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrBatchEmpty   = errors.New("batch is empty")
	ErrBatchTimeout = errors.New("batch timeout exceeded")
)

// BatchProcessor processes items in batches
type BatchProcessor struct {
	batchSize     int
	flushInterval time.Duration
	processor     func([]interface{}) error
	buffer        []interface{}
	mu            sync.Mutex
	timer         *time.Timer
	ctx           context.Context
	cancel        context.CancelFunc
}

// BatchConfig holds batch processor configuration
type BatchConfig struct {
	BatchSize     int           // Max items per batch
	FlushInterval time.Duration // Max time to wait before flushing
	Processor     func([]interface{}) error
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(config BatchConfig) *BatchProcessor {
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	bp := &BatchProcessor{
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
		processor:     config.Processor,
		buffer:        make([]interface{}, 0, config.BatchSize),
		ctx:           ctx,
		cancel:        cancel,
	}

	return bp
}

// Add adds an item to the batch
func (bp *BatchProcessor) Add(item interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, item)

	// Reset timer
	if bp.timer != nil {
		bp.timer.Stop()
	}
	bp.timer = time.AfterFunc(bp.flushInterval, func() {
		bp.Flush()
	})

	// Flush if batch is full
	if len(bp.buffer) >= bp.batchSize {
		return bp.flush()
	}

	return nil
}

// Flush flushes the current batch
func (bp *BatchProcessor) Flush() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.flush()
}

// flush internal flush without locking
func (bp *BatchProcessor) flush() error {
	if len(bp.buffer) == 0 {
		return nil
	}

	// Stop timer
	if bp.timer != nil {
		bp.timer.Stop()
	}

	// Process batch
	batch := make([]interface{}, len(bp.buffer))
	copy(batch, bp.buffer)
	bp.buffer = bp.buffer[:0] // Clear buffer

	// Process in background
	go func() {
		if bp.processor != nil {
			bp.processor(batch)
		}
	}()

	return nil
}

// ProcessBatch processes a batch of items
func (bp *BatchProcessor) ProcessBatch(items []interface{}) error {
	if len(items) == 0 {
		return ErrBatchEmpty
	}

	if bp.processor == nil {
		return errors.New("no processor configured")
	}

	return bp.processor(items)
}

// Close closes the batch processor and flushes remaining items
func (bp *BatchProcessor) Close() error {
	bp.cancel()
	return bp.Flush()
}

// BatchBySize splits data into batches by size
func BatchBySize(data []interface{}, size int) [][]interface{} {
	var batches [][]interface{}

	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		batches = append(batches, data[i:end])
	}

	return batches
}

// BatchByCount splits data into N batches
func BatchByCount(data []interface{}, count int) [][]interface{} {
	if count <= 0 {
		return [][]interface{}{data}
	}

	size := (len(data) + count - 1) / count
	return BatchBySize(data, size)
}

// ProcessBatchesParallel processes batches in parallel
func ProcessBatchesParallel(batches [][]interface{}, processor func([]interface{}) error, maxConcurrency int) error {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	semaphore := make(chan struct{}, maxConcurrency)
	errChan := make(chan error, len(batches))
	var wg sync.WaitGroup

	for _, batch := range batches {
		wg.Add(1)
		go func(b []interface{}) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := processor(b); err != nil {
				errChan <- err
			}
		}(batch)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// ProcessBatchesSequential processes batches sequentially
func ProcessBatchesSequential(batches [][]interface{}, processor func([]interface{}) error) error {
	for _, batch := range batches {
		if err := processor(batch); err != nil {
			return err
		}
	}
	return nil
}

// BatchWithTimeout processes batch with timeout
func BatchWithTimeout(batch []interface{}, processor func([]interface{}) error, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- processor(batch)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return ErrBatchTimeout
	}
}

// ChunkData splits data into chunks for streaming
func ChunkData(data []interface{}, chunkSize int) <-chan []interface{} {
	chunks := make(chan []interface{})

	go func() {
		defer close(chunks)

		for i := 0; i < len(data); i += chunkSize {
			end := i + chunkSize
			if end > len(data) {
				end = len(data)
			}
			chunks <- data[i:end]
		}
	}()

	return chunks
}
