package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostSeeder seeds initial posts
type PostSeeder struct{}

// NewPostSeeder creates a new post seeder
func NewPostSeeder() *PostSeeder {
	return &PostSeeder{}
}

// GetName returns the seeder name
func (ps *PostSeeder) GetName() string {
	return "PostSeeder"
}

// Run executes the post seeder
func (ps *PostSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	// Get author user ID
	var authorID uuid.UUID
	err := db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'author'`).Scan(&authorID)
	if err != nil {
		return err
	}

	posts := []struct {
		ID          uuid.UUID
		Title       string
		Slug        string
		Excerpt     string
		Content     string
		AuthorID    uuid.UUID
		Status      string
		PublishedAt *time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}{
		{
			ID:          uuid.New(),
			Title:       "Getting Started with Go Programming",
			Slug:        "getting-started-with-go-programming",
			Excerpt:     "Learn the basics of Go programming language and start building your first application.",
			Content:     "Go is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. Go is syntactically similar to C, but with memory safety, garbage collection, structural typing, and CSP-style concurrency.\n\nIn this post, we'll explore the fundamentals of Go programming and create a simple web application.",
			AuthorID:    authorID,
			Status:      "published",
			PublishedAt: &time.Time{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Building RESTful APIs with Fiber",
			Slug:        "building-restful-apis-with-fiber",
			Excerpt:     "Discover how to build fast and efficient RESTful APIs using the Fiber web framework for Go.",
			Content:     "Fiber is an Express inspired web framework built on top of Fasthttp, the fastest HTTP engine for Go. Designed to ease things up for fast development with zero memory allocation and performance in mind.\n\nIn this tutorial, we'll build a complete RESTful API with authentication, database integration, and proper error handling.",
			AuthorID:    authorID,
			Status:      "published",
			PublishedAt: &time.Time{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Database Design Best Practices",
			Slug:        "database-design-best-practices",
			Excerpt:     "Learn essential database design principles and best practices for scalable applications.",
			Content:     "Good database design is crucial for the performance, scalability, and maintainability of your application. In this comprehensive guide, we'll cover normalization, indexing strategies, and performance optimization techniques.\n\nWe'll also discuss when to use different types of databases and how to design for both read and write performance.",
			AuthorID:    authorID,
			Status:      "published",
			PublishedAt: &time.Time{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Microservices Architecture Patterns",
			Slug:        "microservices-architecture-patterns",
			Excerpt:     "Explore common patterns and anti-patterns in microservices architecture.",
			Content:     "Microservices architecture has become increasingly popular for building scalable and maintainable applications. However, it comes with its own set of challenges and patterns.\n\nIn this article, we'll discuss service discovery, API gateways, circuit breakers, and other essential patterns for successful microservices implementation.",
			AuthorID:    authorID,
			Status:      "draft",
			PublishedAt: nil,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Docker and Kubernetes for Developers",
			Slug:        "docker-and-kubernetes-for-developers",
			Excerpt:     "A practical guide to containerization and orchestration for modern applications.",
			Content:     "Docker and Kubernetes have revolutionized how we deploy and manage applications. Docker provides containerization, while Kubernetes offers orchestration capabilities.\n\nThis guide will walk you through creating Docker images, writing Dockerfiles, and deploying applications to Kubernetes clusters.",
			AuthorID:    authorID,
			Status:      "published",
			PublishedAt: &time.Time{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, post := range posts {
		// Set published time for published posts
		if post.Status == "published" {
			now := time.Now()
			post.PublishedAt = &now
		}

		_, err := db.Exec(ctx, `
			INSERT INTO posts (id, title, slug, excerpt, content, author_id, status, published_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (slug) DO NOTHING
		`, post.ID, post.Title, post.Slug, post.Excerpt, post.Content, post.AuthorID, post.Status, post.PublishedAt, post.CreatedAt, post.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
