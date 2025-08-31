package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/turahe/go-restfull/pkg/storage"
)

func main() {
	// Get the project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Go up to project root if we're in scripts directory
	if filepath.Base(projectRoot) == "scripts" {
		projectRoot = filepath.Dir(projectRoot)
	}

	// Path to storage config
	configPath := filepath.Join(projectRoot, "config", "storage.yaml")

	fmt.Printf("Testing MinIO connection with config: %s\n", configPath)
	fmt.Printf("Project root: %s\n", projectRoot)

	// Create storage loader
	loader, err := storage.NewStorageLoader(configPath)
	if err != nil {
		log.Fatalf("Failed to create storage loader: %v", err)
	}

	// Load MinIO storage service
	minioStorage, err := loader.LoadStorage("minio")
	if err != nil {
		log.Fatalf("Failed to load MinIO storage: %v", err)
	}

	fmt.Printf("MinIO storage service created successfully\n")
	fmt.Printf("Provider: %s\n", minioStorage.GetProvider())
	fmt.Printf("Config: %+v\n", minioStorage.GetStorageInfo())

	// Test the connection
	fmt.Println("\nTesting MinIO connection...")
	if err := minioStorage.TestConnection(); err != nil {
		log.Fatalf("MinIO connection test failed: %v", err)
	}

	fmt.Println("✅ MinIO connection test successful!")

	// Test bucket access
	fmt.Println("\nTesting bucket access...")
	if err := minioStorage.EnsureBucketExists(); err != nil {
		log.Fatalf("Bucket access test failed: %v", err)
	}

	fmt.Println("✅ Bucket access test successful!")
	fmt.Println("\n🎉 All MinIO tests passed! The storage service is properly configured.")
}
