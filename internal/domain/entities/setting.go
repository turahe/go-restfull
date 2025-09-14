// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Setting entity for managing
// application settings and configuration with polymorphic relationship support.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Setting represents a setting entity in the domain layer that stores
// configuration values and application settings. It supports polymorphic
// relationships through model_type and model_id fields, allowing settings
// to be associated with different entity types.
//
// The entity includes:
// - Key-value pair storage for configuration data
// - Polymorphic relationships for entity-specific settings
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for setting preservation
// - Validation support for data integrity
type Setting struct {
	ID        uuid.UUID  `json:"id"`                   // Unique identifier for the setting
	ModelType string     `json:"model_type,omitempty"` // Type of entity this setting belongs to (optional)
	ModelID   *uuid.UUID `json:"model_id,omitempty"`   // ID of the entity this setting belongs to (optional)
	Key       string     `json:"key"`                  // Setting key/name for identification
	Value     string     `json:"value"`                // Setting value/content
	KeyType   string     `json:"key_type"`             // Setting key type
	Status    bool       `json:"status"`               // Setting status
	CreatedBy uuid.UUID  `json:"created_by"`           // ID of user who created this setting
	UpdatedBy uuid.UUID  `json:"updated_by"`           // ID of user who last updated this setting
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty"` // ID of user who deleted this setting (soft delete)
	CreatedAt time.Time  `json:"created_at"`           // Timestamp when setting was created
	UpdatedAt time.Time  `json:"updated_at"`           // Timestamp when setting was last updated
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Timestamp when setting was soft deleted
}

// NewSetting creates a new setting instance with validation.
// This constructor validates required fields and initializes the setting
// with generated UUID and timestamps.
//
// Parameters:
//   - key: Setting key/name for identification (required)
//   - value: Setting value/content
//   - modelType: Type of entity this setting belongs to (optional)
//   - modelID: ID of the entity this setting belongs to (optional)
//
// Returns:
//   - *Setting: Pointer to the newly created setting entity
//   - error: Validation error if key is empty
//
// Validation rules:
// - key cannot be empty
// - value, modelType, and modelID are optional
//
// Note: ModelType and ModelID enable polymorphic relationships for entity-specific settings
func NewSetting(key, value, modelType string, modelID *uuid.UUID, keyType string, status bool) (*Setting, error) {
	// Validate required fields
	if key == "" {
		return nil, errors.New("key is required")
	}

	// Create setting with current timestamp
	now := time.Now()
	return &Setting{
		ID:        uuid.New(), // Generate new unique identifier
		ModelType: modelType,  // Set model type (can be empty)
		ModelID:   modelID,    // Set model ID (can be nil)
		Key:       key,        // Set setting key
		Value:     value,      // Set setting value
		KeyType:   keyType,    // Set setting key type
		Status:    status,     // Set setting status
		CreatedAt: now,        // Set creation timestamp
		UpdatedAt: now,        // Set initial update timestamp
	}, nil
}

// UpdateSetting updates setting information with new values.
// This method allows partial updates - only non-empty values are updated.
// The UpdatedAt timestamp is automatically updated when any field changes.
//
// Parameters:
//   - key: New setting key (optional, only updated if not empty)
//   - value: New setting value (optional, only updated if not empty)
//   - modelType: New model type (optional, only updated if not empty)
//   - modelID: New model ID (optional, only updated if not nil)
//
// Returns:
//   - error: Always nil, included for interface consistency
//
// Note: This method automatically updates the UpdatedAt timestamp
func (s *Setting) UpdateSetting(key, value, modelType string, modelID *uuid.UUID, keyType string, status bool) error {
	// Update fields only if new values are provided
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
	if keyType != "" {
		s.KeyType = keyType
	}
	if status != s.Status {
		s.Status = status
	}
	// Update modification timestamp
	s.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the setting as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The setting will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (s *Setting) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now // Set deletion timestamp
	s.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the setting has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
// This method is useful for filtering out deleted settings from queries.
//
// Returns:
//   - bool: true if setting is deleted, false if active
func (s *Setting) IsDeleted() bool {
	return s.DeletedAt != nil
}

// Validate validates the setting to ensure data integrity.
// This method checks that all required fields are properly set.
//
// Returns:
//   - error: Validation error if any required field is invalid, nil if valid
//
// Validation rules:
// - key cannot be empty
func (s *Setting) Validate() error {
	if s.Key == "" {
		return errors.New("key is required")
	}
	return nil
}
