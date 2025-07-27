package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TaxonomySeeder seeds initial taxonomies
type TaxonomySeeder struct{}

// NewTaxonomySeeder creates a new taxonomy seeder
func NewTaxonomySeeder() *TaxonomySeeder {
	return &TaxonomySeeder{}
}

// GetName returns the seeder name
func (ts *TaxonomySeeder) GetName() string {
	return "TaxonomySeeder"
}

// Run executes the taxonomy seeder
func (ts *TaxonomySeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	taxonomies := []struct {
		ID             uuid.UUID
		Name           string
		Slug           string
		Code           string
		Description    string
		ParentID       *uuid.UUID
		RecordLeft     int64
		RecordRight    int64
		RecordOrdering int64
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}{
		{
			ID:             uuid.New(),
			Name:           "Technology",
			Slug:           "technology",
			Code:           "tech",
			Description:    "Technology related content",
			ParentID:       nil,
			RecordLeft:     1,
			RecordRight:    8,
			RecordOrdering: 0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Name:           "Programming",
			Slug:           "programming",
			Code:           "prog",
			Description:    "Programming languages and development",
			ParentID:       nil,
			RecordLeft:     2,
			RecordRight:    5,
			RecordOrdering: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Name:           "Web Development",
			Slug:           "web-development",
			Code:           "web",
			Description:    "Web development technologies",
			ParentID:       nil,
			RecordLeft:     3,
			RecordRight:    4,
			RecordOrdering: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Name:           "Mobile Development",
			Slug:           "mobile-development",
			Code:           "mobile",
			Description:    "Mobile app development",
			ParentID:       nil,
			RecordLeft:     6,
			RecordRight:    7,
			RecordOrdering: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Name:           "Business",
			Slug:           "business",
			Code:           "biz",
			Description:    "Business and entrepreneurship",
			ParentID:       nil,
			RecordLeft:     9,
			RecordRight:    12,
			RecordOrdering: 0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Name:           "Marketing",
			Slug:           "marketing",
			Code:           "mkt",
			Description:    "Digital marketing strategies",
			ParentID:       nil,
			RecordLeft:     10,
			RecordRight:    11,
			RecordOrdering: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, taxonomy := range taxonomies {
		_, err := db.Exec(ctx, `
			INSERT INTO taxonomies (id, name, slug, code, description, parent_id, record_left, record_right, record_ordering, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (slug) DO NOTHING
		`, taxonomy.ID, taxonomy.Name, taxonomy.Slug, taxonomy.Code, taxonomy.Description, taxonomy.ParentID, taxonomy.RecordLeft, taxonomy.RecordRight, taxonomy.RecordOrdering, taxonomy.CreatedAt, taxonomy.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
