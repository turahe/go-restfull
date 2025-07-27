package model

import (
	"time"
)

// Taxonomy represents a taxonomy database model
type Taxonomy struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Code        string     `json:"code"`
	Description string     `json:"description"`
	ParentID    *string    `json:"parent_id"`
	RecordLeft  int64      `json:"record_left"`
	RecordRight int64      `json:"record_right"`
	RecordDepth int64      `json:"record_depth"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedBy   string     `json:"created_by"`
	UpdatedBy   string     `json:"updated_by"`
	DeletedBy   string     `json:"deleted_by"`
}
