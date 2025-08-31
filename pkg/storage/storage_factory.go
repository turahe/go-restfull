package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// StorageFactory creates storage services for different providers
type StorageFactory struct {
	config map[string]interface{}
}

// NewStorageFactory creates a new storage factory from configuration file
func NewStorageFactory(configPath string) (*StorageFactory, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("storage config file not found: %s", absPath)
	}

	// Read configuration file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage config: %w", err)
	}

	// Parse YAML configuration
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse storage config: %w", err)
	}

	return &StorageFactory{
		config: config,
	}, nil
}

// CreateStorageService creates a storage service for the specified provider
func (f *StorageFactory) CreateStorageService(providerName string) (*StorageService, error) {
	if providerName == "" {
		return nil, fmt.Errorf("provider name cannot be empty")
	}

	providerConfig, exists := f.config[providerName]
	if !exists {
		return nil, fmt.Errorf("storage provider '%s' not found in configuration", providerName)
	}

	// Convert provider config to map with better error handling
	providerMap, ok := providerConfig.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid configuration format for provider '%s'", providerName)
	}

	// Extract and validate provider type
	providerType, ok := providerMap["provider"].(string)
	if !ok || providerType == "" {
		return nil, fmt.Errorf("provider type not specified or invalid for '%s'", providerName)
	}

	// Create storage config based on provider type
	config := StorageConfig{
		Provider: StorageProvider(providerType),
	}

	// Configure based on provider type
	if err := f.configureProvider(&config, providerMap, providerType); err != nil {
		return nil, fmt.Errorf("failed to configure provider '%s': %w", providerName, err)
	}

	// Validate configuration before creating service
	if err := f.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration for provider '%s': %w", providerName, err)
	}

	// Create and return storage service
	return NewStorageService(config)
}

// configureProvider configures the storage config based on provider type
func (f *StorageFactory) configureProvider(config *StorageConfig, providerMap map[string]interface{}, providerType string) error {
	switch StorageProvider(providerType) {
	case StorageLocal:
		if localPath, ok := providerMap["local_path"].(string); ok && localPath != "" {
			config.LocalPath = localPath
		} else {
			config.LocalPath = "./uploads" // Default local path
		}

	case StorageMinIO, StorageS3:
		// Extract S3/MinIO specific configuration
		if accessKey, ok := providerMap["access_key"].(string); ok {
			config.AccessKey = accessKey
		}
		if secretKey, ok := providerMap["secret_key"].(string); ok {
			config.SecretKey = secretKey
		}
		if region, ok := providerMap["region"].(string); ok {
			config.Region = region
		}
		if bucket, ok := providerMap["bucket"].(string); ok {
			config.Bucket = bucket
		}
		if endpoint, ok := providerMap["endpoint"].(string); ok {
			config.Endpoint = endpoint
		}

	default:
		return fmt.Errorf("unsupported storage provider: %s", providerType)
	}

	return nil
}

// validateConfig validates the storage configuration
func (f *StorageFactory) validateConfig(config *StorageConfig) error {
	switch config.Provider {
	case StorageLocal:
		if config.LocalPath == "" {
			return fmt.Errorf("local_path is required for local storage provider")
		}
		// Ensure local path is absolute or relative to current directory
		if !filepath.IsAbs(config.LocalPath) {
			config.LocalPath = filepath.Join(".", config.LocalPath)
		}

	case StorageMinIO, StorageS3:
		if config.AccessKey == "" {
			return fmt.Errorf("access_key is required for %s storage provider", config.Provider)
		}
		if config.SecretKey == "" {
			return fmt.Errorf("secret_key is required for %s storage provider", config.Provider)
		}
		if config.Bucket == "" {
			return fmt.Errorf("bucket is required for %s storage provider", config.Provider)
		}
		if config.Endpoint == "" {
			return fmt.Errorf("endpoint is required for %s storage provider", config.Provider)
		}
		// Set default region if not specified
		if config.Region == "" {
			config.Region = "us-east-1"
		}

	default:
		return fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	return nil
}

// CreateDefaultStorageService creates a storage service using the default provider
func (f *StorageFactory) CreateDefaultStorageService() (*StorageService, error) {
	defaultProvider, ok := f.config["default"].(string)
	if !ok || defaultProvider == "" {
		return nil, fmt.Errorf("default storage provider not specified")
	}

	return f.CreateStorageService(defaultProvider)
}

// GetAvailableProviders returns a list of available storage providers
func (f *StorageFactory) GetAvailableProviders() []string {
	providers := make([]string, 0, len(f.config))
	for provider := range f.config {
		if provider != "default" {
			providers = append(providers, provider)
		}
	}
	return providers
}

// ValidateProvider validates the configuration for a specific storage provider
func (f *StorageFactory) ValidateProvider(providerName string) error {
	if providerName == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	providerConfig, exists := f.config[providerName]
	if !exists {
		return fmt.Errorf("storage provider '%s' not found in configuration", providerName)
	}

	// Convert provider config to map
	providerMap, ok := providerConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid configuration format for provider '%s'", providerName)
	}

	// Extract provider type
	providerType, ok := providerMap["provider"].(string)
	if !ok || providerType == "" {
		return fmt.Errorf("provider type not specified for '%s'", providerName)
	}

	// Create temporary config for validation
	tempConfig := StorageConfig{
		Provider: StorageProvider(providerType),
	}

	// Configure and validate
	if err := f.configureProvider(&tempConfig, providerMap, providerType); err != nil {
		return err
	}

	return f.validateConfig(&tempConfig)
}
