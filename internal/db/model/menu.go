package model

import (
	"time"
)

type Menu struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description"`
	URL            string     `json:"url"`
	Icon           string     `json:"icon"`
	ParentID       *string    `json:"parent_id"`
	RecordLeft     int64      `json:"record_left"`
	RecordRight    int64      `json:"record_right"`
	RecordOrdering int64      `json:"record_ordering"`
	IsActive       bool       `json:"is_active"`
	IsVisible      bool       `json:"is_visible"`
	Target         string     `json:"target"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	CreatedBy      string     `json:"created_by"`
	UpdatedBy      string     `json:"updated_by"`
	DeletedBy      string     `json:"deleted_by"`
}
