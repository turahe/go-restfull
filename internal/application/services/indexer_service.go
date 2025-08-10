package services

import (
	"context"
	"fmt"
	"log"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
)

// IndexerService handles indexing of entities into MeiliSearch
type IndexerService struct {
	searchService ports.SearchService
	postRepo      repositories.PostRepository
	userRepo      repositories.UserRepository
	orgRepo       repositories.OrganizationRepository
	tagRepo       repositories.TagRepository
	taxonomyRepo  repositories.TaxonomyRepository
	mediaRepo     repositories.MediaRepository
	menuRepo      repositories.MenuRepository
	roleRepo      repositories.RoleRepository
	contentRepo   repositories.ContentRepository
	addressRepo   repositories.AddressRepository
	commentRepo   repositories.CommentRepository
}

// NewIndexerService creates a new indexer service
func NewIndexerService(
	searchService ports.SearchService,
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	orgRepo repositories.OrganizationRepository,
	tagRepo repositories.TagRepository,
	taxonomyRepo repositories.TaxonomyRepository,
	mediaRepo repositories.MediaRepository,
	menuRepo repositories.MenuRepository,
	roleRepo repositories.RoleRepository,
	contentRepo repositories.ContentRepository,
	addressRepo repositories.AddressRepository,
	commentRepo repositories.CommentRepository,
) *IndexerService {
	return &IndexerService{
		searchService: searchService,
		postRepo:      postRepo,
		userRepo:      userRepo,
		orgRepo:       orgRepo,
		tagRepo:       tagRepo,
		taxonomyRepo:  taxonomyRepo,
		mediaRepo:     mediaRepo,
		menuRepo:      menuRepo,
		roleRepo:      roleRepo,
		contentRepo:   contentRepo,
		addressRepo:   addressRepo,
		commentRepo:   commentRepo,
	}
}

// InitializeIndexes creates and configures all search indexes
func (s *IndexerService) InitializeIndexes(ctx context.Context) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	log.Println("Initializing MeiliSearch indexes...")

	// Create indexes
	indexes := []struct {
		name       string
		primaryKey string
	}{
		{"posts", "id"},
		{"users", "id"},
		{"organizations", "id"},
		{"tags", "id"},
		{"taxonomies", "id"},
		{"media", "id"},
		{"menus", "id"},
		{"roles", "id"},
		{"content", "id"},
		{"addresses", "id"},
		{"comments", "id"},
	}

	for _, idx := range indexes {
		if err := s.createIndexIfNotExists(ctx, idx.name, idx.primaryKey); err != nil {
			log.Printf("Warning: Failed to create index %s: %v", idx.name, err)
		}
	}

	return nil
}

// createIndexIfNotExists creates an index if it doesn't exist
func (s *IndexerService) createIndexIfNotExists(ctx context.Context, name, primaryKey string) error {
	exists, err := s.searchService.IndexExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check if index %s exists: %w", name, err)
	}

	if !exists {
		if err := s.searchService.CreateIndex(ctx, name, primaryKey); err != nil {
			return fmt.Errorf("failed to create index %s: %w", name, err)
		}
		log.Printf("Created index: %s", name)
	} else {
		log.Printf("Index already exists: %s", name)
	}

	return nil
}

// IndexAllData indexes all data from the database into MeiliSearch
func (s *IndexerService) IndexAllData(ctx context.Context) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	log.Println("Starting full reindex of all data...")

	// Index posts
	if err := s.indexPosts(ctx); err != nil {
		log.Printf("Warning: Failed to index posts: %v", err)
	}

	// Index users
	if err := s.indexUsers(ctx); err != nil {
		log.Printf("Warning: Failed to index users: %v", err)
	}

	// Index organizations
	if err := s.indexOrganizations(ctx); err != nil {
		log.Printf("Warning: Failed to index organizations: %v", err)
	}

	// Index tags
	if err := s.indexTags(ctx); err != nil {
		log.Printf("Warning: Failed to index tags: %v", err)
	}

	// Index taxonomies
	if err := s.indexTaxonomies(ctx); err != nil {
		log.Printf("Warning: Failed to index taxonomies: %v", err)
	}

	// Index content
	if err := s.indexContent(ctx); err != nil {
		log.Printf("Warning: Failed to index content: %v", err)
	}

	log.Println("Full reindex completed")
	return nil
}

// indexPosts indexes all posts
func (s *IndexerService) indexPosts(ctx context.Context) error {
	posts, err := s.postRepo.GetAll(ctx, 10000, 0) // Get all posts
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}

	if len(posts) == 0 {
		log.Println("No posts to index")
		return nil
	}

	// Convert posts to searchable documents
	var documents []interface{}
	for _, post := range posts {
		doc := map[string]interface{}{
			"id":           post.ID,
			"title":        post.Title,
			"slug":         post.Slug,
			"subtitle":     post.Subtitle,
			"description":  post.Description,
			"language":     post.Language,
			"layout":       post.Layout,
			"is_sticky":    post.IsSticky,
			"published_at": post.PublishedAt,
			"created_at":   post.CreatedAt,
			"updated_at":   post.UpdatedAt,
		}
		documents = append(documents, doc)
	}

	if err := s.searchService.ReindexAll(ctx, "posts", documents); err != nil {
		return fmt.Errorf("failed to index posts: %w", err)
	}

	log.Printf("Indexed %d posts", len(posts))
	return nil
}

// indexUsers indexes all users
func (s *IndexerService) indexUsers(ctx context.Context) error {
	// Note: This is a simplified implementation
	// In a real scenario, you'd want to implement GetAll method for users
	log.Println("User indexing not implemented yet")
	return nil
}

// indexOrganizations indexes all organizations
func (s *IndexerService) indexOrganizations(ctx context.Context) error {
	// Note: This is a simplified implementation
	// In a real scenario, you'd want to implement GetAll method for organizations
	log.Println("Organization indexing not implemented yet")
	return nil
}

// indexTags indexes all tags
func (s *IndexerService) indexTags(ctx context.Context) error {
	// Note: This is a simplified implementation
	// In a real scenario, you'd want to implement GetAll method for tags
	log.Println("Tag indexing not implemented yet")
	return nil
}

// indexTaxonomies indexes all taxonomies
func (s *IndexerService) indexTaxonomies(ctx context.Context) error {
	// Note: This is a simplified implementation
	// In a real scenario, you'd want to implement GetAll method for taxonomies
	log.Println("Taxonomy indexing not implemented yet")
	return nil
}

// indexContent indexes all content
func (s *IndexerService) indexContent(ctx context.Context) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	// Get all content from repository
	contents, err := s.contentRepo.GetAll(ctx, 1000, 0) // Adjust limit as needed
	if err != nil {
		return fmt.Errorf("failed to get content for indexing: %w", err)
	}

	if len(contents) == 0 {
		log.Println("No content found for indexing")
		return nil
	}

	// Convert to documents for indexing
	var docs []interface{}
	for _, content := range contents {
		doc := map[string]interface{}{
			"id":           content.ID,
			"model_type":   content.ModelType,
			"model_id":     content.ModelID,
			"content_raw":  content.ContentRaw,
			"content_html": content.ContentHTML,
			"created_by":   content.CreatedBy,
			"updated_by":   content.UpdatedBy,
			"created_at":   content.CreatedAt,
			"updated_at":   content.UpdatedAt,
		}
		docs = append(docs, doc)
	}

	// Index all content
	err = s.searchService.AddDocuments(ctx, "content", docs)
	if err != nil {
		return fmt.Errorf("failed to index content: %w", err)
	}

	log.Printf("Successfully indexed %d content items", len(contents))
	return nil
}

// IndexPost indexes a single post
func (s *IndexerService) IndexPost(ctx context.Context, post *entities.Post) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	doc := map[string]interface{}{
		"id":           post.ID,
		"title":        post.Title,
		"slug":         post.Slug,
		"subtitle":     post.Subtitle,
		"description":  post.Description,
		"language":     post.Language,
		"layout":       post.Layout,
		"is_sticky":    post.IsSticky,
		"published_at": post.PublishedAt,
		"created_at":   post.CreatedAt,
		"updated_at":   post.UpdatedAt,
	}

	return s.searchService.AddDocuments(ctx, "posts", []interface{}{doc})
}

// RemovePost removes a post from the search index
func (s *IndexerService) RemovePost(ctx context.Context, postID string) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	return s.searchService.DeleteDocuments(ctx, "posts", []string{postID})
}

// UpdatePost updates a post in the search index
func (s *IndexerService) UpdatePost(ctx context.Context, post *entities.Post) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	doc := map[string]interface{}{
		"id":           post.ID,
		"title":        post.Title,
		"slug":         post.Slug,
		"subtitle":     post.Subtitle,
		"description":  post.Description,
		"language":     post.Language,
		"layout":       post.Layout,
		"is_sticky":    post.IsSticky,
		"published_at": post.PublishedAt,
		"created_at":   post.CreatedAt,
		"updated_at":   post.UpdatedAt,
	}

	return s.searchService.UpdateDocuments(ctx, "posts", []interface{}{doc})
}

// IndexContent indexes a single content item
func (s *IndexerService) IndexContent(ctx context.Context, content *entities.Content) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	doc := map[string]interface{}{
		"id":           content.ID,
		"model_type":   content.ModelType,
		"model_id":     content.ModelID,
		"content_raw":  content.ContentRaw,
		"content_html": content.ContentHTML,
		"created_by":   content.CreatedBy,
		"updated_by":   content.UpdatedBy,
		"created_at":   content.CreatedAt,
		"updated_at":   content.UpdatedAt,
	}

	return s.searchService.AddDocuments(ctx, "content", []interface{}{doc})
}

// RemoveContent removes a content item from the search index
func (s *IndexerService) RemoveContent(ctx context.Context, contentID string) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	return s.searchService.DeleteDocuments(ctx, "content", []string{contentID})
}

// UpdateContent updates a content item in the search index
func (s *IndexerService) UpdateContent(ctx context.Context, content *entities.Content) error {
	if s.searchService == nil {
		return fmt.Errorf("search service not available")
	}

	doc := map[string]interface{}{
		"id":           content.ID,
		"model_type":   content.ModelType,
		"model_id":     content.ModelID,
		"content_raw":  content.ContentRaw,
		"content_html": content.ContentHTML,
		"created_by":   content.CreatedBy,
		"updated_by":   content.UpdatedBy,
		"created_at":   content.CreatedAt,
		"updated_at":   content.UpdatedAt,
	}

	return s.searchService.UpdateDocuments(ctx, "content", []interface{}{doc})
}

// GetIndexStatus returns the status of all indexes
func (s *IndexerService) GetIndexStatus(ctx context.Context) map[string]interface{} {
	status := map[string]interface{}{
		"meilisearch_available": s.searchService != nil,
		"indexes":               map[string]interface{}{},
	}

	if s.searchService == nil {
		return status
	}

	// Check each index
	indexes := []string{"posts", "users", "organizations", "tags", "taxonomies", "media", "menus", "roles", "content", "addresses", "comments"}

	for _, indexName := range indexes {
		exists, err := s.searchService.IndexExists(ctx, indexName)
		indexStatus := map[string]interface{}{
			"exists": exists,
		}

		if err != nil {
			indexStatus["error"] = err.Error()
		}

		if exists {
			// Try to get stats
			if stats, err := s.searchService.GetIndexStats(ctx, indexName); err == nil {
				indexStatus["stats"] = stats
			}
		}

		status["indexes"].(map[string]interface{})[indexName] = indexStatus
	}

	return status
}
