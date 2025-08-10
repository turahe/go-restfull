package meilisearch

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/go-restfull/config"
)

func TestMeilisearchIntegration(t *testing.T) {
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
		t.Skipf("Skipping integration test: Meilisearch not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test index creation
	indexName := "test_index"
	err = client.CreateIndex(ctx, indexName, "id")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Test if index exists
	// Wait a bit for the index to be fully created
	time.Sleep(2 * time.Second)

	exists, err := client.IndexExists(ctx, indexName)
	if err != nil {
		t.Fatalf("Failed to check if index exists: %v", err)
	}
	if !exists {
		t.Error("Index should exist after creation")
	}

	// Test document addition
	testDoc := map[string]interface{}{
		"id":      "1",
		"title":   "Test Document",
		"content": "This is a test document for Meilisearch integration",
	}

	err = client.AddDocuments(ctx, indexName, []interface{}{testDoc})
	if err != nil {
		t.Fatalf("Failed to add document: %v", err)
	}

	// Wait a bit for indexing
	time.Sleep(1 * time.Second)

	// Test search
	searchOptions := &SearchOptions{
		Query:  "test",
		Limit:  10,
		Offset: 0,
	}

	results, err := client.Search(ctx, indexName, "test", searchOptions)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if results.EstimatedTotalHits == 0 {
		t.Error("Search should return results")
	}

	// Clean up - delete the test index
	err = client.DeleteIndex(ctx, indexName)
	if err != nil {
		t.Logf("Warning: Failed to delete test index: %v", err)
	}
}

func TestMeilisearchHealth(t *testing.T) {
	cfg := &config.Meilisearch{
		Enable:    true,
		Host:      "localhost",
		Port:      7700,
		MasterKey: "your-super-secret-master-key-32-chars-long",
		APIKey:    "your-api-key",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Skipf("Skipping health test: Meilisearch not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test health check
	healthy := client.IsHealthy(ctx)
	if !healthy {
		t.Error("Meilisearch should be healthy")
	}

	// Test getting client
	meiliClient := client.GetClient()
	if meiliClient == nil {
		t.Error("Meilisearch client should not be nil")
	}
}
