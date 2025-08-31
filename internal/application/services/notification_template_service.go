package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/helper/pagination"
)

// NotificationTemplateService implements the notification template service interface
type NotificationTemplateService struct {
	templateRepo repositories.NotificationTemplateRepository
}

// NewNotificationTemplateService creates a new notification template service
func NewNotificationTemplateService(templateRepo repositories.NotificationTemplateRepository) services.NotificationTemplateService {
	return &NotificationTemplateService{
		templateRepo: templateRepo,
	}
}

// CreateTemplate creates a new notification template
func (s *NotificationTemplateService) CreateTemplate(ctx context.Context, template *entities.NotificationTemplate) error {
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.IsActive = true

	return s.templateRepo.Create(ctx, template)
}

// GetTemplateByID gets a template by ID
func (s *NotificationTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*entities.NotificationTemplate, error) {
	return s.templateRepo.GetByID(ctx, id)
}

// GetTemplateByName gets a template by name
func (s *NotificationTemplateService) GetTemplateByName(ctx context.Context, name string) (*entities.NotificationTemplate, error) {
	return s.templateRepo.GetByName(ctx, name)
}

// GetTemplateByType gets a template by notification type
func (s *NotificationTemplateService) GetTemplateByType(ctx context.Context, notificationType entities.NotificationType) (*entities.NotificationTemplate, error) {
	return s.templateRepo.GetByType(ctx, notificationType)
}

// UpdateTemplate updates a template
func (s *NotificationTemplateService) UpdateTemplate(ctx context.Context, template *entities.NotificationTemplate) error {
	template.UpdatedAt = time.Now()
	return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate deletes a template
func (s *NotificationTemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return s.templateRepo.Delete(ctx, id)
}

// GetAllTemplates gets all templates with pagination
func (s *NotificationTemplateService) GetAllTemplates(ctx context.Context, pagination *pagination.PaginationRequest) ([]*entities.NotificationTemplate, error) {
	return s.templateRepo.GetAll(ctx, pagination)
}

// ValidateTemplate validates a notification template
func (s *NotificationTemplateService) ValidateTemplate(ctx context.Context, template *entities.NotificationTemplate) error {
	// Basic validation
	if template.Name == "" {
		return services.ErrInvalidTemplateName
	}

	if template.Title == "" {
		return services.ErrInvalidTemplateTitle
	}

	if template.Message == "" {
		return services.ErrInvalidTemplateMessage
	}

	if len(template.Channels) == 0 {
		return services.ErrInvalidTemplateChannels
	}

	return nil
}

// TestTemplate tests a template with sample data
func (s *NotificationTemplateService) TestTemplate(ctx context.Context, templateID uuid.UUID, testData map[string]interface{}) (string, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return "", err
	}

	// Simple template rendering (you can enhance this with a proper template engine)
	message := template.Message
	for key, value := range testData {
		placeholder := "{{" + key + "}}"
		message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
	}

	return message, nil
}
