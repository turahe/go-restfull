package model

import (
	"time"

	"github.com/google/uuid"
)

// Taxonomy represents a taxonomy entity
type Taxonomy struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name" validate:"required"`
	Description string     `json:"description" db:"description" validate:"required"`
	RecordLeft  uint64     `json:"recordLeft"`
	RecordRight uint64     `json:"recordRight"`
	RecordDepth uint64     `json:"recordDepth"`
	ParentID    *uuid.UUID `json:"parent_id" db:"parent_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
