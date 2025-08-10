package services

import (
	"context"
	"fmt"
	"log"
	"time"

	meili "github.com/meilisearch/meilisearch-go"
	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/application/ports"
	meiliPkg "github.com/turahe/go-restfull/pkg/meilisearch"
)

// searchService implements the SearchService interface
type searchService struct {
	client *meiliPkg.Client
	config *config.Meilisearch
}

// NewSearchService creates a new search service instance
func NewSearchService(cfg *config.Meilisearch) (ports.SearchService, error) {
	client, err := meiliPkg.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create meilisearch client: %w", err)
	}

	return &searchService{
		client: client,
		config: cfg,
	}, nil
}

// CreateIndex creates a new index with the specified configuration
func (s *searchService) CreateIndex(ctx context.Context, name, primaryKey string) error {
	return s.client.CreateIndex(ctx, name, primaryKey)
}

// DeleteIndex deletes an index
func (s *searchService) DeleteIndex(ctx context.Context, name string) error {
	return s.client.DeleteIndex(ctx, name)
}

// IndexExists checks if an index exists
func (s *searchService) IndexExists(ctx context.Context, name string) (bool, error) {
	return s.client.IndexExists(ctx, name)
}

// UpdateIndexSettings updates the settings for an index
func (s *searchService) UpdateIndexSettings(ctx context.Context, name string, settings interface{}) error {
	// Convert interface{} to meilisearch.Settings
	meiliSettings, ok := settings.(*meili.Settings)
	if !ok {
		return fmt.Errorf("invalid settings type, expected *meilisearch.Settings")
	}
	return s.client.UpdateIndexSettings(ctx, name, meiliSettings)
}

// AddDocuments adds documents to an index
func (s *searchService) AddDocuments(ctx context.Context, indexName string, documents []interface{}) error {
	return s.client.AddDocuments(ctx, indexName, documents)
}

// UpdateDocuments updates documents in an index
func (s *searchService) UpdateDocuments(ctx context.Context, indexName string, documents []interface{}) error {
	return s.client.UpdateDocuments(ctx, indexName, documents)
}

// DeleteDocuments deletes documents from an index
func (s *searchService) DeleteDocuments(ctx context.Context, indexName string, documentIDs []string) error {
	return s.client.DeleteDocuments(ctx, indexName, documentIDs)
}

// GetDocument retrieves a document by ID
func (s *searchService) GetDocument(ctx context.Context, indexName, documentID string) (interface{}, error) {
	return s.client.GetDocument(ctx, indexName, documentID)
}

// Search performs a search query on an index
func (s *searchService) Search(ctx context.Context, indexName, query string, options *meiliPkg.SearchOptions) (*meiliPkg.SearchResult, error) {
	return s.client.Search(ctx, indexName, query, options)
}

// SearchWithFilters performs a search with filters
func (s *searchService) SearchWithFilters(ctx context.Context, indexName, query string, filters string, options *meiliPkg.SearchOptions) (*meiliPkg.SearchResult, error) {
	if options == nil {
		options = &meiliPkg.SearchOptions{}
	}
	options.Filters = filters
	return s.client.Search(ctx, indexName, query, options)
}

// SearchWithSorting performs a search with sorting
func (s *searchService) SearchWithSorting(ctx context.Context, indexName, query string, sort []string, options *meiliPkg.SearchOptions) (*meiliPkg.SearchResult, error) {
	if options == nil {
		options = &meiliPkg.SearchOptions{}
	}
	options.Sort = sort
	return s.client.Search(ctx, indexName, query, options)
}

// BulkIndex performs bulk indexing of documents
func (s *searchService) BulkIndex(ctx context.Context, indexName string, documents []interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	// Split documents into batches to avoid memory issues
	batchSize := 1000
	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		if err := s.client.AddDocuments(ctx, indexName, batch); err != nil {
			return fmt.Errorf("failed to index batch %d-%d: %w", i, end-1, err)
		}

		// Small delay between batches to avoid overwhelming the server
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// BulkUpdate performs bulk update of documents
func (s *searchService) BulkUpdate(ctx context.Context, indexName string, documents []interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	// Split documents into batches
	batchSize := 1000
	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		if err := s.client.UpdateDocuments(ctx, indexName, batch); err != nil {
			return fmt.Errorf("failed to update batch %d-%d: %w", i, end-1, err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// BulkDelete performs bulk deletion of documents
func (s *searchService) BulkDelete(ctx context.Context, indexName string, documentIDs []string) error {
	if len(documentIDs) == 0 {
		return nil
	}

	// Split IDs into batches
	batchSize := 1000
	for i := 0; i < len(documentIDs); i += batchSize {
		end := i + batchSize
		if end > len(documentIDs) {
			end = len(documentIDs)
		}

		batch := documentIDs[i:end]
		if err := s.client.DeleteDocuments(ctx, indexName, batch); err != nil {
			return fmt.Errorf("failed to delete batch %d-%d: %w", i, end-1, err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// ReindexAll clears the index and reindexes all documents
func (s *searchService) ReindexAll(ctx context.Context, indexName string, documents []interface{}) error {
	// Clear the index first
	if err := s.ClearIndex(ctx, indexName); err != nil {
		return fmt.Errorf("failed to clear index: %w", err)
	}

	// Reindex all documents
	return s.BulkIndex(ctx, indexName, documents)
}

// ClearIndex clears all documents from an index
func (s *searchService) ClearIndex(ctx context.Context, indexName string) error {
	// Get all documents to get their IDs
	documents, err := s.client.GetAllDocuments(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to get documents for clearing: %w", err)
	}

	if len(documents) == 0 {
		return nil
	}

	// Extract IDs from documents
	var documentIDs []string
	for _, doc := range documents {
		if docMap, ok := doc.(map[string]interface{}); ok {
			if id, exists := docMap["id"]; exists {
				if idStr, ok := id.(string); ok {
					documentIDs = append(documentIDs, idStr)
				}
			}
		}
	}

	// Delete all documents
	return s.BulkDelete(ctx, indexName, documentIDs)
}

// GetIndexStats gets statistics for an index
func (s *searchService) GetIndexStats(ctx context.Context, indexName string) (interface{}, error) {
	return s.client.GetIndexStats(ctx, indexName)
}

// IsHealthy checks if the search service is healthy
func (s *searchService) IsHealthy(ctx context.Context) bool {
	// Try to get a simple health check
	_, err := s.client.GetClient().Health()
	return err == nil
}

// InitializeIndexes initializes all configured indexes
func (s *searchService) InitializeIndexes(ctx context.Context) error {
	for _, indexConfig := range s.config.Indexes {
		// Check if index exists
		exists, err := s.IndexExists(ctx, indexConfig.UID)
		if err != nil {
			log.Printf("Failed to check if index %s exists: %v", indexConfig.UID, err)
			continue
		}

		if !exists {
			// Create index
			if err := s.CreateIndex(ctx, indexConfig.UID, indexConfig.PrimaryKey); err != nil {
				log.Printf("Failed to create index %s: %v", indexConfig.UID, err)
				continue
			}
			log.Printf("Created index: %s", indexConfig.UID)
		}

		// Update index settings if provided
		if indexConfig.Settings.SearchableAttributes != nil ||
			indexConfig.Settings.FilterableAttributes != nil ||
			indexConfig.Settings.SortableAttributes != nil {

			settings := &meili.Settings{}

			if indexConfig.Settings.SearchableAttributes != nil {
				settings.SearchableAttributes = indexConfig.Settings.SearchableAttributes
			}
			if indexConfig.Settings.FilterableAttributes != nil {
				settings.FilterableAttributes = indexConfig.Settings.FilterableAttributes
			}
			if indexConfig.Settings.SortableAttributes != nil {
				settings.SortableAttributes = indexConfig.Settings.SortableAttributes
			}
			if indexConfig.Settings.RankingRules != nil {
				settings.RankingRules = indexConfig.Settings.RankingRules
			}
			if indexConfig.Settings.StopWords != nil {
				settings.StopWords = indexConfig.Settings.StopWords
			}
			if indexConfig.Settings.Synonyms != nil {
				settings.Synonyms = indexConfig.Settings.Synonyms
			}

			if err := s.UpdateIndexSettings(ctx, indexConfig.UID, settings); err != nil {
				log.Printf("Failed to update settings for index %s: %v", indexConfig.UID, err)
			} else {
				log.Printf("Updated settings for index: %s", indexConfig.UID)
			}
		}
	}

	return nil
}
