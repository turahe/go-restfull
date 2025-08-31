// Package services provides application-level business logic for media management.
// This package contains the media service implementation that handles file uploads,
// media storage, retrieval, and file management while ensuring secure and efficient
// media handling.
package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/pkg/logger"
	"github.com/turahe/go-restfull/pkg/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// mediaService implements the MediaService interface and provides comprehensive
// media management functionality. It handles file uploads, media storage, retrieval,
// search capabilities, and file metadata management while ensuring secure file handling.
type mediaService struct {
	mediaRepository repositories.MediaRepository
	storageService  *storage.StorageService
}

// NewMediaService creates a new media service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - mediaRepository: Repository interface for media data access operations
//   - storageService: Storage service for file operations
//
// Returns:
//   - ports.MediaService: The media service interface implementation
func NewMediaService(mediaRepository repositories.MediaRepository, storageService *storage.StorageService) ports.MediaService {
	return &mediaService{
		mediaRepository: mediaRepository,
		storageService:  storageService,
	}
}

// UploadMedia uploads a new media file and creates the corresponding media entity.
// This method handles file validation, unique filename generation, and metadata
// extraction for secure and organized file storage.
//
// Business Rules:
//   - File must be provided and valid
//   - Unique filename is generated to prevent conflicts
//   - File metadata is extracted and stored
//   - User ID is associated with the upload for ownership tracking
//   - File paths and URLs are generated for access
//
// Security Features:
//   - Unique filename generation prevents path traversal attacks
//   - File type validation through MIME type checking
//   - User ownership tracking for access control
//
// Parameters:
//   - ctx: Context for the operation
//   - file: Multipart file header containing the uploaded file
//   - userID: UUID of the user uploading the file
//
// Returns:
//   - *entities.Media: The created media entity
//   - error: Any error that occurred during the operation
func (s *mediaService) UploadMedia(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*entities.Media, error) {
	logger.Log.Info("UploadMedia: Starting media upload",
		zap.String("user_id", userID.String()),
		zap.String("filename", file.Filename),
		zap.Int64("file_size", file.Size),
		zap.String("content_type", file.Header.Get("Content-Type")),
	)

	// Use storage service to upload file and create media entity
	media, err := s.storageService.UploadFile(ctx, file, userID)
	if err != nil {
		logger.Log.Error("UploadMedia: Storage service upload failed",
			zap.String("user_id", userID.String()),
			zap.String("filename", file.Filename),
			zap.Error(err),
		)

		// Provide more specific error messages based on the error type
		if strings.Contains(err.Error(), "InvalidAccessKeyId") {
			return nil, fmt.Errorf("storage configuration error: invalid access credentials")
		} else if strings.Contains(err.Error(), "NoSuchBucket") {
			return nil, fmt.Errorf("storage configuration error: bucket does not exist")
		} else if strings.Contains(err.Error(), "AccessDenied") {
			return nil, fmt.Errorf("storage access denied: check permissions and credentials")
		} else if strings.Contains(err.Error(), "network") || strings.Contains(err.Error(), "connection") {
			return nil, fmt.Errorf("storage connection failed: check network and endpoint configuration")
		}

		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Log.Info("UploadMedia: File uploaded successfully, persisting to database",
		zap.String("user_id", userID.String()),
		zap.String("filename", file.Filename),
		zap.String("media_id", media.ID.String()),
	)

	// Persist the media entity to the repository
	err = s.mediaRepository.Create(ctx, media)
	if err != nil {
		logger.Log.Error("UploadMedia: Database persistence failed",
			zap.String("user_id", userID.String()),
			zap.String("filename", file.Filename),
			zap.String("media_id", media.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to persist media entity: %w", err)
	}

	logger.Log.Info("UploadMedia: Media upload completed successfully",
		zap.String("user_id", userID.String()),
		zap.String("filename", file.Filename),
		zap.String("media_id", media.ID.String()),
	)

	return media, nil
}

// GetMediaByID retrieves a media file by its unique identifier.
// This method provides access to individual media details and metadata.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the media to retrieve
//
// Returns:
//   - *entities.Media: The media entity if found
//   - error: Error if media not found or other issues occur
func (s *mediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	return s.mediaRepository.GetByID(ctx, id)
}

// GetMediaByUserID retrieves all media files uploaded by a specific user with pagination.
// This method is useful for user media galleries and personal file management.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to get media for
//   - limit: Maximum number of media files to return
//   - offset: Number of media files to skip for pagination
//
// Returns:
//   - []*entities.Media: List of media files by the user
//   - error: Any error that occurred during the operation
func (s *mediaService) GetMediaByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.GetByUserID(ctx, userID, limit, offset)
}

// GetAllMedia retrieves all media files in the system with pagination.
// This method is useful for administrative purposes and system-wide media management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of media files to return
//   - offset: Number of media files to skip for pagination
//
// Returns:
//   - []*entities.Media: List of all media files
//   - error: Any error that occurred during the operation
func (s *mediaService) GetAllMedia(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.GetAll(ctx, limit, offset)
}

// SearchMedia searches for media files based on a query string.
// This method supports full-text search capabilities for finding media files
// by filename, original name, or other metadata.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Media: List of matching media files
//   - error: Any error that occurred during the operation
func (s *mediaService) SearchMedia(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.Search(ctx, query, limit, offset)
}

// UpdateMedia updates an existing media file's metadata and information.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Media must exist and be accessible
//   - Updated metadata must be provided and validated
//   - File paths and URLs are updated accordingly
//   - File size is updated for accurate storage tracking
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the media to update
//   - fileName: Updated filename
//   - originalName: Updated original filename
//   - mimeType: Updated MIME type
//   - path: Updated file path
//   - url: Updated access URL
//   - size: Updated file size in bytes
//
// Returns:
//   - *entities.Media: The updated media entity
//   - error: Any error that occurred during the operation
func (s *mediaService) UpdateMedia(ctx context.Context, id uuid.UUID, fileName, originalName, mimeType, path, url string, size int64) (*entities.Media, error) {
	// Retrieve existing media to ensure it exists and is accessible
	media, err := s.mediaRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the media entity with new metadata
	err = media.UpdateMedia(fileName, originalName, mimeType, path, url, size)
	if err != nil {
		return nil, err
	}

	// Persist the updated media to the repository
	err = s.mediaRepository.Update(ctx, media)
	if err != nil {
		return nil, err
	}

	return media, nil
}

// DeleteMedia performs a soft delete of a media file by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the media to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *mediaService) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	return s.mediaRepository.Delete(ctx, id)
}

// GetMediaCount returns the total number of media files in the system.
// This method is useful for statistics and administrative reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of media files
//   - error: Any error that occurred during the operation
func (s *mediaService) GetMediaCount(ctx context.Context) (int64, error) {
	return s.mediaRepository.Count(ctx)
}

// GetMediaCountByUserID returns the total number of media files uploaded by a specific user.
// This method is useful for user storage quotas and activity tracking.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to count media for
//
// Returns:
//   - int64: Total count of media files by the user
//   - error: Any error that occurred during the operation
func (s *mediaService) GetMediaCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.mediaRepository.CountByUserID(ctx, userID)
}

// GetAvatarByUserID retrieves the avatar image for a specific user.
// This method returns the first JPEG image uploaded by the user, which is typically used as their avatar.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to get avatar for
//
// Returns:
//   - *entities.Media: The user's avatar media entity, or nil if no avatar exists
//   - error: Any error that occurred during the operation
func (s *mediaService) GetAvatarByUserID(ctx context.Context, userID uuid.UUID) (*entities.Media, error) {
	return s.mediaRepository.GetAvatarByUserID(ctx, userID)
}

// GetMediaByGroup retrieves media by group for a specific entity.
// This method is useful for getting different types of media (avatar, cover, gallery, etc.)
// for any entity type (users, posts, organizations, etc.).
//
// Parameters:
//   - ctx: Context for the operation
//   - mediableID: UUID of the entity to get media for
//   - mediableType: Type of the entity (e.g., "users", "posts", "organizations")
//   - group: Group/category of the media (e.g., "avatar", "cover", "gallery")
//
// Returns:
//   - *entities.Media: The media entity if found, or nil if no media exists
//   - error: Any error that occurred during the operation
func (s *mediaService) GetMediaByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string) (*entities.Media, error) {
	return s.mediaRepository.GetByGroup(ctx, mediableID, mediableType, group)
}

// GetAllMediaByGroup retrieves all media by group for a specific entity with pagination.
// This method is useful for getting galleries or collections of media for any entity type.
//
// Parameters:
//   - ctx: Context for the operation
//   - mediableID: UUID of the entity to get media for
//   - mediableType: Type of the entity (e.g., "users", "posts", "organizations")
//   - group: Group/category of the media (e.g., "avatar", "cover", "gallery")
//   - limit: Maximum number of media files to return
//   - offset: Number of media files to skip for pagination
//
// Returns:
//   - []*entities.Media: List of media entities
//   - error: Any error that occurred during the operation
func (s *mediaService) GetAllMediaByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.GetAllByGroup(ctx, mediableID, mediableType, group, limit, offset)
}
