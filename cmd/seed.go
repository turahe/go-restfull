package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/turahe/go-restfull/internal/db/seeds"
)

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: "seed", Title: "Seed:"})
	rootCmd.AddCommand(seedCommand)
	rootCmd.AddCommand(seedSpecificCommand)
	rootCmd.AddCommand(seedAdminCommand)
	rootCmd.AddCommand(seedMenusCommand)
	rootCmd.AddCommand(seedOrganizationsCommand)
}

var seedCommand = &cobra.Command{
	Use:     "seed",
	Short:   "Seed the database with initial data",
	GroupID: "seed",
	Run: func(cmd *cobra.Command, args []string) {
		// Setup all the required dependencies
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		log.Println("Starting database seeding...")

		// Run all seeders by default
		if err := seeds.RunAllSeeders(); err != nil {
			log.Printf("Error during seeding: %v", err)
			return
		}

		log.Println("Database seeding completed successfully!")
	},
}

var seedSpecificCommand = &cobra.Command{
	Use:     "seed:roles",
	Short:   "Seed only roles",
	GroupID: "seed",
	Run: func(cmd *cobra.Command, args []string) {
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		if err := seeds.SeedRoles(); err != nil {
			log.Printf("Error seeding roles: %v", err)
			return
		}
		log.Println("Roles seeded successfully!")
	},
}

var seedAdminCommand = &cobra.Command{
	Use:     "seed:admin",
	Short:   "Seed admin user with roles and menus",
	GroupID: "seed",
	Run: func(cmd *cobra.Command, args []string) {
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		if err := seeds.SeedAdminUser(); err != nil {
			log.Printf("Error seeding admin user: %v", err)
			return
		}
		log.Println("Admin user seeded successfully!")
	},
}

var seedMenusCommand = &cobra.Command{
	Use:     "seed:menus",
	Short:   "Seed only menus",
	GroupID: "seed",
	Run: func(cmd *cobra.Command, args []string) {
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		if err := seeds.SeedMenus(); err != nil {
			log.Printf("Error seeding menus: %v", err)
			return
		}
		log.Println("Menus seeded successfully!")
	},
}

var seedOrganizationsCommand = &cobra.Command{
	Use:     "seed:organizations",
	Short:   "Seed only organizations",
	GroupID: "seed",
	Run: func(cmd *cobra.Command, args []string) {
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		if err := seeds.SeedOrganizations(); err != nil {
			log.Printf("Error seeding organizations: %v", err)
			return
		}
		log.Println("Organizations seeded successfully!")
	},
}
