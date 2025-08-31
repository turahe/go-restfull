package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/pagination"
)

// PostgresNotificationRepository implements the notification repository interface
type PostgresNotificationRepository struct {
	db *pgxpool.Pool
}

// NewPostgresNotificationRepository creates a new PostgreSQL notification repository
func NewPostgresNotificationRepository(db *pgxpool.Pool) repositories.NotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

// Create creates a new notification
func (r *PostgresNotificationRepository) Create(ctx context.Context, notification *entities.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, type, title, message, data, priority, status, 
			channels, read_at, archived_at, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	channelsJSON, err := json.Marshal(notification.Channels)
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}

	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		dataJSON,
		notification.Priority,
		notification.Status,
		channelsJSON,
		notification.ReadAt,
		notification.ArchivedAt,
		notification.ExpiresAt,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	return err
}

// GetByID gets a notification by ID
func (r *PostgresNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE id = $1 AND deleted_at IS NULL
	`

	var notification entities.Notification
	var channelsJSON, dataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&dataJSON,
		&notification.Priority,
		&notification.Status,
		&channelsJSON,
		&notification.ReadAt,
		&notification.ArchivedAt,
		&notification.ExpiresAt,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, err
	}

	// Parse JSON fields
	if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
	}

	if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &notification, nil
}

// Update updates a notification
func (r *PostgresNotificationRepository) Update(ctx context.Context, notification *entities.Notification) error {
	query := `
		UPDATE notifications
		SET type = $1, title = $2, message = $3, data = $4, priority = $5,
		    status = $6, channels = $7, read_at = $8, archived_at = $9,
		    expires_at = $10, updated_at = $11
		WHERE id = $12
	`

	channelsJSON, err := json.Marshal(notification.Channels)
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}

	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		notification.Type,
		notification.Title,
		notification.Message,
		dataJSON,
		notification.Priority,
		notification.Status,
		channelsJSON,
		notification.ReadAt,
		notification.ArchivedAt,
		notification.ExpiresAt,
		notification.UpdatedAt,
		notification.ID,
	)

	return err
}

// Delete deletes a notification
func (r *PostgresNotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// SoftDelete soft deletes a notification
func (r *PostgresNotificationRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

// GetByUserID gets notifications for a user with pagination
func (r *PostgresNotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// GetUnreadByUserID gets unread notifications for a user
func (r *PostgresNotificationRepository) GetUnreadByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND status = 'unread' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// GetReadByUserID gets read notifications for a user
func (r *PostgresNotificationRepository) GetReadByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND status = 'read' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// GetArchivedByUserID gets archived notifications for a user
func (r *PostgresNotificationRepository) GetArchivedByUserID(ctx context.Context, userID uuid.UUID, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND status = 'archived' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`
	_, err := r.db.Exec(ctx, query, time.Now(), time.Now(), id, userID)
	return err
}

// MarkAsUnread marks a notification as unread
func (r *PostgresNotificationRepository) MarkAsUnread(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'unread', read_at = NULL, updated_at = $1
		WHERE id = $2 AND user_id = $3
	`
	_, err := r.db.Exec(ctx, query, time.Now(), id, userID)
	return err
}

// Archive archives a notification
func (r *PostgresNotificationRepository) Archive(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'archived', archived_at = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`
	_, err := r.db.Exec(ctx, query, time.Now(), time.Now(), id, userID)
	return err
}

// Unarchive unarchives a notification
func (r *PostgresNotificationRepository) Unarchive(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'unread', archived_at = NULL, updated_at = $1
		WHERE id = $2 AND user_id = $3
	`
	_, err := r.db.Exec(ctx, query, time.Now(), id, userID)
	return err
}

// MarkMultipleAsRead marks multiple notifications as read
func (r *PostgresNotificationRepository) MarkMultipleAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	// Build the query with placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+2)

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	args[len(ids)] = time.Now() // read_at
	args[len(ids)+1] = userID   // user_id

	query := fmt.Sprintf(`
		UPDATE notifications
		SET status = 'read', read_at = $%d, updated_at = $%d
		WHERE id IN (%s) AND user_id = $%d
	`, len(ids)+1, len(ids)+1, strings.Join(placeholders, ","), len(ids)+2)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// MarkMultipleAsUnread marks multiple notifications as unread
func (r *PostgresNotificationRepository) MarkMultipleAsUnread(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	// Build the query with placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+2)

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	args[len(ids)] = time.Now() // updated_at
	args[len(ids)+1] = userID   // user_id

	query := fmt.Sprintf(`
		UPDATE notifications
		SET status = 'unread', read_at = NULL, updated_at = $%d
		WHERE id IN (%s) AND user_id = $%d
	`, len(ids)+1, strings.Join(placeholders, ","), len(ids)+2)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// ArchiveMultiple archives multiple notifications
func (r *PostgresNotificationRepository) ArchiveMultiple(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	// Build the query with placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+2)

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	args[len(ids)] = time.Now() // archived_at
	args[len(ids)+1] = userID   // user_id

	query := fmt.Sprintf(`
		UPDATE notifications
		SET status = 'archived', archived_at = $%d, updated_at = $%d
		WHERE id IN (%s) AND user_id = $%d
	`, len(ids)+1, len(ids)+1, strings.Join(placeholders, ","), len(ids)+2)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// CountByUserID counts notifications for a user
func (r *PostgresNotificationRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// CountUnreadByUserID counts unread notifications for a user
func (r *PostgresNotificationRepository) CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND status = 'unread' AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// GetByType gets notifications by type for a user
func (r *PostgresNotificationRepository) GetByType(ctx context.Context, userID uuid.UUID, notificationType entities.NotificationType, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND type = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, notificationType, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// GetByPriority gets notifications by priority for a user
func (r *PostgresNotificationRepository) GetByPriority(ctx context.Context, userID uuid.UUID, priority entities.NotificationPriority, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND priority = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, priority, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// GetByDateRange gets notifications by date range for a user
func (r *PostgresNotificationRepository) GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate string, pagination *pagination.PaginationRequest) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, priority, status,
		       channels, read_at, archived_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	offset := (pagination.Page - 1) * pagination.PerPage

	rows, err := r.db.Query(ctx, query, userID, startDate, endDate, pagination.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		var channelsJSON, dataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&dataJSON,
			&notification.Priority,
			&notification.Status,
			&channelsJSON,
			&notification.ReadAt,
			&notification.ArchivedAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// DeleteExpired deletes expired notifications
func (r *PostgresNotificationRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM notifications WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	_, err := r.db.Exec(ctx, query)
	return err
}

// DeleteOldArchived deletes old archived notifications
func (r *PostgresNotificationRepository) DeleteOldArchived(ctx context.Context, daysOld int) error {
	query := `DELETE FROM notifications WHERE status = 'archived' AND archived_at < NOW() - INTERVAL '1 day' * $1`
	_, err := r.db.Exec(ctx, query, daysOld)
	return err
}
