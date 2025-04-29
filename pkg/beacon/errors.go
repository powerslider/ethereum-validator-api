package beacon

import (
	"fmt"
)

// APIError defines an API error structure.
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func ErrUnexpectedStatusCode(statusCode int) error {
	//nolint:wrapcheck
	return fmt.Errorf("unexpected status code: %d", statusCode)
}
