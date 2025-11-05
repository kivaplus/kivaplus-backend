package concurrent

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// BatchProcessor provides utilities for concurrent batch processing
type BatchProcessor struct {
	maxWorkers int
}

// NewBatchProcessor creates a new batch processor with optimal worker count
func NewBatchProcessor() *BatchProcessor {
	// Use number of CPU cores as default worker count
	maxWorkers := runtime.NumCPU()
	if maxWorkers < 2 {
		maxWorkers = 2
	}
	if maxWorkers > 10 {
		maxWorkers = 10 // Cap at 10 to avoid overwhelming the database
	}

	return &BatchProcessor{
		maxWorkers: maxWorkers,
	}
}

// NewBatchProcessorWithWorkers creates a batch processor with specific worker count
func NewBatchProcessorWithWorkers(workers int) *BatchProcessor {
	if workers < 1 {
		workers = 1
	}
	return &BatchProcessor{
		maxWorkers: workers,
	}
}

// ProcessBatch processes a batch of items concurrently
func (bp *BatchProcessor) ProcessBatch(
	ctx context.Context,
	items []interface{},
	processor func(ctx context.Context, item interface{}) (interface{}, error),
) ([]interface{}, []error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Determine optimal number of workers for this batch
	workers := bp.maxWorkers
	if len(items) < workers {
		workers = len(items)
	}

	// Create channels for work distribution and result collection
	workChan := make(chan interface{}, len(items))
	resultChan := make(chan struct {
		result interface{}
		err    error
		index  int
	}, len(items))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
					result, err := processor(ctx, item)
					resultChan <- struct {
						result interface{}
						err    error
						index  int
					}{result, err, 0} // Index will be handled differently
				}
			}
		}()
	}

	// Send work to workers
	go func() {
		defer close(workChan)
		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case workChan <- item:
			}
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []interface{}
	var errors []error

	for i := 0; i < len(items); i++ {
		select {
		case <-ctx.Done():
			return results, append(errors, ctx.Err())
		case res := <-resultChan:
			if res.err != nil {
				errors = append(errors, res.err)
			} else {
				results = append(results, res.result)
			}
		}
	}

	return results, errors
}

// ProcessBatchWithIndex processes items while preserving order
func (bp *BatchProcessor) ProcessBatchWithIndex(
	ctx context.Context,
	items []interface{},
	processor func(ctx context.Context, item interface{}, index int) (interface{}, error),
) ([]interface{}, []error) {
	if len(items) == 0 {
		return nil, nil
	}

	type workItem struct {
		item  interface{}
		index int
	}

	type resultItem struct {
		result interface{}
		err    error
		index  int
	}

	workers := bp.maxWorkers
	if len(items) < workers {
		workers = len(items)
	}

	workChan := make(chan workItem, len(items))
	resultChan := make(chan resultItem, len(items))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
					result, err := processor(ctx, work.item, work.index)
					resultChan <- resultItem{result, err, work.index}
				}
			}
		}()
	}

	// Send work
	go func() {
		defer close(workChan)
		for i, item := range items {
			select {
			case <-ctx.Done():
				return
			case workChan <- workItem{item, i}:
			}
		}
	}()

	// Wait for completion
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results in order
	results := make([]interface{}, len(items))
	errors := make([]error, len(items))
	hasResults := make([]bool, len(items))

	for i := 0; i < len(items); i++ {
		select {
		case <-ctx.Done():
			return results[:i], errors[:i]
		case res := <-resultChan:
			if res.index < len(items) {
				results[res.index] = res.result
				errors[res.index] = res.err
				hasResults[res.index] = true
			}
		}
	}

	// Filter out empty results
	var finalResults []interface{}
	var finalErrors []error

	for i, hasResult := range hasResults {
		if hasResult {
			if errors[i] != nil {
				finalErrors = append(finalErrors, errors[i])
			} else {
				finalResults = append(finalResults, results[i])
			}
		}
	}

	return finalResults, finalErrors
}

// ParallelExecute executes multiple functions concurrently
func (bp *BatchProcessor) ParallelExecute(ctx context.Context, functions ...func(context.Context) error) []error {
	if len(functions) == 0 {
		return nil
	}

	errorChan := make(chan error, len(functions))
	var wg sync.WaitGroup

	for _, fn := range functions {
		wg.Add(1)
		go func(f func(context.Context) error) {
			defer wg.Done()
			if err := f(ctx); err != nil {
				errorChan <- err
			} else {
				errorChan <- nil
			}
		}(fn)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	var errors []error
	for i := 0; i < len(functions); i++ {
		select {
		case <-ctx.Done():
			errors = append(errors, ctx.Err())
		case err := <-errorChan:
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// SafeGoroutine executes a function in a goroutine with panic recovery
func SafeGoroutine(fn func(), onPanic func(interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if onPanic != nil {
					onPanic(r)
				}
			}
		}()
		fn()
	}()
}

// Generic standalone functions for type-safe concurrent processing

// ProcessBatchGeneric processes a batch of items concurrently with type safety
func ProcessBatchGeneric[T any, R any](
	ctx context.Context,
	items []T,
	processor func(ctx context.Context, item T) (R, error),
	maxWorkers int,
) ([]R, []error) {
	if len(items) == 0 {
		return nil, nil
	}

	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	// Determine optimal number of workers for this batch
	workers := maxWorkers
	if len(items) < workers {
		workers = len(items)
	}

	// Create channels for work distribution and result collection
	workChan := make(chan T, len(items))
	resultChan := make(chan struct {
		result R
		err    error
		index  int
	}, len(items))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
					result, err := processor(ctx, item)
					resultChan <- struct {
						result R
						err    error
						index  int
					}{result, err, 0}
				}
			}
		}()
	}

	// Send work to workers
	go func() {
		defer close(workChan)
		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case workChan <- item:
			}
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []R
	var errors []error

	for i := 0; i < len(items); i++ {
		select {
		case <-ctx.Done():
			return results, append(errors, ctx.Err())
		case res := <-resultChan:
			if res.err != nil {
				errors = append(errors, res.err)
			} else {
				results = append(results, res.result)
			}
		}
	}

	return results, errors
}

// ConcurrentMap applies a function to each item in a slice concurrently
func ConcurrentMap[T any, R any](
	ctx context.Context,
	items []T,
	mapper func(T) R,
	maxWorkers int,
) []R {
	if len(items) == 0 {
		return nil
	}

	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	if len(items) < maxWorkers {
		maxWorkers = len(items)
	}

	type workItem struct {
		item  T
		index int
	}

	type resultItem struct {
		result R
		index  int
	}

	workChan := make(chan workItem, len(items))
	resultChan := make(chan resultItem, len(items))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := mapper(work.item)
					resultChan <- resultItem{result, work.index}
				}
			}
		}()
	}

	// Send work
	go func() {
		defer close(workChan)
		for i, item := range items {
			select {
			case <-ctx.Done():
				return
			case workChan <- workItem{item, i}:
			}
		}
	}()

	// Wait for completion
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results in order
	results := make([]R, len(items))
	for i := 0; i < len(items); i++ {
		select {
		case <-ctx.Done():
			return results[:i]
		case res := <-resultChan:
			if res.index < len(items) {
				results[res.index] = res.result
			}
		}
	}

	return results
}

// TimeoutContext creates a context with timeout and automatic cleanup
func TimeoutContext(ctx context.Context, timeout int64) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
}
