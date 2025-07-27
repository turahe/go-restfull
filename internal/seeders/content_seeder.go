package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ContentSeeder seeds initial content
type ContentSeeder struct{}

// NewContentSeeder creates a new content seeder
func NewContentSeeder() *ContentSeeder {
	return &ContentSeeder{}
}

// GetName returns the seeder name
func (cs *ContentSeeder) GetName() string {
	return "ContentSeeder"
}

// Run executes the content seeder
func (cs *ContentSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	contents := []struct {
		ID        uuid.UUID
		Title     string
		Content   string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		{
			ID:        uuid.New(),
			Title:     "About Us",
			Content:   "We are a team of passionate developers and technology enthusiasts dedicated to sharing knowledge and building amazing applications. Our mission is to provide high-quality content and tools that help developers grow and succeed in their careers.",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Title:     "Privacy Policy",
			Content:   "This privacy policy describes how we collect, use, and protect your personal information. We are committed to protecting your privacy and ensuring the security of your data. We only collect information that is necessary to provide our services and improve your experience.",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Title:     "Terms of Service",
			Content:   "By using our services, you agree to these terms of service. We provide our services as-is and make no warranties about their availability or reliability. You are responsible for your use of our services and any content you create or share.",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Title:     "Contact Information",
			Content:   "You can reach us at contact@example.com or through our contact form. We typically respond to inquiries within 24 hours during business days. For technical support, please include relevant details about your issue.",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, content := range contents {
		_, err := db.Exec(ctx, `
			INSERT INTO contents (id, title, content, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (title) DO NOTHING
		`, content.ID, content.Title, content.Content, content.CreatedAt, content.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
