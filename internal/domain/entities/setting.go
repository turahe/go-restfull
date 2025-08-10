package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Setting represents a setting entity in the domain layer
type Setting struct {
	ID        uuid.UUID  `json:"id"`
	ModelType string     `json:"model_type,omitempty"`
	ModelID   *uuid.UUID `json:"model_id,omitempty"`
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewSetting creates a new setting instance
func NewSetting(key, value, modelType string, modelID *uuid.UUID) (*Setting, error) {
	if key == "" {
		return nil, errors.New("key is required")
	}

	now := time.Now()
	return &Setting{
		ID:        uuid.New(),
		ModelType: modelType,
		ModelID:   modelID,
		Key:       key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateSetting updates setting information
func (s *Setting) UpdateSetting(key, value, modelType string, modelID *uuid.UUID) error {
	if key != "" {
		s.Key = key
	}
	if value != "" {
		s.Value = value
	}
	if modelType != "" {
		s.ModelType = modelType
	}
	if modelID != nil {
		s.ModelID = modelID
	}
	s.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the setting as deleted
func (s *Setting) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.UpdatedAt = now
}

// IsDeleted checks if the setting is soft deleted
func (s *Setting) IsDeleted() bool {
	return s.DeletedAt != nil
}

// Validate validates the setting
func (s *Setting) Validate() error {
	if s.Key == "" {
		return errors.New("key is required")
	}
	return nil
}
