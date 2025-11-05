package usecases

import "strings"

// isPersonNotFoundError checks if the error is a "person not found" error
func isPersonNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "person not found")
}
