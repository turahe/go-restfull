package seeds

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// SeedNotificationPreferences seeds default notification preferences for all users
func SeedNotificationPreferences(db *pgxpool.Pool) error {
	ctx := context.Background()

	// Get all user IDs
	rows, err := db.Query(ctx, "SELECT id FROM users WHERE deleted_at IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Create default preferences for each user
	for _, userID := range userIDs {
		if err := CreateDefaultNotificationPreferences(ctx, db, userID); err != nil {
			// Log error but continue with other users
			continue
		}
	}

	return nil
}

// CreateDefaultNotificationPreferences creates default notification preferences for a specific user
func CreateDefaultNotificationPreferences(ctx context.Context, db *pgxpool.Pool, userID string) error {
	// Define notification types
	notificationTypes := []entities.NotificationType{
		entities.NotificationTypeUserRegistration,
		entities.NotificationTypePasswordChange,
		entities.NotificationTypeEmailVerification,
		entities.NotificationTypeProfileUpdate,
		entities.NotificationTypeSystemAlert,
		entities.NotificationTypeMaintenance,
		entities.NotificationTypeSecurityAlert,
		entities.NotificationTypeNewPost,
		entities.NotificationTypeCommentReply,
		entities.NotificationTypeCommentRejected,
		entities.NotificationTypeMention,
		entities.NotificationTypeLike,
		entities.NotificationTypeInvitation,
		entities.NotificationTypeRoleChange,
		entities.NotificationTypePermissionUpdate,
	}

	// Create default preferences for each notification type
	for _, notificationType := range notificationTypes {
		if err := CreateUserNotificationPreference(ctx, db, userID, notificationType); err != nil {
			// Log error but continue with other notification types
			continue
		}
	}

	return nil
}

// CreateUserNotificationPreference creates a single notification preference for a user
func CreateUserNotificationPreference(ctx context.Context, db *pgxpool.Pool, userID string, notificationType entities.NotificationType) error {
	// Set default preferences based on notification type
	preferences := getDefaultPreferencesForType(notificationType)

	// Insert or update preference
	query := `
		INSERT INTO notification_preferences (user_id, type, email, sms, push, in_app, webhook)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, type) DO UPDATE SET
			email = EXCLUDED.email,
			sms = EXCLUDED.sms,
			push = EXCLUDED.push,
			in_app = EXCLUDED.in_app,
			webhook = EXCLUDED.webhook,
			updated_at = NOW()
	`

	_, err := db.Exec(ctx, query,
		userID,
		notificationType,
		preferences["email"],
		preferences["sms"],
		preferences["push"],
		preferences["in_app"],
		preferences["webhook"],
	)

	return err
}

// getDefaultPreferencesForType returns default preferences for a specific notification type
func getDefaultPreferencesForType(notificationType entities.NotificationType) map[string]interface{} {
	// Base preferences - most notifications enabled by default
	preferences := map[string]interface{}{
		"email":   true,  // Most notifications via email
		"sms":     false, // SMS disabled by default
		"push":    true,  // Push notifications enabled
		"in_app":  true,  // In-app notifications enabled
		"webhook": false, // Webhooks disabled by default
	}

	// Special cases for certain notification types
	switch notificationType {
	case entities.NotificationTypeSecurityAlert:
		preferences["email"] = true
		preferences["sms"] = true // Enable SMS for security alerts
		preferences["push"] = true
		preferences["in_app"] = true
	case entities.NotificationTypeMaintenance:
		preferences["email"] = true
		preferences["sms"] = false
		preferences["push"] = true
		preferences["in_app"] = true
	case entities.NotificationTypePasswordChange:
		preferences["email"] = true
		preferences["sms"] = false
		preferences["push"] = false
		preferences["in_app"] = true
	case entities.NotificationTypeEmailVerification:
		preferences["email"] = true
		preferences["sms"] = false
		preferences["push"] = false
		preferences["in_app"] = false
	case entities.NotificationTypeCommentRejected:
		preferences["email"] = true
		preferences["sms"] = false
		preferences["push"] = true
		preferences["in_app"] = true
	case entities.NotificationTypeCommentReply:
		preferences["email"] = true
		preferences["sms"] = false
		preferences["push"] = true
		preferences["in_app"] = true
	case entities.NotificationTypeNewPost:
		preferences["email"] = false // Disable email for new posts to avoid spam
		preferences["sms"] = false
		preferences["push"] = true
		preferences["in_app"] = true
	case entities.NotificationTypeMention:
		preferences["email"] = true
		preferences["sms"] = false
		preferences["push"] = true
		preferences["in_app"] = true
	case entities.NotificationTypeLike:
		preferences["email"] = false // Disable email for likes to avoid spam
		preferences["sms"] = false
		preferences["push"] = true
		preferences["in_app"] = true
	}

	return preferences
}

// SeedNotificationPreferencesForUser seeds notification preferences for a specific user
func SeedNotificationPreferencesForUser(db *pgxpool.Pool, userID string) error {
	ctx := context.Background()
	return CreateDefaultNotificationPreferences(ctx, db, userID)
}

// UpdateUserNotificationPreference updates a specific notification preference for a user
func UpdateUserNotificationPreference(ctx context.Context, db *pgxpool.Pool, userID string, notificationType entities.NotificationType, preferences map[string]interface{}) error {
	query := `
		UPDATE notification_preferences 
		SET email = $3, sms = $4, push = $5, in_app = $6, webhook = $7, updated_at = NOW()
		WHERE user_id = $1 AND type = $2
	`

	_, err := db.Exec(ctx, query,
		userID,
		notificationType,
		preferences["email"],
		preferences["sms"],
		preferences["push"],
		preferences["in_app"],
		preferences["webhook"],
	)

	return err
}

// GetUserNotificationPreferences retrieves all notification preferences for a user
func GetUserNotificationPreferences(ctx context.Context, db *pgxpool.Pool, userID string) ([]map[string]interface{}, error) {
	query := `
		SELECT type, email, sms, push, in_app, webhook
		FROM notification_preferences
		WHERE user_id = $1
		ORDER BY type
	`

	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []map[string]interface{}
	for rows.Next() {
		var notificationType string
		var email, sms, push, inApp, webhook bool

		if err := rows.Scan(&notificationType, &email, &sms, &push, &inApp, &webhook); err != nil {
			continue
		}

		preferences = append(preferences, map[string]interface{}{
			"type":    notificationType,
			"email":   email,
			"sms":     sms,
			"push":    push,
			"in_app":  inApp,
			"webhook": webhook,
		})
	}

	return preferences, nil
}
