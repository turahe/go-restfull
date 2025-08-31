package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/services"
)

// NotificationPreferenceService implements the notification preference service interface
type NotificationPreferenceService struct {
	preferenceRepo repositories.NotificationPreferenceRepository
}

// NewNotificationPreferenceService creates a new notification preference service
func NewNotificationPreferenceService(preferenceRepo repositories.NotificationPreferenceRepository) services.NotificationPreferenceService {
	return &NotificationPreferenceService{
		preferenceRepo: preferenceRepo,
	}
}

// CreatePreference creates a new notification preference
func (s *NotificationPreferenceService) CreatePreference(ctx context.Context, preference *entities.NotificationPreference) error {
	preference.ID = uuid.New()
	preference.CreatedAt = time.Now()
	preference.UpdatedAt = time.Now()

	return s.preferenceRepo.Create(ctx, preference)
}

// GetPreferenceByID gets a preference by ID
func (s *NotificationPreferenceService) GetPreferenceByID(ctx context.Context, id uuid.UUID) (*entities.NotificationPreference, error) {
	return s.preferenceRepo.GetByID(ctx, id)
}

// GetUserPreferenceByType gets a user's preference for a specific notification type
func (s *NotificationPreferenceService) GetUserPreferenceByType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType) (*entities.NotificationPreference, error) {
	return s.preferenceRepo.GetByUserIDAndType(ctx, userID, notificationType)
}

// GetAllUserPreferences gets all preferences for a user
func (s *NotificationPreferenceService) GetAllUserPreferences(ctx context.Context, userID uuid.UUID) ([]*entities.NotificationPreference, error) {
	return s.preferenceRepo.GetAllByUserID(ctx, userID)
}

// UpdatePreference updates a preference
func (s *NotificationPreferenceService) UpdatePreference(ctx context.Context, preference *entities.NotificationPreference) error {
	preference.UpdatedAt = time.Now()
	return s.preferenceRepo.Update(ctx, preference)
}

// DeletePreference deletes a preference
func (s *NotificationPreferenceService) DeletePreference(ctx context.Context, id uuid.UUID) error {
	return s.preferenceRepo.Delete(ctx, id)
}

// DeleteUserPreferences deletes all preferences for a user
func (s *NotificationPreferenceService) DeleteUserPreferences(ctx context.Context, userID uuid.UUID) error {
	return s.preferenceRepo.DeleteByUserID(ctx, userID)
}

// CreateDefaultUserPreferences creates default preferences for a user
func (s *NotificationPreferenceService) CreateDefaultUserPreferences(ctx context.Context, userID uuid.UUID) error {
	return s.preferenceRepo.CreateDefaultPreferences(ctx, userID)
}

// UpdateMultipleUserPreferences updates multiple preferences
func (s *NotificationPreferenceService) UpdateMultipleUserPreferences(ctx context.Context, preferences []*entities.NotificationPreference) error {
	return s.preferenceRepo.UpdateMultiplePreferences(ctx, preferences)
}

// ValidatePreferences validates notification preferences
func (s *NotificationPreferenceService) ValidatePreferences(ctx context.Context, preferences []*entities.NotificationPreference) error {
	for _, preference := range preferences {
		// Basic validation
		if preference.UserID == uuid.Nil {
			return services.ErrInvalidUserID
		}

		if preference.Type == "" {
			return services.ErrInvalidNotificationType
		}
	}

	return nil
}
