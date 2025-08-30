package storage

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// StorageFactory creates storage services for different providers
type StorageFactory struct {
	config map[string]interface{}
}

// NewStorageFactory creates a new storage factory from configuration file
func NewStorageFactory(configPath string) (*StorageFactory, error) {
	// Read configuration file
	data, err := os.ReadFile(configPath)
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
	providerConfig, exists := f.config[providerName]
	if !exists {
		return nil, fmt.Errorf("storage provider '%s' not found in configuration", providerName)
	}

	// Convert provider config to map
	providerMap, ok := providerConfig.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid configuration format for provider '%s'", providerName)
	}

	// Extract provider type
	providerType, ok := providerMap["provider"].(string)
	if !ok {
		return nil, fmt.Errorf("provider type not specified for '%s'", providerName)
	}

	// Create storage config based on provider type
	var config StorageConfig
	config.Provider = StorageProvider(providerType)

	switch config.Provider {
	case StorageLocal:
		if localPath, ok := providerMap["local_path"].(string); ok {
			config.LocalPath = localPath
		} else {
			config.LocalPath = "./uploads" // Default local path
		}

	case StorageMinIO, StorageS3:
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
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	// Create and return storage service
	return NewStorageService(config)
}

// CreateDefaultStorageService creates a storage service using the default provider
func (f *StorageFactory) CreateDefaultStorageService() (*StorageService, error) {
	defaultProvider, ok := f.config["default"].(string)
	if !ok {
		return nil, fmt.Errorf("default storage provider not specified")
	}

	return f.CreateStorageService(defaultProvider)
}

// GetAvailableProviders returns a list of available storage providers
func (f *StorageFactory) GetAvailableProviders() []string {
	var providers []string
	for key := range f.config {
		if key != "default" {
			providers = append(providers, key)
		}
	}
	return providers
}

// ValidateProvider validates if a storage provider configuration is complete
func (f *StorageFactory) ValidateProvider(providerName string) error {
	providerConfig, exists := f.config[providerName]
	if !exists {
		return fmt.Errorf("storage provider '%s' not found", providerName)
	}

	providerMap, ok := providerConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid configuration format for provider '%s'", providerName)
	}

	providerType, ok := providerMap["provider"].(string)
	if !ok {
		return fmt.Errorf("provider type not specified for '%s'", providerName)
	}

	switch StorageProvider(providerType) {
	case StorageLocal:
		// Local storage only needs local_path
		if _, ok := providerMap["local_path"]; !ok {
			return fmt.Errorf("local_path not specified for local storage")
		}

	case StorageMinIO, StorageS3:
		// S3/MinIO needs access_key, secret_key, region, bucket, endpoint
		required := []string{"access_key", "secret_key", "region", "bucket", "endpoint"}
		for _, field := range required {
			if _, ok := providerMap[field]; !ok {
				return fmt.Errorf("required field '%s' not specified for %s storage", field, providerType)
			}
		}

	default:
		return fmt.Errorf("unsupported storage provider: %s", providerType)
	}

	return nil
}
