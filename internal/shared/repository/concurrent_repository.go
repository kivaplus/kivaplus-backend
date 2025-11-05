package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/kivaplus/kivaplus-backend/internal/shared/concurrent"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// ConcurrentRepository provides concurrent database operations
type ConcurrentRepository struct {
	db        *database.Connection
	poolMgr   *database.PoolManager
	processor *concurrent.BatchProcessor
}

// NewConcurrentRepository creates a new concurrent repository
func NewConcurrentRepository(db *database.Connection) *ConcurrentRepository {
	return &ConcurrentRepository{
		db:        db,
		poolMgr:   database.NewPoolManager(db.DB),
		processor: concurrent.NewBatchProcessor(),
	}
}

// BatchGetByIDs fetches multiple records by IDs concurrently
func (r *ConcurrentRepository) BatchGetByIDs(
	ctx context.Context,
	ids []int64,
	query string,
	scanner func(*sql.Rows) (interface{}, error),
) ([]interface{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Process IDs in batches to avoid overwhelming the database
	batchSize := 10
	var allResults []interface{}
	var mu sync.Mutex

	batches := make([][]int64, 0, (len(ids)+batchSize-1)/batchSize)
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batches = append(batches, ids[i:end])
	}

	// Convert batches to interface{} for ProcessBatch
	batchInterfaces := make([]interface{}, len(batches))
	for i, batch := range batches {
		batchInterfaces[i] = batch
	}

	// Process batches concurrently
	results, errors := r.processor.ProcessBatch(ctx, batchInterfaces, func(ctx context.Context, item interface{}) (interface{}, error) {
		batch, ok := item.([]int64)
		if !ok {
			return nil, fmt.Errorf("invalid batch type")
		}
		return r.fetchBatch(ctx, batch, query, scanner)
	})

	// Combine results
	for _, result := range results {
		if batchResult, ok := result.([]interface{}); ok {
			mu.Lock()
			allResults = append(allResults, batchResult...)
			mu.Unlock()
		}
	}

	// Return first error if any
	if len(errors) > 0 {
		return allResults, errors[0]
	}

	return allResults, nil
}

// fetchBatch fetches a batch of records
func (r *ConcurrentRepository) fetchBatch(
	ctx context.Context,
	ids []int64,
	query string,
	scanner func(*sql.Rows) (interface{}, error),
) ([]interface{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Build query with placeholders
	placeholders := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = id
	}

	rows, err := r.poolMgr.GetReadConnection().QueryContext(ctx, query, placeholders...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch query: %w", err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		result, err := scanner(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// ConcurrentInsert performs multiple inserts concurrently
func (r *ConcurrentRepository) ConcurrentInsert(
	ctx context.Context,
	items []interface{},
	insertFunc func(context.Context, interface{}) error,
) []error {
	if len(items) == 0 {
		return nil
	}

	_, errors := r.processor.ProcessBatch(ctx, items, func(ctx context.Context, item interface{}) (interface{}, error) {
		err := insertFunc(ctx, item)
		return nil, err
	})

	return errors
}

// ConcurrentUpdate performs multiple updates concurrently
func (r *ConcurrentRepository) ConcurrentUpdate(
	ctx context.Context,
	items []interface{},
	updateFunc func(context.Context, interface{}) error,
) []error {
	if len(items) == 0 {
		return nil
	}

	_, errors := r.processor.ProcessBatch(ctx, items, func(ctx context.Context, item interface{}) (interface{}, error) {
		err := updateFunc(ctx, item)
		return nil, err
	})

	return errors
}

// ParallelQueries executes multiple independent queries concurrently
func (r *ConcurrentRepository) ParallelQueries(ctx context.Context, queries ...func(context.Context) error) []error {
	return r.processor.ParallelExecute(ctx, queries...)
}

// TransactionalBatch executes multiple operations in a single transaction
func (r *ConcurrentRepository) TransactionalBatch(
	ctx context.Context,
	operations func(context.Context, *sql.Tx) error,
) error {
	return r.poolMgr.ExecuteTransaction(ctx, operations)
}

// OptimizedExists checks existence of multiple items concurrently
func (r *ConcurrentRepository) OptimizedExists(
	ctx context.Context,
	items []interface{},
	existsQuery string,
	paramExtractor func(interface{}) interface{},
) (map[interface{}]bool, error) {
	if len(items) == 0 {
		return make(map[interface{}]bool), nil
	}

	results := make(map[interface{}]bool)
	var mu sync.Mutex

	_, errors := r.processor.ProcessBatch(ctx, items, func(ctx context.Context, item interface{}) (interface{}, error) {
		param := paramExtractor(item)
		var exists bool

		err := r.poolMgr.GetReadConnection().QueryRowContext(ctx, existsQuery, param).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		mu.Lock()
		results[item] = exists
		mu.Unlock()

		return nil, nil
	})

	if len(errors) > 0 {
		return results, errors[0]
	}

	return results, nil
}

// BulkUpsert performs bulk upsert operations efficiently
func (r *ConcurrentRepository) BulkUpsert(
	ctx context.Context,
	items []interface{},
	upsertQuery string,
	valueExtractor func(interface{}) []interface{},
) error {
	if len(items) == 0 {
		return nil
	}

	// Extract values for batch operation
	var values [][]interface{}
	for _, item := range items {
		values = append(values, valueExtractor(item))
	}

	return r.poolMgr.BatchInsert(ctx, upsertQuery, values)
}

// GetStats returns repository performance statistics
func (r *ConcurrentRepository) GetStats() sql.DBStats {
	return r.poolMgr.GetPoolStats()
}

// Close closes the repository and its connections
func (r *ConcurrentRepository) Close() error {
	return r.poolMgr.Close()
}
