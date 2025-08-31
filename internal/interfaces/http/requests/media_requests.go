// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// UpdateMediaRequest represents the request for updating media metadata.
// This struct defines the fields that can be updated for a media entity,
// including validation tags for field constraints and business rules.
type UpdateMediaRequest struct {
	// Name is the display name for the media file (optional, max 255 characters if provided)
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// FileName is the filename of the media file (optional, max 255 characters if provided)
	FileName string `json:"file_name,omitempty" validate:"omitempty,min=1,max=255"`
	// Hash is the file hash for integrity verification (optional, max 255 characters if provided)
	Hash string `json:"hash,omitempty" validate:"omitempty,min=1,max=255"`
	// Disk is the storage disk identifier (optional, max 100 characters if provided)
	Disk string `json:"disk,omitempty" validate:"omitempty,min=1,max=100"`
	// MimeType is the MIME type of the file (optional, max 100 characters if provided)
	MimeType string `json:"mime_type,omitempty" validate:"omitempty,min=1,max=100"`
	// Size is the file size in bytes (optional, must be > 0 if provided)
	Size *int64 `json:"size,omitempty" validate:"omitempty,gt=0"`
}

// CreateMediaFormRequest represents the form data for creating a new media entity.
// This struct defines the form fields for media upload, including validation tags.
type CreateMediaFormRequest struct {
	// Name is the custom name for the media file (optional, max 255 characters if provided)
	Name string `form:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Description provides additional details about the media file (optional, max 1000 characters if provided)
	Description string `form:"description,omitempty" validate:"omitempty,max=1000"`
}

// Validate performs validation on the UpdateMediaRequest using the validator package.
// This method checks all field constraints including length limits and size validation.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *UpdateMediaRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs validation on the CreateMediaFormRequest using the validator package.
// This method checks all field constraints including length limits.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateMediaFormRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the UpdateMediaRequest to update an existing Media domain entity.
// This method updates the media entity with the new values provided in the request.
//
// Parameters:
//   - existingMedia: The existing media entity to update
//
// Returns:
//   - *entities.Media: The updated media entity
func (r *UpdateMediaRequest) ToEntity(existingMedia *entities.Media) *entities.Media {
	// Update fields if provided, otherwise keep existing values
	if r.Name != "" {
		existingMedia.Name = r.Name
	}
	if r.FileName != "" {
		existingMedia.FileName = r.FileName
	}
	if r.Hash != "" {
		existingMedia.Hash = r.Hash
	}
	if r.Disk != "" {
		existingMedia.Disk = r.Disk
	}
	if r.MimeType != "" {
		existingMedia.MimeType = r.MimeType
	}
	if r.Size != nil {
		existingMedia.Size = *r.Size
	}

	return existingMedia
}

// ToEntity transforms the CreateMediaFormRequest to create a new Media domain entity.
// This method creates a new media entity with the provided form data.
//
// Parameters:
//   - fileName: The original filename of the uploaded file
//   - hash: The file hash for integrity verification
//   - disk: The storage disk identifier
//   - mimeType: The MIME type of the file
//   - size: The file size in bytes
//   - userID: The UUID of the user uploading the media
//
// Returns:
//   - *entities.Media: The created media entity
//   - error: Any error that occurred during entity creation
func (r *CreateMediaFormRequest) ToEntity(fileName, hash, disk, mimeType string, size int64, userID uuid.UUID) (*entities.Media, error) {
	// Use provided name or default to original filename
	name := r.Name
	if name == "" {
		name = fileName
	}

	// Create and populate the media entity
	media := &entities.Media{
		ID:        uuid.New(),
		Name:      name,
		FileName:  fileName,
		Hash:      hash,
		Disk:      disk,
		MimeType:  mimeType,
		Size:      size,
		CreatedBy: userID,
	}

	return media, nil
}
