package helpers

import "github.com/pborman/uuid"

// NewRequestID generates a random ID to use for traking of log messages per request
func NewRequestID() string {
	return uuid.New()[:8]
}
