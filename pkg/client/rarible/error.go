package rarible

import "fmt"

// Helper function for error handling.
func (e *APIError) Error() string {
	return fmt.Sprintf("integration error code: %s, description: %s", e.Code, e.Message)
}
