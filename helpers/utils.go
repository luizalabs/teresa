package helpers

import "github.com/satori/go.uuid"

// NewShortUUID generates a random ID to use for traking of log messages per request
func NewShortUUID() string {
	return uuid.NewV4().String()[:8]
}
