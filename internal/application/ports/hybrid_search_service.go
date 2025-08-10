package ports

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// HybridSearchService defines the interface for hybrid search operations
// that can use both Meilisearch and SQL fallback
type HybridSearchService interface {
	// Search engine availability
	IsMeilisearchAvailable() bool
	GetSearchEngineInfo() map[string]interface{}

	// Entity-specific search methods
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error)
	SearchOrganizations(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error)
	SearchTags(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error)
	SearchTaxonomies(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	SearchMedia(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error)
	SearchMenus(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error)
	SearchRoles(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error)
	SearchContent(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error)
	SearchAddresses(ctx context.Context, query string, limit, offset int) ([]*entities.Address, error)
	SearchComments(ctx context.Context, query string, limit, offset int) ([]*entities.Comment, error)
}
