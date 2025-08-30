package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/qor/oss"
	"github.com/qor/oss/filesystem"
	"github.com/qor/oss/s3"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// StorageProvider represents the type of storage backend
type StorageProvider string

const (
	StorageLocal   StorageProvider = "local"
	StorageMinIO   StorageProvider = "minio"
	StorageS3      StorageProvider = "s3"
	StorageAzure   StorageProvider = "azure"
	StorageGCS     StorageProvider = "gcs"
	StorageAlibaba StorageProvider = "alibaba"
	StorageTencent StorageProvider = "tencent"
	StorageQiniu   StorageProvider = "qiniu"
)

// StorageConfig holds configuration for different storage providers
type StorageConfig struct {
	Provider StorageProvider `json:"provider" yaml:"provider"`

	// Local filesystem config
	LocalPath string `json:"local_path" yaml:"local_path"`

	// S3/MinIO config
	AccessKey string `json:"access_key" yaml:"access_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	Region    string `json:"region" yaml:"region"`
	Bucket    string `json:"bucket" yaml:"bucket"`
	Endpoint  string `json:"endpoint" yaml:"endpoint"`

	// Additional configs for specific providers
	ACL string `json:"acl" yaml:"acl"`
}

// StorageService provides a unified interface for file operations across different storage providers
type StorageService struct {
	config  StorageConfig
	storage oss.StorageInterface
}

// NewStorageService creates a new storage service with the specified configuration
func NewStorageService(config StorageConfig) (*StorageService, error) {
	var storage oss.StorageInterface

	switch config.Provider {
	case StorageLocal:
		storage = filesystem.New(config.LocalPath)
	case StorageMinIO, StorageS3:
		storage = s3.New(&s3.Config{
			AccessID:  config.AccessKey,
			AccessKey: config.SecretKey,
			Region:    config.Region,
			Bucket:    config.Bucket,
			Endpoint:  config.Endpoint,
		})
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	return &StorageService{
		config:  config,
		storage: storage,
	}, nil
}

// UploadFile uploads a file to the configured storage provider
func (s *StorageService) UploadFile(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*entities.Media, error) {
	// Validate file
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Create storage path
	storagePath := fmt.Sprintf("/uploads/%s/%s", userID.String(), fileName)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Calculate file hash
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Reopen the file for upload since we consumed it for hash calculation
	src, err = file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to reopen uploaded file: %w", err)
	}
	defer src.Close()

	// Upload to storage
	if _, err := s.storage.Put(storagePath, src); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	// Generate public URL
	if _, err := s.storage.GetURL(storagePath); err != nil {
		return nil, fmt.Errorf("failed to generate public URL: %w", err)
	}

	// Create media entity
	media, err := entities.NewMedia(
		fileName,      // name
		file.Filename, // fileName
		fileHash,      // hash (actual file hash)
		string(s.config.Provider),                   // disk (storage provider)
		file.Header.Get("Content-Type"),             // mimeType
		file.Size,                                   // size
		userID,                                      // userID
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create media entity: %w", err)
	}

	// Set additional storage-specific fields
	media.Disk = string(s.config.Provider)

	// Store the storage path and public URL in custom attributes or extend the entity
	// For now, we'll use the existing fields

	return media, nil
}

// GetFile retrieves a file from storage
func (s *StorageService) GetFile(ctx context.Context, path string) (*os.File, error) {
	return s.storage.Get(path)
}

// GetFileStream retrieves a file stream from storage
func (s *StorageService) GetFileStream(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.storage.GetStream(path)
}

// DeleteFile deletes a file from storage
func (s *StorageService) DeleteFile(ctx context.Context, path string) error {
	return s.storage.Delete(path)
}

// GetFileURL generates a public URL for a file
func (s *StorageService) GetFileURL(ctx context.Context, path string) (string, error) {
	return s.storage.GetURL(path)
}

// ListFiles lists files in a directory
func (s *StorageService) ListFiles(ctx context.Context, path string) ([]*oss.Object, error) {
	return s.storage.List(path)
}

// GetStorageInfo returns information about the current storage configuration
func (s *StorageService) GetStorageInfo() StorageConfig {
	return s.config
}

// GetProvider returns the current storage provider
func (s *StorageService) GetProvider() StorageProvider {
	return s.config.Provider
}

// IsLocalStorage checks if the current storage is local filesystem
func (s *StorageService) IsLocalStorage() bool {
	return s.config.Provider == StorageLocal
}

// IsCloudStorage checks if the current storage is cloud-based
func (s *StorageService) IsCloudStorage() bool {
	return s.config.Provider != StorageLocal
}
