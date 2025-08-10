package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/infrastructure/container"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Initialize and manage MeiliSearch indexes",
	Long:  `Initialize and manage MeiliSearch indexes for the application`,
}

var initIndexesCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize all MeiliSearch indexes",
	Long:  `Create and configure all MeiliSearch indexes for the application`,
	Run: func(cmd *cobra.Command, args []string) {
		initIndexes()
	},
}

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Reindex all data into MeiliSearch",
	Long:  `Reindex all data from the database into MeiliSearch indexes`,
	Run: func(cmd *cobra.Command, args []string) {
		reindexAll()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show MeiliSearch index status",
	Long:  `Show the current status of all MeiliSearch indexes`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

func init() {
	searchCmd.AddCommand(initIndexesCmd)
	searchCmd.AddCommand(reindexCmd)
	searchCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(searchCmd)
}

func initIndexes() {
	log.Println("Initializing MeiliSearch indexes...")

	// Setup configuration and database
	SetupAll()

	// Get database connection pool
	db := pgx.GetPgxPool()
	if db == nil {
		log.Fatal("Database connection pool is not available")
	}

	// Create container
	container := container.NewContainer(db)

	// Check if search service is available
	if container.SearchService == nil {
		log.Fatal("Search service is not available. Please check MeiliSearch configuration.")
	}

	// Initialize indexes
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := container.IndexerService.InitializeIndexes(ctx); err != nil {
		log.Fatalf("Failed to initialize indexes: %v", err)
	}

	log.Println("✅ MeiliSearch indexes initialized successfully!")
}

func reindexAll() {
	log.Println("Starting full reindex of all data...")

	// Setup configuration and database
	SetupAll()

	// Get database connection pool
	db := pgx.GetPgxPool()
	if db == nil {
		log.Fatal("Database connection pool is not available")
	}

	// Create container
	container := container.NewContainer(db)

	// Check if search service is available
	if container.SearchService == nil {
		log.Fatal("Search service is not available. Please check MeiliSearch configuration.")
	}

	// Reindex all data
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes timeout
	defer cancel()

	if err := container.IndexerService.IndexAllData(ctx); err != nil {
		log.Fatalf("Failed to reindex data: %v", err)
	}

	log.Println("✅ Full reindex completed successfully!")
}

func showStatus() {
	log.Println("Checking MeiliSearch index status...")

	// Setup configuration and database
	SetupAll()

	// Get database connection pool
	db := pgx.GetPgxPool()
	if db == nil {
		log.Fatal("Database connection pool is not available")
	}

	// Create container
	container := container.NewContainer(db)

	// Check if search service is available
	if container.SearchService == nil {
		log.Println("❌ Search service is not available")
		log.Println("Please check MeiliSearch configuration in config.yaml")
		return
	}

	// Get index status
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status := container.IndexerService.GetIndexStatus(ctx)

	// Display status
	fmt.Println("\n📊 MeiliSearch Index Status")
	fmt.Println("==============================")

	if available, ok := status["meilisearch_available"].(bool); ok && available {
		fmt.Println("✅ MeiliSearch is available")

		if indexes, ok := status["indexes"].(map[string]interface{}); ok {
			fmt.Println("\nIndexes:")
			for name, indexStatus := range indexes {
				if indexMap, ok := indexStatus.(map[string]interface{}); ok {
					if exists, ok := indexMap["exists"].(bool); ok {
						if exists {
							fmt.Printf("  ✅ %s: exists", name)
							if stats, ok := indexMap["stats"]; ok && stats != nil {
								fmt.Printf(" (has stats)")
							}
							fmt.Println()
						} else {
							fmt.Printf("  ❌ %s: does not exist", name)
							if err, ok := indexMap["error"].(string); ok && err != "" {
								fmt.Printf(" (error: %s)", err)
							}
							fmt.Println()
						}
					}
				}
			}
		}
	} else {
		fmt.Println("❌ MeiliSearch is not available")
	}
}
