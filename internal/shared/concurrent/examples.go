package concurrent

import (
	"context"
	"fmt"
	"strconv"
)

// Example usage of the concurrent processing utilities

// ExampleBatchProcessor demonstrates how to use the BatchProcessor
func ExampleBatchProcessor() {
	ctx := context.Background()
	processor := NewBatchProcessor()

	// Example 1: Using non-generic ProcessBatch method
	items := []interface{}{1, 2, 3, 4, 5}

	results, errors := processor.ProcessBatch(ctx, items, func(ctx context.Context, item interface{}) (interface{}, error) {
		num := item.(int)
		return num * 2, nil
	})

	fmt.Printf("Results: %v, Errors: %v\n", results, errors)

	// Example 2: Using generic ProcessBatchGeneric function (recommended for type safety)
	numbers := []int{1, 2, 3, 4, 5}

	doubledNumbers, errs := ProcessBatchGeneric(ctx, numbers, func(ctx context.Context, num int) (int, error) {
		return num * 2, nil
	}, 3) // Use 3 workers

	fmt.Printf("Doubled numbers: %v, Errors: %v\n", doubledNumbers, errs)

	// Example 3: Using ConcurrentMap for simple transformations
	strings := []string{"1", "2", "3", "4", "5"}

	integers := ConcurrentMap(ctx, strings, func(s string) int {
		num, _ := strconv.Atoi(s)
		return num
	}, 2) // Use 2 workers

	fmt.Printf("Converted to integers: %v\n", integers)
}

// ExampleDatabaseBatch shows how to use concurrent processing for database operations
func ExampleDatabaseBatch() {
	ctx := context.Background()

	// Simulate user IDs to fetch
	userIDs := []int64{1, 2, 3, 4, 5}

	// Use generic function for type-safe database operations
	users, errors := ProcessBatchGeneric(ctx, userIDs, func(ctx context.Context, userID int64) (*User, error) {
		// Simulate database fetch
		return fetchUserFromDB(ctx, userID)
	}, 3)

	fmt.Printf("Fetched %d users with %d errors\n", len(users), len(errors))
}

// User represents a user entity
type User struct {
	ID   int64
	Name string
}

// fetchUserFromDB simulates a database fetch operation
func fetchUserFromDB(ctx context.Context, userID int64) (*User, error) {
	// Simulate database operation
	return &User{
		ID:   userID,
		Name: fmt.Sprintf("User %d", userID),
	}, nil
}

// ExampleParallelOperations shows how to execute multiple independent operations
func ExampleParallelOperations() {
	ctx := context.Background()
	processor := NewBatchProcessor()

	// Execute multiple independent operations concurrently
	errors := processor.ParallelExecute(ctx,
		func(ctx context.Context) error {
			// Operation 1: Update cache
			fmt.Println("Updating cache...")
			return nil
		},
		func(ctx context.Context) error {
			// Operation 2: Send notification
			fmt.Println("Sending notification...")
			return nil
		},
		func(ctx context.Context) error {
			// Operation 3: Log activity
			fmt.Println("Logging activity...")
			return nil
		},
	)

	if len(errors) > 0 {
		fmt.Printf("Some operations failed: %v\n", errors)
	} else {
		fmt.Println("All operations completed successfully")
	}
}

// ExampleSafeGoroutine demonstrates panic-safe goroutine execution
func ExampleSafeGoroutine() {
	SafeGoroutine(func() {
		// This might panic
		panic("something went wrong")
	}, func(recovered interface{}) {
		fmt.Printf("Recovered from panic: %v\n", recovered)
	})

	// Program continues normally
	fmt.Println("Program continues after panic recovery")
}
