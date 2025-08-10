package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/pkg/meilisearch"
)

// HybridSearchService provides search functionality using Meilisearch when available,
// falling back to SQL search when not
type HybridSearchService struct {
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

// NewHybridSearchService creates a new hybrid search service
func NewHybridSearchService(
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
) *HybridSearchService {
	return &HybridSearchService{
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

// IsMeilisearchAvailable checks if Meilisearch is available and healthy
func (s *HybridSearchService) IsMeilisearchAvailable() bool {
	if s.searchService == nil {
		return false
	}

	// Check if Meilisearch is healthy
	ctx := context.Background()
	return s.searchService.IsHealthy(ctx)
}

// SearchPosts searches posts using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for posts search: %s", query)
		return s.searchPostsWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for posts: %s", query)
	return s.postRepo.Search(ctx, query, limit, offset)
}

// SearchUsers searches users using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for users search: %s", query)
		return s.searchUsersWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for users: %s", query)
	// Convert to SearchUsersQuery format
	searchQuery := queries.SearchUsersQuery{
		Query:    query,
		Page:     (offset / limit) + 1,
		PageSize: limit,
	}

	result, err := s.userRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	// Convert from UserAggregate to User entities
	var users []*entities.User
	for _, userAgg := range result.Items {
		// Convert UserAggregate to User entity
		user := &entities.User{
			ID:              userAgg.ID,
			UserName:        userAgg.UserName,
			Email:           userAgg.Email.String(),
			Phone:           userAgg.Phone.String(),
			EmailVerifiedAt: userAgg.EmailVerifiedAt,
			PhoneVerifiedAt: userAgg.PhoneVerifiedAt,
			CreatedAt:       userAgg.CreatedAt,
			UpdatedAt:       userAgg.UpdatedAt,
			DeletedAt:       userAgg.DeletedAt,
		}
		users = append(users, user)
	}

	return users, nil
}

// SearchOrganizations searches organizations using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchOrganizations(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for organizations search: %s", query)
		return s.searchOrganizationsWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for organizations: %s", query)
	return s.orgRepo.Search(ctx, query, limit, offset)
}

// SearchTags searches tags using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchTags(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for tags search: %s", query)
		return s.searchTagsWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for tags: %s", query)
	return s.tagRepo.Search(ctx, query, limit, offset)
}

// SearchTaxonomies searches taxonomies using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchTaxonomies(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for taxonomies search: %s", query)
		return s.searchTaxonomiesWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for taxonomies: %s", query)
	return s.taxonomyRepo.Search(ctx, query, limit, offset)
}

// SearchMedia searches media using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchMedia(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for media search: %s", query)
		return s.searchMediaWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for media: %s", query)
	return s.mediaRepo.Search(ctx, query, limit, offset)
}

// SearchMenus searches menus using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchMenus(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for menus search: %s", query)
		return s.searchMenusWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for menus: %s", query)
	return s.menuRepo.Search(ctx, query, limit, offset)
}

// SearchRoles searches roles using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchRoles(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for roles search: %s", query)
		return s.searchRolesWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for roles: %s", query)
	return s.roleRepo.Search(ctx, query, limit, offset)
}

// SearchContent searches content using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchContent(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for content search: %s", query)
		return s.searchContentWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for content: %s", query)
	return s.contentRepo.Search(ctx, query, limit, offset)
}

// SearchAddresses searches addresses using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchAddresses(ctx context.Context, query string, limit, offset int) ([]*entities.Address, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for addresses search: %s", query)
		return s.searchAddressesWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for addresses: %s", query)
	// Use the general search method
	return s.addressRepo.Search(ctx, query, limit, offset)
}

// SearchComments searches comments using Meilisearch if available, otherwise falls back to SQL
func (s *HybridSearchService) SearchComments(ctx context.Context, query string, limit, offset int) ([]*entities.Comment, error) {
	if s.IsMeilisearchAvailable() {
		log.Printf("Using Meilisearch for comments search: %s", query)
		return s.searchCommentsWithMeilisearch(ctx, query, limit, offset)
	}

	log.Printf("Meilisearch not available, falling back to SQL search for comments: %s", query)
	// Use the SQL search method
	return s.commentRepo.Search(ctx, query, limit, offset)
}

// GetSearchEngineInfo returns information about the current search engine
func (s *HybridSearchService) GetSearchEngineInfo() map[string]interface{} {
	info := map[string]interface{}{
		"primary_engine":        "sql",
		"meilisearch_available": false,
		"meilisearch_healthy":   false,
	}

	if s.searchService != nil {
		info["meilisearch_available"] = true
		info["meilisearch_healthy"] = s.searchService.IsHealthy(context.Background())

		if info["meilisearch_healthy"].(bool) {
			info["primary_engine"] = "meilisearch"
		}
	}

	return info
}

// Meilisearch search implementations
func (s *HybridSearchService) searchPostsWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "posts", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Post entities
	var posts []*entities.Post
	for _, hit := range result.Hits {
		if postMap, ok := hit.(map[string]interface{}); ok {
			// This is a simplified conversion - in a real implementation,
			// you'd want to properly map all fields
			post := &entities.Post{}
			if id, exists := postMap["id"]; exists {
				if _, ok := id.(string); ok {
					// Convert string ID to UUID if needed
					// post.ID = uuid.MustParse(idStr)
				}
			}
			if title, exists := postMap["title"]; exists {
				if titleStr, ok := title.(string); ok {
					post.Title = titleStr
				}
			}
			// Add more field mappings as needed
			posts = append(posts, post)
		}
	}

	return posts, nil
}

func (s *HybridSearchService) searchUsersWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "users", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to User entities
	var users []*entities.User
	for _, hit := range result.Hits {
		if userMap, ok := hit.(map[string]interface{}); ok {
			user := &entities.User{}
			if username, exists := userMap["username"]; exists {
				if usernameStr, ok := username.(string); ok {
					user.UserName = usernameStr
				}
			}
			if email, exists := userMap["email"]; exists {
				if emailStr, ok := email.(string); ok {
					user.Email = emailStr
				}
			}
			// Add more field mappings as needed
			users = append(users, user)
		}
	}

	return users, nil
}

func (s *HybridSearchService) searchOrganizationsWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "organizations", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Organization entities
	var organizations []*entities.Organization
	for _, hit := range result.Hits {
		if orgMap, ok := hit.(map[string]interface{}); ok {
			org := &entities.Organization{}
			if name, exists := orgMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					org.Name = nameStr
				}
			}
			// Add more field mappings as needed
			organizations = append(organizations, org)
		}
	}

	return organizations, nil
}

func (s *HybridSearchService) searchTagsWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "tags", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Tag entities
	var tags []*entities.Tag
	for _, hit := range result.Hits {
		if tagMap, ok := hit.(map[string]interface{}); ok {
			tag := &entities.Tag{}
			if name, exists := tagMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					tag.Name = nameStr
				}
			}
			// Add more field mappings as needed
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func (s *HybridSearchService) searchTaxonomiesWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "taxonomies", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Taxonomy entities
	var taxonomies []*entities.Taxonomy
	for _, hit := range result.Hits {
		if taxMap, ok := hit.(map[string]interface{}); ok {
			tax := &entities.Taxonomy{}
			if name, exists := taxMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					tax.Name = nameStr
				}
			}
			// Add more field mappings as needed
			taxonomies = append(taxonomies, tax)
		}
	}

	return taxonomies, nil
}

func (s *HybridSearchService) searchMediaWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "media", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Media entities
	var media []*entities.Media
	for _, hit := range result.Hits {
		if mediaMap, ok := hit.(map[string]interface{}); ok {
			med := &entities.Media{}
			if filename, exists := mediaMap["filename"]; exists {
				if filenameStr, ok := filename.(string); ok {
					med.FileName = filenameStr
				}
			}
			// Add more field mappings as needed
			media = append(media, med)
		}
	}

	return media, nil
}

func (s *HybridSearchService) searchMenusWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "menus", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Menu entities
	var menus []*entities.Menu
	for _, hit := range result.Hits {
		if menuMap, ok := hit.(map[string]interface{}); ok {
			menu := &entities.Menu{}
			if name, exists := menuMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					menu.Name = nameStr
				}
			}
			// Add more field mappings as needed
			menus = append(menus, menu)
		}
	}

	return menus, nil
}

func (s *HybridSearchService) searchRolesWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "roles", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Role entities
	var roles []*entities.Role
	for _, hit := range result.Hits {
		if roleMap, ok := hit.(map[string]interface{}); ok {
			role := &entities.Role{}
			if name, exists := roleMap["name"]; exists {
				if nameStr, ok := name.(string); ok {
					role.Name = nameStr
				}
			}
			// Add more field mappings as needed
			roles = append(roles, role)
		}
	}

	return roles, nil
}

func (s *HybridSearchService) searchContentWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	options := &meilisearch.SearchOptions{
		Limit:                 limit,
		Offset:                offset,
		AttributesToRetrieve:  []string{"id", "model_type", "model_id", "content_raw", "content_html", "created_by", "updated_by", "created_at", "updated_at"},
		AttributesToHighlight: []string{"content_raw", "content_html"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
	}

	result, err := s.searchService.Search(ctx, "content", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Content entities
	var contents []*entities.Content
	for _, hit := range result.Hits {
		if contentMap, ok := hit.(map[string]interface{}); ok {
			content := &entities.Content{}

			// Map ID
			if id, exists := contentMap["id"]; exists {
				if idStr, ok := id.(string); ok {
					if parsedID, err := uuid.Parse(idStr); err == nil {
						content.ID = parsedID
					}
				}
			}

			// Map ModelType
			if modelType, exists := contentMap["model_type"]; exists {
				if modelTypeStr, ok := modelType.(string); ok {
					content.ModelType = modelTypeStr
				}
			}

			// Map ModelID
			if modelID, exists := contentMap["model_id"]; exists {
				if modelIDStr, ok := modelID.(string); ok {
					if parsedModelID, err := uuid.Parse(modelIDStr); err == nil {
						content.ModelID = parsedModelID
					}
				}
			}

			// Map ContentRaw
			if contentRaw, exists := contentMap["content_raw"]; exists {
				if contentRawStr, ok := contentRaw.(string); ok {
					content.ContentRaw = contentRawStr
				}
			}

			// Map ContentHTML
			if contentHTML, exists := contentMap["content_html"]; exists {
				if contentHTMLStr, ok := contentHTML.(string); ok {
					content.ContentHTML = contentHTMLStr
				}
			}

			// Map CreatedBy
			if createdBy, exists := contentMap["created_by"]; exists {
				if createdByStr, ok := createdBy.(string); ok {
					if parsedCreatedBy, err := uuid.Parse(createdByStr); err == nil {
						content.CreatedBy = parsedCreatedBy
					}
				}
			}

			// Map UpdatedBy
			if updatedBy, exists := contentMap["updated_by"]; exists {
				if updatedByStr, ok := updatedBy.(string); ok {
					if parsedUpdatedBy, err := uuid.Parse(updatedByStr); err == nil {
						content.UpdatedBy = parsedUpdatedBy
					}
				}
			}

			// Map CreatedAt
			if createdAt, exists := contentMap["created_at"]; exists {
				if createdAtStr, ok := createdAt.(string); ok {
					if parsedTime, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
						content.CreatedAt = parsedTime
					}
				}
			}

			// Map UpdatedAt
			if updatedAt, exists := contentMap["updated_at"]; exists {
				if updatedAtStr, ok := updatedAt.(string); ok {
					if parsedTime, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
						content.UpdatedAt = parsedTime
					}
				}
			}

			contents = append(contents, content)
		}
	}

	return contents, nil
}

func (s *HybridSearchService) searchAddressesWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Address, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "addresses", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Address entities
	var addresses []*entities.Address
	for _, hit := range result.Hits {
		if addrMap, ok := hit.(map[string]interface{}); ok {
			addr := &entities.Address{}
			if city, exists := addrMap["city"]; exists {
				if cityStr, ok := city.(string); ok {
					addr.City = cityStr
				}
			}
			// Add more field mappings as needed
			addresses = append(addresses, addr)
		}
	}

	return addresses, nil
}

func (s *HybridSearchService) searchCommentsWithMeilisearch(ctx context.Context, query string, limit, offset int) ([]*entities.Comment, error) {
	options := &meilisearch.SearchOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.searchService.Search(ctx, "comments", query, options)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert search results back to Comment entities
	var comments []*entities.Comment
	for _, hit := range result.Hits {
		if commentMap, ok := hit.(map[string]interface{}); ok {
			comment := &entities.Comment{}
			if _, exists := commentMap["content"]; exists {
				// Note: Comment entity doesn't have a Content field
				// This would need to be added to the Comment entity or handled differently
			}
			// Add more field mappings as needed
			comments = append(comments, comment)
		}
	}

	return comments, nil
}
