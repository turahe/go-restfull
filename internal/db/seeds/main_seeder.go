package seeds

import (
	"log"
)

// RunAllSeeders runs all seeders in the correct dependency order
func RunAllSeeders() error {
	log.Println("Starting database seeding process...")

	// 1. Seed roles first (required for other seeders)
	log.Println("Seeding roles...")
	if err := SeedRoles(); err != nil {
		return err
	}

	// 2. Seed menus (required for admin user and organizations)
	log.Println("Seeding menus...")
	if err := SeedMenus(); err != nil {
		return err
	}

	// 3. Seed organizations (can depend on roles and menus)
	log.Println("Seeding organizations...")
	if err := SeedOrganizations(); err != nil {
		return err
	}

	// 4. Seed admin user (depends on roles and menus)
	log.Println("Seeding admin user...")
	if err := SeedAdminUser(); err != nil {
		return err
	}

	log.Println("All seeders completed successfully!")
	return nil
}

// RunSpecificSeeder runs a specific seeder by name
func RunSpecificSeeder(seederName string) error {
	switch seederName {
	case "roles":
		return SeedRoles()
	case "menus":
		return SeedMenus()
	case "admin":
		return SeedAdminUser()
	case "all":
		return RunAllSeeders()
	default:
		log.Printf("Unknown seeder: %s", seederName)
		log.Println("Available seeders: roles, menus, admin, all")
		return nil
	}
}
