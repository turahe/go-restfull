package storage

import (
	"fmt"
	"path/filepath"
)

// StorageLoader handles storage service initialization and dependency injection
type StorageLoader struct {
	factory *StorageFactory
}

// NewStorageLoader creates a new storage loader
func NewStorageLoader(configPath string) (*StorageLoader, error) {
	// Resolve config path relative to project root
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(".", configPath)
	}

	factory, err := NewStorageFactory(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage factory: %w", err)
	}

	return &StorageLoader{
		factory: factory,
	}, nil
}

// LoadDefaultStorage loads the default storage service
func (l *StorageLoader) LoadDefaultStorage() (*StorageService, error) {
	return l.factory.CreateDefaultStorageService()
}

// LoadStorage loads a specific storage service by name
func (l *StorageLoader) LoadStorage(providerName string) (*StorageService, error) {
	return l.factory.CreateStorageService(providerName)
}

// GetAvailableProviders returns available storage providers
func (l *StorageLoader) GetAvailableProviders() []string {
	return l.factory.GetAvailableProviders()
}

// ValidateStorage validates storage configuration
func (l *StorageLoader) ValidateStorage(providerName string) error {
	return l.factory.ValidateProvider(providerName)
}

// LoadAllProviders loads all available storage providers
func (l *StorageLoader) LoadAllProviders() (map[string]*StorageService, error) {
	providers := make(map[string]*StorageService)

	// Load default storage service
	defaultStorage, err := l.LoadDefaultStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to load default storage: %w", err)
	}
	providers["default"] = defaultStorage

	// Load all available providers
	for _, provider := range l.GetAvailableProviders() {
		storage, err := l.LoadStorage(provider)
		if err != nil {
			// Log warning but continue with other providers
			fmt.Printf("Warning: Failed to load storage provider '%s': %v\n", provider, err)
			continue
		}
		providers[provider] = storage
	}

	return providers, nil
}
