package services

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// organizationService implements the OrganizationService interface
type organizationService struct {
	organizationRepo repositories.OrganizationRepository
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	organizationRepo repositories.OrganizationRepository,
) ports.OrganizationService {
	return &organizationService{
		organizationRepo: organizationRepo,
	}
}

func (s *organizationService) CreateOrganization(ctx context.Context, name, description, code, email, phone, address, website, logoURL string, parentID *uuid.UUID) (*entities.Organization, error) {
	// Validate code uniqueness if provided
	if code != "" {
		exists, err := s.organizationRepo.ExistsByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("organization with this code already exists")
		}
	}

	// Validate parent exists if provided
	if parentID != nil {
		exists, err := s.organizationRepo.ExistsByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("parent organization not found")
		}
	}

	// Create organization entity
	organization, err := entities.NewOrganization(name, description, code, email, phone, address, website, logoURL, parentID)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.organizationRepo.Create(ctx, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

func (s *organizationService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if organization.IsDeleted() {
		return nil, errors.New("organization not found")
	}
	return organization, nil
}

func (s *organizationService) GetOrganizationByCode(ctx context.Context, code string) (*entities.Organization, error) {
	organization, err := s.organizationRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if organization.IsDeleted() {
		return nil, errors.New("organization not found")
	}
	return organization, nil
}

func (s *organizationService) GetAllOrganizations(ctx context.Context, limit, offset int) ([]*entities.Organization, error) {
	return s.organizationRepo.GetAll(ctx, limit, offset)
}

func (s *organizationService) UpdateOrganization(ctx context.Context, id uuid.UUID, name, description, code, email, phone, address, website, logoURL string) (*entities.Organization, error) {
	// Get existing organization
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if organization.IsDeleted() {
		return nil, errors.New("organization not found")
	}

	// Validate code uniqueness if changed
	if code != "" && (organization.Code == nil || *organization.Code != code) {
		exists, err := s.organizationRepo.ExistsByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("organization with this code already exists")
		}
	}

	// Update organization
	if err := organization.UpdateOrganization(name, description, code, email, phone, address, website, logoURL); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.organizationRepo.Update(ctx, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if organization.IsDeleted() {
		return errors.New("organization not found")
	}

	return s.organizationRepo.Delete(ctx, id)
}

func (s *organizationService) GetRootOrganizations(ctx context.Context) ([]*entities.Organization, error) {
	return s.organizationRepo.GetRoots(ctx)
}

func (s *organizationService) GetChildOrganizations(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	// Validate parent exists
	exists, err := s.organizationRepo.ExistsByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("parent organization not found")
	}

	return s.organizationRepo.GetChildren(ctx, parentID)
}

func (s *organizationService) GetDescendantOrganizations(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error) {
	// Validate parent exists
	exists, err := s.organizationRepo.ExistsByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("parent organization not found")
	}

	return s.organizationRepo.GetDescendants(ctx, parentID)
}

func (s *organizationService) GetAncestorOrganizations(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Validate organization exists
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("organization not found")
	}

	return s.organizationRepo.GetAncestors(ctx, organizationID)
}

func (s *organizationService) GetSiblingOrganizations(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Validate organization exists
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("organization not found")
	}

	return s.organizationRepo.GetSiblings(ctx, organizationID)
}

func (s *organizationService) GetOrganizationPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error) {
	// Validate organization exists
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("organization not found")
	}

	return s.organizationRepo.GetPath(ctx, organizationID)
}

func (s *organizationService) GetOrganizationTree(ctx context.Context) ([]*entities.Organization, error) {
	return s.organizationRepo.GetTree(ctx)
}

func (s *organizationService) GetOrganizationSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error) {
	// Validate root exists
	exists, err := s.organizationRepo.ExistsByID(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("root organization not found")
	}

	return s.organizationRepo.GetSubtree(ctx, rootID)
}

func (s *organizationService) AddChildOrganization(ctx context.Context, parentID, childID uuid.UUID) error {
	// Validate both organizations exist
	parentExists, err := s.organizationRepo.ExistsByID(ctx, parentID)
	if err != nil {
		return err
	}
	if !parentExists {
		return errors.New("parent organization not found")
	}

	childExists, err := s.organizationRepo.ExistsByID(ctx, childID)
	if err != nil {
		return err
	}
	if !childExists {
		return errors.New("child organization not found")
	}

	// Validate hierarchy
	if err := s.ValidateOrganizationHierarchy(ctx, parentID, childID); err != nil {
		return err
	}

	return s.organizationRepo.AddChild(ctx, parentID, childID)
}

func (s *organizationService) MoveOrganization(ctx context.Context, organizationID, newParentID uuid.UUID) error {
	// Validate both organizations exist
	orgExists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if !orgExists {
		return errors.New("organization not found")
	}

	parentExists, err := s.organizationRepo.ExistsByID(ctx, newParentID)
	if err != nil {
		return err
	}
	if !parentExists {
		return errors.New("new parent organization not found")
	}

	// Validate hierarchy
	if err := s.ValidateOrganizationHierarchy(ctx, newParentID, organizationID); err != nil {
		return err
	}

	return s.organizationRepo.MoveSubtree(ctx, organizationID, newParentID)
}

func (s *organizationService) DeleteOrganizationSubtree(ctx context.Context, organizationID uuid.UUID) error {
	// Validate organization exists
	exists, err := s.organizationRepo.ExistsByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("organization not found")
	}

	return s.organizationRepo.DeleteSubtree(ctx, organizationID)
}

func (s *organizationService) SetOrganizationStatus(ctx context.Context, id uuid.UUID, status entities.OrganizationStatus) error {
	organization, err := s.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if organization.IsDeleted() {
		return errors.New("organization not found")
	}

	if err := organization.SetStatus(status); err != nil {
		return err
	}

	return s.organizationRepo.Update(ctx, organization)
}

func (s *organizationService) GetOrganizationsByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error) {
	return s.organizationRepo.GetByStatus(ctx, status, limit, offset)
}

func (s *organizationService) SearchOrganizations(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error) {
	return s.organizationRepo.Search(ctx, query, limit, offset)
}

func (s *organizationService) GetOrganizationsWithPagination(ctx context.Context, page, perPage int, search string, status entities.OrganizationStatus) ([]*entities.Organization, int64, error) {
	limit := perPage
	offset := (page - 1) * perPage

	var organizations []*entities.Organization
	var total int64
	var err error

	if search != "" {
		organizations, err = s.organizationRepo.Search(ctx, search, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.organizationRepo.CountBySearch(ctx, search)
	} else if status != "" {
		organizations, err = s.organizationRepo.GetByStatus(ctx, status, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.organizationRepo.CountByStatus(ctx, status)
	} else {
		organizations, err = s.organizationRepo.GetAll(ctx, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.organizationRepo.Count(ctx)
	}

	if err != nil {
		return nil, 0, err
	}

	return organizations, total, nil
}

func (s *organizationService) ValidateOrganizationHierarchy(ctx context.Context, parentID, childID uuid.UUID) error {
	// Check if child is already a descendant of parent
	isDescendant, err := s.organizationRepo.IsDescendant(ctx, parentID, childID)
	if err != nil {
		return err
	}
	if isDescendant {
		return errors.New("cannot create circular reference in organization hierarchy")
	}

	// Check if parent is a descendant of child
	isAncestor, err := s.organizationRepo.IsAncestor(ctx, childID, parentID)
	if err != nil {
		return err
	}
	if isAncestor {
		return errors.New("cannot create circular reference in organization hierarchy")
	}

	return nil
}

func (s *organizationService) CheckOrganizationExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.organizationRepo.ExistsByID(ctx, id)
}

func (s *organizationService) CheckOrganizationCodeExists(ctx context.Context, code string) (bool, error) {
	return s.organizationRepo.ExistsByCode(ctx, code)
}

func (s *organizationService) GetOrganizationsCount(ctx context.Context, search string, status entities.OrganizationStatus) (int64, error) {
	if search != "" {
		return s.organizationRepo.CountBySearch(ctx, search)
	}
	if status != "" {
		return s.organizationRepo.CountByStatus(ctx, status)
	}
	return s.organizationRepo.Count(ctx)
}

func (s *organizationService) GetChildrenCount(ctx context.Context, parentID uuid.UUID) (int64, error) {
	return s.organizationRepo.CountChildren(ctx, parentID)
}

func (s *organizationService) GetDescendantsCount(ctx context.Context, parentID uuid.UUID) (int64, error) {
	return s.organizationRepo.CountDescendants(ctx, parentID)
}
