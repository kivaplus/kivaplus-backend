package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// PoolManager manages database connection pools for concurrent operations
type PoolManager struct {
	db           *sql.DB
	readPool     *sql.DB
	writePool    *sql.DB
	maxOpenConns int
	maxIdleConns int
	mu           sync.RWMutex
}

// NewPoolManager creates a new pool manager with optimized settings
func NewPoolManager(db *sql.DB) *PoolManager {
	pm := &PoolManager{
		db:           db,
		maxOpenConns: 25, // Optimal for most applications
		maxIdleConns: 10, // Keep some connections idle for quick reuse
	}

	// Configure connection pool settings
	pm.configurePool(db)

	return pm
}

// configurePool sets optimal connection pool parameters
func (pm *PoolManager) configurePool(db *sql.DB) {
	// Set maximum number of open connections
	db.SetMaxOpenConns(pm.maxOpenConns)

	// Set maximum number of idle connections
	db.SetMaxIdleConns(pm.maxIdleConns)

	// Set maximum lifetime of connections
	db.SetConnMaxLifetime(time.Hour)

	// Set maximum idle time for connections
	db.SetConnMaxIdleTime(time.Minute * 30)
}

// GetReadConnection returns a connection optimized for read operations
func (pm *PoolManager) GetReadConnection() *sql.DB {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.readPool != nil {
		return pm.readPool
	}
	return pm.db
}

// GetWriteConnection returns a connection optimized for write operations
func (pm *PoolManager) GetWriteConnection() *sql.DB {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.writePool != nil {
		return pm.writePool
	}
	return pm.db
}

// ExecuteConcurrentReads executes multiple read queries concurrently
func (pm *PoolManager) ExecuteConcurrentReads(ctx context.Context, queries []func(context.Context, *sql.DB) error) []error {
	if len(queries) == 0 {
		return nil
	}

	errorChan := make(chan error, len(queries))
	var wg sync.WaitGroup

	for _, query := range queries {
		wg.Add(1)
		go func(q func(context.Context, *sql.DB) error) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("panic in concurrent read: %v", r)
				}
			}()

			err := q(ctx, pm.GetReadConnection())
			errorChan <- err
		}(query)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	var errors []error
	for i := 0; i < len(queries); i++ {
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

// ExecuteTransaction executes multiple operations in a single transaction
func (pm *PoolManager) ExecuteTransaction(ctx context.Context, operations func(context.Context, *sql.Tx) error) error {
	tx, err := pm.GetWriteConnection().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := operations(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("operation failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BatchInsert performs batch insert operations efficiently
func (pm *PoolManager) BatchInsert(ctx context.Context, query string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	return pm.ExecuteTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, valueSet := range values {
			if _, err := stmt.ExecContext(ctx, valueSet...); err != nil {
				return fmt.Errorf("failed to execute batch insert: %w", err)
			}
		}

		return nil
	})
}

// GetPoolStats returns connection pool statistics
func (pm *PoolManager) GetPoolStats() sql.DBStats {
	return pm.db.Stats()
}

// Close closes all database connections
func (pm *PoolManager) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errors []error

	if pm.readPool != nil && pm.readPool != pm.db {
		if err := pm.readPool.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if pm.writePool != nil && pm.writePool != pm.db {
		if err := pm.writePool.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if err := pm.db.Close(); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	return nil
}
