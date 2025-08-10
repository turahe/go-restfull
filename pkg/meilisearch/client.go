package meilisearch

import (
	"context"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
	"github.com/turahe/go-restfull/config"
)

// Client represents a Meilisearch client
type Client struct {
	client meilisearch.ServiceManager
	config *config.Meilisearch
}

// SearchResult represents a generic search result
type SearchResult struct {
	Hits               []interface{} `json:"hits"`
	EstimatedTotalHits int64         `json:"estimatedTotalHits"`
	ProcessingTimeMs   int64         `json:"processingTimeMs"`
	Query              string        `json:"query"`
	Limit              int           `json:"limit"`
	Offset             int           `json:"offset"`
	TotalPages         int           `json:"totalPages"`
	CurrentPage        int           `json:"currentPage"`
}

// SearchOptions represents search options
type SearchOptions struct {
	Query                 string   `json:"query"`
	Offset                int      `json:"offset"`
	Limit                 int      `json:"limit"`
	Filters               string   `json:"filters,omitempty"`
	Sort                  []string `json:"sort,omitempty"`
	AttributesToRetrieve  []string `json:"attributesToRetrieve,omitempty"`
	AttributesToHighlight []string `json:"attributesToHighlight,omitempty"`
	AttributesToCrop      []string `json:"attributesToCrop,omitempty"`
	HighlightPreTag       string   `json:"highlightPreTag,omitempty"`
	HighlightPostTag      string   `json:"highlightPostTag,omitempty"`
	CropLength            int      `json:"cropLength,omitempty"`
	ShowMatchesPosition   bool     `json:"showMatchesPosition,omitempty"`
	Facets                []string `json:"facets,omitempty"`
	PlaceholderSearch     bool     `json:"placeholderSearch,omitempty"`
}

// NewClient creates a new Meilisearch client
func NewClient(cfg *config.Meilisearch) (*Client, error) {
	if cfg == nil || !cfg.Enable {
		return nil, fmt.Errorf("meilisearch is not enabled")
	}

	// Create Meilisearch client
	client := meilisearch.New(fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port), meilisearch.WithAPIKey(cfg.MasterKey))

	// Test connection
	_, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meilisearch: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// GetClient returns the underlying Meilisearch client
func (c *Client) GetClient() meilisearch.ServiceManager {
	return c.client
}

// GetIndex returns a Meilisearch index
func (c *Client) GetIndex(name string) meilisearch.IndexManager {
	return c.client.Index(name)
}

// CreateIndex creates a new index with the specified configuration
func (c *Client) CreateIndex(ctx context.Context, name, primaryKey string) error {
	_, err := c.client.CreateIndexWithContext(ctx, &meilisearch.IndexConfig{
		Uid:        name,
		PrimaryKey: primaryKey,
	})
	return err
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(ctx context.Context, name string) error {
	_, err := c.client.DeleteIndexWithContext(ctx, name)
	return err
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(ctx context.Context, name string) (bool, error) {
	// For now, we'll try to get the index and see if it exists
	// This is a simple approach - in production you might want to use a different method
	index := c.client.Index(name)
	_, err := index.GetStats()
	return err == nil, nil
}

// UpdateIndexSettings updates index settings
func (c *Client) UpdateIndexSettings(ctx context.Context, name string, settings interface{}) error {
	index := c.client.Index(name)

	// Convert interface{} to meilisearch.Settings
	meiliSettings, ok := settings.(*meilisearch.Settings)
	if !ok {
		return fmt.Errorf("invalid settings type, expected *meilisearch.Settings")
	}

	_, err := index.UpdateSettings(meiliSettings)
	return err
}

// AddDocuments adds documents to an index
func (c *Client) AddDocuments(ctx context.Context, name string, documents []interface{}) error {
	index := c.client.Index(name)
	_, err := index.AddDocuments(documents, nil)
	return err
}

// UpdateDocuments updates documents in an index
func (c *Client) UpdateDocuments(ctx context.Context, name string, documents []interface{}) error {
	index := c.client.Index(name)
	_, err := index.UpdateDocuments(documents, nil)
	return err
}

// DeleteDocuments deletes documents from an index
func (c *Client) DeleteDocuments(ctx context.Context, name string, documentIDs []string) error {
	index := c.client.Index(name)
	_, err := index.DeleteDocuments(documentIDs)
	return err
}

// Search performs a search on an index
func (c *Client) Search(ctx context.Context, name string, query string, options *SearchOptions) (*SearchResult, error) {
	index := c.client.Index(name)

	// Convert our SearchOptions to Meilisearch SearchRequest
	searchRequest := &meilisearch.SearchRequest{
		Query: query,
	}

	if options != nil {
		searchRequest.Offset = int64(options.Offset)
		searchRequest.Limit = int64(options.Limit)
		searchRequest.Filter = options.Filters
		searchRequest.Sort = options.Sort
		searchRequest.AttributesToRetrieve = options.AttributesToRetrieve
		searchRequest.AttributesToHighlight = options.AttributesToHighlight
		searchRequest.AttributesToCrop = options.AttributesToCrop
		searchRequest.HighlightPreTag = options.HighlightPreTag
		searchRequest.HighlightPostTag = options.HighlightPostTag
		searchRequest.CropLength = int64(options.CropLength)
		searchRequest.ShowMatchesPosition = options.ShowMatchesPosition
		searchRequest.Facets = options.Facets
	}

	// Perform search
	searchResponse, err := index.Search(query, searchRequest)
	if err != nil {
		return nil, err
	}

	// Convert response to our SearchResult format
	result := &SearchResult{
		Query:              query,
		EstimatedTotalHits: searchResponse.EstimatedTotalHits,
		ProcessingTimeMs:   searchResponse.ProcessingTimeMs,
		Limit:              int(searchRequest.Limit),
		Offset:             int(searchRequest.Offset),
	}

	// Convert hits
	if searchResponse.Hits != nil {
		result.Hits = make([]interface{}, len(searchResponse.Hits))
		for i, hit := range searchResponse.Hits {
			result.Hits[i] = hit
		}
	}

	// Calculate pagination
	if searchRequest.Limit > 0 {
		result.TotalPages = int((searchResponse.EstimatedTotalHits + searchRequest.Limit - 1) / searchRequest.Limit)
		result.CurrentPage = int((searchRequest.Offset / searchRequest.Limit) + 1)
	}

	return result, nil
}

// SearchWithFilters performs a search with filters
func (c *Client) SearchWithFilters(ctx context.Context, name string, query string, filters string, options *SearchOptions) (*SearchResult, error) {
	if options == nil {
		options = &SearchOptions{}
	}
	options.Filters = filters
	return c.Search(ctx, name, query, options)
}

// SearchWithSorting performs a search with sorting
func (c *Client) SearchWithSorting(ctx context.Context, name string, query string, sort []string, options *SearchOptions) (*SearchResult, error) {
	if options == nil {
		options = &SearchOptions{}
	}
	options.Sort = sort
	return c.Search(ctx, name, query, options)
}

// GetDocument retrieves a document by ID
func (c *Client) GetDocument(ctx context.Context, name, documentID string) (interface{}, error) {
	index := c.client.Index(name)
	var document interface{}
	err := index.GetDocument(documentID, &meilisearch.DocumentQuery{}, &document)
	if err != nil {
		return nil, err
	}
	return document, nil
}

// GetAllDocuments retrieves all documents from an index
func (c *Client) GetAllDocuments(ctx context.Context, name string) ([]interface{}, error) {
	index := c.client.Index(name)
	var documents meilisearch.DocumentsResult
	err := index.GetDocuments(&meilisearch.DocumentsQuery{}, &documents)
	if err != nil {
		return nil, err
	}

	// Convert to interface{} slice
	result := make([]interface{}, len(documents.Results))
	for i, doc := range documents.Results {
		result[i] = doc
	}

	return result, nil
}

// BulkIndex performs bulk indexing of documents
func (c *Client) BulkIndex(ctx context.Context, name string, documents []interface{}) error {
	index := c.client.Index(name)
	_, err := index.AddDocuments(documents, nil)
	return err
}

// BulkUpdate performs bulk update of documents
func (c *Client) BulkUpdate(ctx context.Context, name string, documents []interface{}) error {
	index := c.client.Index(name)
	_, err := index.UpdateDocuments(documents, nil)
	return err
}

// BulkDelete performs bulk deletion of documents
func (c *Client) BulkDelete(ctx context.Context, name string, documentIDs []string) error {
	index := c.client.Index(name)
	_, err := index.DeleteDocuments(documentIDs)
	return err
}

// ReindexAll reindexes all documents in an index
func (c *Client) ReindexAll(ctx context.Context, name string, documents []interface{}) error {
	index := c.client.Index(name)

	// First, clear the index
	if err := c.ClearIndex(ctx, name); err != nil {
		return fmt.Errorf("failed to clear index: %w", err)
	}

	// Then add all documents
	_, err := index.AddDocuments(documents, nil)
	return err
}

// ClearIndex clears all documents from an index
func (c *Client) ClearIndex(ctx context.Context, name string) error {
	// For now, we'll use a simple approach by deleting the index and recreating it
	// This is more efficient than fetching all documents
	indexName := name

	// Delete the existing index
	if err := c.DeleteIndex(ctx, indexName); err != nil {
		return fmt.Errorf("failed to delete index for clearing: %w", err)
	}

	// Recreate the index with the same name
	if err := c.CreateIndex(ctx, indexName, "id"); err != nil {
		return fmt.Errorf("failed to recreate index after clearing: %w", err)
	}

	return nil
}

// IsHealthy checks if the Meilisearch service is healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	_, err := c.client.Health()
	return err == nil
}

// GetIndexStats gets index statistics
func (c *Client) GetIndexStats(ctx context.Context, name string) (interface{}, error) {
	index := c.client.Index(name)
	stats, err := index.GetStats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	c.client.Close()
	return nil
}
