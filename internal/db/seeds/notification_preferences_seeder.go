package seeds

import (
	"context"
	"encoding/json"

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
		entities.NotificationTypeMention,
		entities.NotificationTypeLike,
		entities.NotificationTypeInvitation,
		entities.NotificationTypeRoleChange,
		entities.NotificationTypePermissionUpdate,
	}

	// Create default preferences for each user and notification type
	for _, userID := range userIDs {
		for _, notificationType := range notificationTypes {
			// Set default preferences based on notification type
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
			}

			_, err = json.Marshal(preferences)
			if err != nil {
				continue
			}

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

			_, err = db.Exec(ctx, query,
				userID,
				notificationType,
				preferences["email"],
				preferences["sms"],
				preferences["push"],
				preferences["in_app"],
				preferences["webhook"],
			)

			if err != nil {
				// Log error but continue with other preferences
				continue
			}
		}
	}

	return nil
}
