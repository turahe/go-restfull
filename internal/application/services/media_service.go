package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// mediaService implements MediaService interface
type mediaService struct {
	mediaRepository repositories.MediaRepository
}

// NewMediaService creates a new media service
func NewMediaService(mediaRepository repositories.MediaRepository) ports.MediaService {
	return &mediaService{
		mediaRepository: mediaRepository,
	}
}

// UploadMedia uploads a new media file
func (s *mediaService) UploadMedia(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*entities.Media, error) {
	// Validate file
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Create media entity
	media, err := entities.NewMedia(
		fileName,
		file.Filename,
		file.Header.Get("Content-Type"),
		fmt.Sprintf("/uploads/%s", fileName),
		fmt.Sprintf("/api/v1/media/%s", fileName),
		file.Size,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.mediaRepository.Create(ctx, media)
	if err != nil {
		return nil, err
	}

	return media, nil
}

// GetMediaByID retrieves media by ID
func (s *mediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	return s.mediaRepository.GetByID(ctx, id)
}

// GetMediaByUserID retrieves media by user ID
func (s *mediaService) GetMediaByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.GetByUserID(ctx, userID, limit, offset)
}

// GetAllMedia retrieves all media with pagination
func (s *mediaService) GetAllMedia(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.GetAll(ctx, limit, offset)
}

// SearchMedia searches media by query
func (s *mediaService) SearchMedia(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	return s.mediaRepository.Search(ctx, query, limit, offset)
}

// UpdateMedia updates media information
func (s *mediaService) UpdateMedia(ctx context.Context, id uuid.UUID, fileName, originalName, mimeType, path, url string, size int64) (*entities.Media, error) {
	media, err := s.mediaRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = media.UpdateMedia(fileName, originalName, mimeType, path, url, size)
	if err != nil {
		return nil, err
	}

	err = s.mediaRepository.Update(ctx, media)
	if err != nil {
		return nil, err
	}

	return media, nil
}

// DeleteMedia deletes media
func (s *mediaService) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	return s.mediaRepository.Delete(ctx, id)
}

// GetMediaCount returns the total number of media
func (s *mediaService) GetMediaCount(ctx context.Context) (int64, error) {
	return s.mediaRepository.Count(ctx)
}

// GetMediaCountByUserID returns the total number of media by user ID
func (s *mediaService) GetMediaCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.mediaRepository.CountByUserID(ctx, userID)
}
