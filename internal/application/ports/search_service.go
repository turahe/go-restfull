package ports

import (
	"context"

	"github.com/turahe/go-restfull/pkg/meilisearch"
)

// SearchService defines the interface for search operations
type SearchService interface {
	// Index management
	CreateIndex(ctx context.Context, name, primaryKey string) error
	DeleteIndex(ctx context.Context, name string) error
	IndexExists(ctx context.Context, name string) (bool, error)
	UpdateIndexSettings(ctx context.Context, name string, settings interface{}) error

	// Document management
	AddDocuments(ctx context.Context, indexName string, documents []interface{}) error
	UpdateDocuments(ctx context.Context, indexName string, documents []interface{}) error
	DeleteDocuments(ctx context.Context, indexName string, documentIDs []string) error
	GetDocument(ctx context.Context, indexName, documentID string) (interface{}, error)

	// Search operations
	Search(ctx context.Context, indexName, query string, options *meilisearch.SearchOptions) (*meilisearch.SearchResult, error)
	SearchWithFilters(ctx context.Context, indexName, query string, filters string, options *meilisearch.SearchOptions) (*meilisearch.SearchResult, error)
	SearchWithSorting(ctx context.Context, indexName, query string, sort []string, options *meilisearch.SearchOptions) (*meilisearch.SearchResult, error)

	// Bulk operations
	BulkIndex(ctx context.Context, indexName string, documents []interface{}) error
	BulkUpdate(ctx context.Context, indexName string, documents []interface{}) error
	BulkDelete(ctx context.Context, indexName string, documentIDs []string) error

	// Index management
	ReindexAll(ctx context.Context, indexName string, documents []interface{}) error
	ClearIndex(ctx context.Context, indexName string) error
	GetIndexStats(ctx context.Context, indexName string) (interface{}, error)

	// Health check
	IsHealthy(ctx context.Context) bool
}
