package model

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID             uuid.UUID  `json:"id"`
	ModelType      string     `json:"model_type"`
	ModelID        uuid.UUID  `json:"model_id"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	ParentID       *uuid.UUID `json:"parent_id"`
	RecordLeft     int64      `json:"record_left"`
	RecordRight    int64      `json:"record_right"`
	RecordDepth    int64      `json:"record_depth"`
	RecordOrdering int64      `json:"record_ordering"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	UpdatedBy      uuid.UUID  `json:"updated_by"`
	DeletedBy      uuid.UUID  `json:"deleted_by"`
	DeletedAt      *time.Time `json:"deleted_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Contents       []Content  `json:"contents,omitempty"`
}
