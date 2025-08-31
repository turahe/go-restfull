package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/services"
)

// NotificationDeliveryService implements the notification delivery service interface
type NotificationDeliveryService struct {
	deliveryRepo repositories.NotificationDeliveryRepository
}

// NewNotificationDeliveryService creates a new notification delivery service
func NewNotificationDeliveryService(deliveryRepo repositories.NotificationDeliveryRepository) services.NotificationDeliveryService {
	return &NotificationDeliveryService{
		deliveryRepo: deliveryRepo,
	}
}

// CreateDelivery creates a new notification delivery
func (s *NotificationDeliveryService) CreateDelivery(ctx context.Context, delivery *entities.NotificationDelivery) error {
	delivery.ID = uuid.New()
	delivery.CreatedAt = time.Now()
	delivery.UpdatedAt = time.Now()
	delivery.Status = "pending"
	delivery.Attempts = 0

	return s.deliveryRepo.Create(ctx, delivery)
}

// GetDeliveryByID gets a delivery by ID
func (s *NotificationDeliveryService) GetDeliveryByID(ctx context.Context, id uuid.UUID) (*entities.NotificationDelivery, error) {
	return s.deliveryRepo.GetByID(ctx, id)
}

// GetDeliveriesByNotificationID gets all deliveries for a notification
func (s *NotificationDeliveryService) GetDeliveriesByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]*entities.NotificationDelivery, error) {
	return s.deliveryRepo.GetByNotificationID(ctx, notificationID)
}

// UpdateDelivery updates a delivery
func (s *NotificationDeliveryService) UpdateDelivery(ctx context.Context, delivery *entities.NotificationDelivery) error {
	delivery.UpdatedAt = time.Now()
	return s.deliveryRepo.Update(ctx, delivery)
}

// DeleteDelivery deletes a delivery
func (s *NotificationDeliveryService) DeleteDelivery(ctx context.Context, id uuid.UUID) error {
	return s.deliveryRepo.Delete(ctx, id)
}

// MarkAsDelivered marks a delivery as delivered
func (s *NotificationDeliveryService) MarkAsDelivered(ctx context.Context, id uuid.UUID, deliveredAt string) error {
	return s.deliveryRepo.MarkAsDelivered(ctx, id, deliveredAt)
}

// MarkAsFailed marks a delivery as failed
func (s *NotificationDeliveryService) MarkAsFailed(ctx context.Context, id uuid.UUID, error string) error {
	return s.deliveryRepo.MarkAsFailed(ctx, id, error)
}

// IncrementDeliveryAttempts increments the delivery attempts
func (s *NotificationDeliveryService) IncrementDeliveryAttempts(ctx context.Context, id uuid.UUID) error {
	return s.deliveryRepo.IncrementAttempts(ctx, id)
}

// ProcessPendingDeliveries processes pending deliveries
func (s *NotificationDeliveryService) ProcessPendingDeliveries(ctx context.Context) error {
	// This is a placeholder implementation
	// In a real application, you would:
	// 1. Query for pending deliveries
	// 2. Process them based on their channel (email, SMS, push, etc.)
	// 3. Update their status accordingly
	return nil
}

// RetryFailedDeliveries retries failed deliveries
func (s *NotificationDeliveryService) RetryFailedDeliveries(ctx context.Context, maxAttempts int) error {
	// This is a placeholder implementation
	// In a real application, you would:
	// 1. Query for failed deliveries with attempts < maxAttempts
	// 2. Reset their status to pending
	// 3. Queue them for retry
	return nil
}

// CleanupOldDeliveries cleans up old delivery records
func (s *NotificationDeliveryService) CleanupOldDeliveries(ctx context.Context, daysOld int) error {
	return s.deliveryRepo.DeleteOldDeliveries(ctx, daysOld)
}
