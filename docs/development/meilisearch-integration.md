# MeiliSearch Integration

This document describes how to use the MeiliSearch integration in the Go RESTful application.

## Overview

MeiliSearch is a powerful, fast, and easy-to-use search engine that has been integrated into the application to replace the default search functionality. It provides:

- Fast full-text search
- Typo tolerance
- Faceted search
- Real-time indexing
- RESTful API

## Configuration

### Enable MeiliSearch

In your `config/config.yaml` file, ensure MeiliSearch is enabled:

```yaml
# Meilisearch configuration
meilisearch:
  enable: true  # Set to true to enable MeiliSearch
  host: "localhost"
  port: 7700
  masterKey: "change_me"  # Change this to a secure key
  apiKey: "your-api-key"  # Optional API key for client access
```

### Environment Variables

You can also configure MeiliSearch using environment variables:

```bash
export MEILISEARCH_HOST=localhost
export MEILISEARCH_PORT=7700
export MEILISEARCH_MASTER_KEY=your_secure_master_key
export MEILISEARCH_API_KEY=your_api_key
```

## Running MeiliSearch

### Using Docker

The easiest way to run MeiliSearch is using Docker:

```bash
docker run -p 7700:7700 \
  -e MEILI_MASTER_KEY=your_secure_master_key \
  -v $(pwd)/meili_data:/meili_data \
  getmeili/meilisearch:latest
```

### Using Docker Compose

You can also use the provided `docker-compose.yml` file:

```bash
docker-compose up -d meilisearch
```

## Initializing Indexes

### Using the CLI

The application provides several CLI commands to manage MeiliSearch indexes:

```bash
# Initialize all indexes
go run main.go index init

# Reindex all data
go run main.go index reindex

# Check index status
go run main.go index status
```

### Programmatically

You can also initialize indexes programmatically:

```go
// Get the indexer service from the container
indexerService := container.GetIndexerService()

// Initialize all indexes
ctx := context.Background()
err := indexerService.InitializeIndexes(ctx)

// Reindex all data
err = indexerService.IndexAllData(ctx)

// Get index status
status := indexerService.GetIndexStatus(ctx)
```

## Indexing Data

### Automatic Indexing

The application automatically indexes data when:

- Posts are created, updated, or deleted
- Users are created, updated, or deleted
- Organizations are created, updated, or deleted
- Tags and taxonomies are modified

### Manual Indexing

You can manually index specific entities:

```go
// Index a single post
err := indexerService.IndexPost(ctx, post)

// Update a post in the index
err := indexerService.UpdatePost(ctx, post)

// Remove a post from the index
err := indexerService.RemovePost(ctx, postID)
```

## Search Functionality

### Using the Search Service

The search service provides a unified interface for searching across all indexed entities:

```go
// Search across all indexes
results, err := searchService.Search(ctx, "query", "all", 10, 0)

// Search specific index
results, err := searchService.Search(ctx, "query", "posts", 10, 0)

// Search with filters
results, err := searchService.Search(ctx, "query", "posts", 10, 0, map[string]interface{}{
    "status": "published",
    "language": "en",
})
```

### Search API Endpoints

The application provides REST API endpoints for search:

- `GET /api/v1/search?q=query` - Search across all indexes
- `GET /api/v1/search/posts?q=query` - Search posts only
- `GET /api/v1/search/users?q=query` - Search users only

## Health Monitoring

### Health Check Endpoint

MeiliSearch health is included in the main health check endpoint:

```bash
curl http://localhost:8080/healthz
```

The response will include MeiliSearch status when enabled.

### Index Status

You can check the status of all indexes:

```bash
go run main.go index status
```

This will show:
- Which indexes exist
- Index statistics
- Any errors that occurred

## Performance Considerations

### Indexing Performance

- Large datasets should be indexed in batches
- Use the `BulkIndex` method for bulk operations
- Consider using background jobs for reindexing

### Search Performance

- MeiliSearch automatically optimizes search performance
- Use filters to narrow down search results
- Consider implementing search result caching

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure MeiliSearch is running on the configured host and port
   - Check firewall settings

2. **Authentication Failed**
   - Verify the master key is correct
   - Check if the API key is properly configured

3. **Index Not Found**
   - Run `go run main.go index init` to create missing indexes
   - Check if the index name is correct

4. **Search Returns No Results**
   - Ensure data has been indexed
   - Check if the search query is valid
   - Verify index configuration

### Debugging

Enable debug logging by setting the log level to debug in your configuration:

```yaml
log:
  level: debug
```

### Monitoring

Monitor MeiliSearch performance using:

- Application logs
- MeiliSearch dashboard (if enabled)
- Health check endpoints
- Index status commands

## Migration from Default Search

If you're migrating from the default search implementation:

1. **Enable MeiliSearch** in configuration
2. **Initialize indexes** using `go run main.go index init`
3. **Reindex existing data** using `go run main.go index reindex`
4. **Update your code** to use the new search service interface
5. **Test thoroughly** to ensure search functionality works as expected

## Best Practices

1. **Security**: Use strong master keys and restrict API access
2. **Backup**: Regularly backup your MeiliSearch data
3. **Monitoring**: Monitor search performance and index health
4. **Testing**: Test search functionality thoroughly in development
5. **Documentation**: Keep search configuration and usage documented

## Additional Resources

- [MeiliSearch Documentation](https://docs.meilisearch.com/)
- [MeiliSearch Go Client](https://github.com/meilisearch/meilisearch-go)
- [Application Search Service Interface](internal/application/ports/search_service.go)
- [MeiliSearch Client Implementation](pkg/meilisearch/client.go)
