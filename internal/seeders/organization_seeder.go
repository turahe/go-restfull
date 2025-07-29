package seeders

import (
	"context"
	"fmt"
	"log"

	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"

	"github.com/google/uuid"
)

// OrganizationSeeder handles seeding of organization data
type OrganizationSeeder struct {
	organizationRepo repositories.OrganizationRepository
}

// NewOrganizationSeeder creates a new organization seeder
func NewOrganizationSeeder(organizationRepo repositories.OrganizationRepository) *OrganizationSeeder {
	return &OrganizationSeeder{
		organizationRepo: organizationRepo,
	}
}

// Seed runs the organization seeder
func (s *OrganizationSeeder) Seed(ctx context.Context) error {
	log.Println("Starting organization seeder...")

	// Sample organization data
	organizations := []struct {
		name        string
		description string
		code        string
		email       string
		phone       string
		address     string
		website     string
		logoURL     string
		parentCode  string // Reference to parent organization code
	}{
		{
			name:        "TechCorp Global",
			description: "Global technology corporation with multiple divisions",
			code:        "TECH-CORP",
			email:       "info@techcorp.com",
			phone:       "+1-555-0100",
			address:     "123 Tech Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com",
			logoURL:     "https://techcorp.com/logo.png",
			parentCode:  "",
		},
		{
			name:        "Software Development Division",
			description: "Main software development division",
			code:        "SOFT-DEV",
			email:       "dev@techcorp.com",
			phone:       "+1-555-0101",
			address:     "456 Dev Avenue, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/dev",
			logoURL:     "https://techcorp.com/dev-logo.png",
			parentCode:  "TECH-CORP",
		},
		{
			name:        "Web Development Team",
			description: "Specialized web development team",
			code:        "WEB-DEV",
			email:       "web@techcorp.com",
			phone:       "+1-555-0102",
			address:     "789 Web Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/web",
			logoURL:     "https://techcorp.com/web-logo.png",
			parentCode:  "SOFT-DEV",
		},
		{
			name:        "Mobile Development Team",
			description: "Specialized mobile development team",
			code:        "MOBILE-DEV",
			email:       "mobile@techcorp.com",
			phone:       "+1-555-0103",
			address:     "321 Mobile Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/mobile",
			logoURL:     "https://techcorp.com/mobile-logo.png",
			parentCode:  "SOFT-DEV",
		},
		{
			name:        "Marketing Division",
			description: "Marketing and communications division",
			code:        "MARKETING",
			email:       "marketing@techcorp.com",
			phone:       "+1-555-0104",
			address:     "654 Marketing Avenue, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/marketing",
			logoURL:     "https://techcorp.com/marketing-logo.png",
			parentCode:  "TECH-CORP",
		},
		{
			name:        "Digital Marketing Team",
			description: "Digital marketing and social media team",
			code:        "DIGITAL-MARKETING",
			email:       "digital@techcorp.com",
			phone:       "+1-555-0105",
			address:     "987 Digital Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/digital",
			logoURL:     "https://techcorp.com/digital-logo.png",
			parentCode:  "MARKETING",
		},
		{
			name:        "Sales Division",
			description: "Sales and business development division",
			code:        "SALES",
			email:       "sales@techcorp.com",
			phone:       "+1-555-0106",
			address:     "147 Sales Boulevard, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/sales",
			logoURL:     "https://techcorp.com/sales-logo.png",
			parentCode:  "TECH-CORP",
		},
		{
			name:        "Enterprise Sales Team",
			description: "Enterprise sales and account management team",
			code:        "ENTERPRISE-SALES",
			email:       "enterprise@techcorp.com",
			phone:       "+1-555-0107",
			address:     "258 Enterprise Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/enterprise",
			logoURL:     "https://techcorp.com/enterprise-logo.png",
			parentCode:  "SALES",
		},
		{
			name:        "SMB Sales Team",
			description: "Small and medium business sales team",
			code:        "SMB-SALES",
			email:       "smb@techcorp.com",
			phone:       "+1-555-0108",
			address:     "369 SMB Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/smb",
			logoURL:     "https://techcorp.com/smb-logo.png",
			parentCode:  "SALES",
		},
		{
			name:        "HR Division",
			description: "Human resources and talent management division",
			code:        "HR",
			email:       "hr@techcorp.com",
			phone:       "+1-555-0109",
			address:     "741 HR Avenue, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/hr",
			logoURL:     "https://techcorp.com/hr-logo.png",
			parentCode:  "TECH-CORP",
		},
		{
			name:        "Recruitment Team",
			description: "Talent acquisition and recruitment team",
			code:        "RECRUITMENT",
			email:       "recruitment@techcorp.com",
			phone:       "+1-555-0110",
			address:     "852 Recruitment Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/recruitment",
			logoURL:     "https://techcorp.com/recruitment-logo.png",
			parentCode:  "HR",
		},
		{
			name:        "Employee Relations Team",
			description: "Employee relations and engagement team",
			code:        "EMPLOYEE-RELATIONS",
			email:       "relations@techcorp.com",
			phone:       "+1-555-0111",
			address:     "963 Relations Street, Silicon Valley, CA 94025",
			website:     "https://techcorp.com/relations",
			logoURL:     "https://techcorp.com/relations-logo.png",
			parentCode:  "HR",
		},
	}

	// Create a map to store organization codes to IDs for parent references
	orgCodeToID := make(map[string]string)

	// Create organizations
	for _, orgData := range organizations {
		// Check if organization already exists
		exists, err := s.organizationRepo.ExistsByCode(ctx, orgData.code)
		if err != nil {
			return fmt.Errorf("failed to check organization existence: %w", err)
		}
		if exists {
			log.Printf("Organization with code %s already exists, skipping...", orgData.code)
			continue
		}

		// Get parent ID if parent code is provided
		var parentID *string
		if orgData.parentCode != "" {
			if parentUUID, exists := orgCodeToID[orgData.parentCode]; exists {
				parentID = &parentUUID
			} else {
				log.Printf("Warning: Parent organization with code %s not found for %s", orgData.parentCode, orgData.code)
			}
		}

		// Create organization entity
		organization, err := entities.NewOrganization(
			orgData.name,
			orgData.description,
			orgData.code,
			orgData.email,
			orgData.phone,
			orgData.address,
			orgData.website,
			orgData.logoURL,
			nil, // We'll handle parent relationship after creation
		)
		if err != nil {
			return fmt.Errorf("failed to create organization entity: %w", err)
		}

		// Save organization
		err = s.organizationRepo.Create(ctx, organization)
		if err != nil {
			return fmt.Errorf("failed to create organization %s: %w", orgData.code, err)
		}

		// Store the organization ID for parent references
		orgCodeToID[orgData.code] = organization.ID.String()

		// Set parent relationship if parent exists
		if parentID != nil {
			parentUUID, err := uuid.Parse(*parentID)
			if err != nil {
				log.Printf("Warning: Failed to parse parent UUID for %s: %v", orgData.code, err)
				continue
			}

			organization.SetParent(&parentUUID)
			err = s.organizationRepo.Update(ctx, organization)
			if err != nil {
				log.Printf("Warning: Failed to update parent relationship for %s: %v", orgData.code, err)
			}
		}

		log.Printf("Created organization: %s (%s)", orgData.name, orgData.code)
	}

	log.Printf("Organization seeder completed. Created %d organizations.", len(organizations))
	return nil
}

// Cleanup removes all seeded organization data
func (s *OrganizationSeeder) Cleanup(ctx context.Context) error {
	log.Println("Cleaning up organization data...")

	// Get all organizations
	organizations, err := s.organizationRepo.GetAll(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get organizations for cleanup: %w", err)
	}

	// Delete organizations in reverse order (children first)
	for i := len(organizations) - 1; i >= 0; i-- {
		org := organizations[i]
		err := s.organizationRepo.Delete(ctx, org.ID)
		if err != nil {
			log.Printf("Warning: Failed to delete organization %s: %v", org.Name, err)
		} else {
			log.Printf("Deleted organization: %s", org.Name)
		}
	}

	log.Println("Organization cleanup completed.")
	return nil
}
