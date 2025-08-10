package meilisearch

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/go-restfull/config"
)

func TestNewClient(t *testing.T) {
	// Test with disabled Meilisearch
	cfg := &config.Meilisearch{
		Enable: false,
	}

	client, err := NewClient(cfg)
	if err == nil {
		t.Error("Expected error when Meilisearch is disabled")
	}
	if client != nil {
		t.Error("Expected nil client when Meilisearch is disabled")
	}
}

func TestClientConnection(t *testing.T) {
	// Skip if Meilisearch is not running
	cfg := &config.Meilisearch{
		Enable:    true,
		Host:      "localhost",
		Port:      7700,
		MasterKey: "your-super-secret-master-key-32-chars-long",
		APIKey:    "your-api-key",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Skipping test: Meilisearch not available: %v", err)
	}

	// Test basic operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test health check
	healthy := client.IsHealthy(ctx)
	if !healthy {
		t.Error("Expected Meilisearch to be healthy")
	}

	// Test getting client
	meiliClient := client.GetClient()
	if meiliClient == nil {
		t.Error("Expected non-nil Meilisearch client")
	}
}

func TestSearchResult(t *testing.T) {
	// Test SearchResult struct
	result := &SearchResult{
		Hits:               []interface{}{"test"},
		EstimatedTotalHits: 1,
		ProcessingTimeMs:   10,
		Query:              "test",
		Limit:              10,
		Offset:             0,
		TotalPages:         1,
		CurrentPage:        1,
	}

	if result.EstimatedTotalHits != 1 {
		t.Error("Expected EstimatedTotalHits to be 1")
	}

	if result.Query != "test" {
		t.Error("Expected Query to be 'test'")
	}
}
