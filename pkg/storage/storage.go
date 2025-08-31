package storage

import (
	"time"
)

// StorageHealth represents the health status of a storage provider
type StorageHealth struct {
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	LastCheck    time.Time              `json:"last_check"`
	Provider     string                 `json:"provider"`
	ResponseTime time.Duration          `json:"response_time"`
}
