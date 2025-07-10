package model

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID             uuid.UUID  `json:"id"`
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Subtitle       string     `json:"subtitle"`
	Description    string     `json:"description"`
	Type           string     `json:"type"`
	IsSticky       bool       `json:"is_sticky"`
	PublishedAt    int64      `json:"published_at"`
	Language       string     `json:"language"`
	Layout         string     `json:"layout"`
	RecordOrdering int64      `json:"record_ordering"`
	Contents       []Content  `json:"contents,omitempty"`
	CreatedBy      *uuid.UUID `json:"created_by"`
	UpdatedBy      *uuid.UUID `json:"updated_by"`
	DeletedBy      *uuid.UUID `json:"deleted_by"`
	DeletedAt      *time.Time `json:"deleted_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
