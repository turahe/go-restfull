package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createNotificationTables)
}

var createNotificationTables = &Migration{
	Name: "20250831103013_create_notification_tables",
	Up: func() error {
		// Create notifications table
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS notifications (
				"id" UUID NOT NULL PRIMARY KEY,
				"user_id" UUID NOT NULL,
				"type" varchar(50) NOT NULL,
				"title" varchar(255) NOT NULL,
				"message" text NOT NULL,
				"data" JSONB DEFAULT '{}',
				"status" varchar(20) DEFAULT 'unread',
				"priority" varchar(20) DEFAULT 'normal',
				"channels" JSONB DEFAULT '[]',
				"read_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"archived_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "notifications_user_id_foreign" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION,
				CONSTRAINT "notifications_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notifications_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notifications_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "check_notification_type" CHECK (type IN (
					'user_registration', 'password_change', 'email_verification', 'profile_update',
					'system_alert', 'maintenance', 'security_alert',
					'new_post', 'comment_reply', 'comment_rejected', 'mention', 'like',
					'invitation', 'role_change', 'permission_update'
				)),
				CONSTRAINT "check_notification_status" CHECK (status IN ('unread', 'read', 'archived')),
				CONSTRAINT "check_notification_priority" CHECK (priority IN ('low', 'normal', 'high', 'urgent'))
			)
		`)
		if err != nil {
			return err
		}

		// Create notification_templates table
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS notification_templates (
				"id" UUID NOT NULL PRIMARY KEY,
				"name" varchar(100) NOT NULL UNIQUE,
				"type" varchar(50) NOT NULL,
				"title" varchar(255) NOT NULL,
				"message" text NOT NULL,
				"subject" varchar(255),
				"channels" JSONB DEFAULT '[]',
				"priority" varchar(20) DEFAULT 'normal',
				"is_active" boolean DEFAULT true,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "notification_templates_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notification_templates_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notification_templates_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "check_template_notification_type" CHECK (type IN (
					'user_registration', 'password_change', 'email_verification', 'profile_update',
					'system_alert', 'maintenance', 'security_alert',
					'new_post', 'comment_reply', 'comment_rejected', 'mention', 'like',
					'invitation', 'role_change', 'permission_update'
				)),
				CONSTRAINT "check_template_notification_priority" CHECK (priority IN ('low', 'normal', 'high', 'urgent'))
			)
		`)
		if err != nil {
			return err
		}

		// Create notification_preferences table
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS notification_preferences (
				"id" UUID NOT NULL PRIMARY KEY,
				"user_id" UUID NOT NULL,
				"type" varchar(50) NOT NULL,
				"email" boolean DEFAULT true,
				"sms" boolean DEFAULT false,
				"push" boolean DEFAULT true,
				"in_app" boolean DEFAULT true,
				"webhook" boolean DEFAULT false,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "notification_preferences_user_id_foreign" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION,
				CONSTRAINT "notification_preferences_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notification_preferences_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notification_preferences_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "check_preference_notification_type" CHECK (type IN (
					'user_registration', 'password_change', 'email_verification', 'profile_update',
					'system_alert', 'maintenance', 'security_alert',
					'new_post', 'comment_reply', 'comment_rejected', 'mention', 'like',
					'invitation', 'role_change', 'permission_update'
				)),
				UNIQUE("user_id", "type")
			)
		`)
		if err != nil {
			return err
		}

		// Create notification_deliveries table
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS notification_deliveries (
				"id" UUID NOT NULL PRIMARY KEY,
				"notification_id" UUID NOT NULL,
				"channel" varchar(20) NOT NULL,
				"status" varchar(20) DEFAULT 'pending',
				"sent_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"delivered_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"failed_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"error_message" text,
				"retry_count" integer DEFAULT 0,
				"max_retries" integer DEFAULT 3,
				"next_retry_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "notification_deliveries_notification_id_foreign" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON DELETE CASCADE ON UPDATE NO ACTION,
				CONSTRAINT "notification_deliveries_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notification_deliveries_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "notification_deliveries_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "check_delivery_channel" CHECK (channel IN ('email', 'sms', 'push', 'in_app', 'webhook')),
				CONSTRAINT "check_delivery_status" CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'cancelled'))
			)
		`)
		if err != nil {
			return err
		}

		// Create indexes
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
			CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
			CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
			CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
			CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status);
			
			CREATE INDEX IF NOT EXISTS idx_notification_templates_type ON notification_templates(type);
			CREATE INDEX IF NOT EXISTS idx_notification_templates_active ON notification_templates(is_active);
			
			CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
			CREATE INDEX IF NOT EXISTS idx_notification_preferences_type ON notification_preferences(type);
			
			CREATE INDEX IF NOT EXISTS idx_notification_deliveries_notification_id ON notification_deliveries(notification_id);
			CREATE INDEX IF NOT EXISTS idx_notification_deliveries_channel ON notification_deliveries(channel);
			CREATE INDEX IF NOT EXISTS idx_notification_deliveries_status ON notification_deliveries(status);
		`)
		if err != nil {
			return err
		}

		// Insert default notification templates
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			INSERT INTO notification_templates (name, type, title, message, subject, channels, priority) VALUES
			('welcome_user', 'user_registration', 'Welcome to Our Platform!', 'Hi {{username}}, welcome to our platform! We''re excited to have you on board.', 'Welcome to Our Platform', '["email", "in_app"]', 'normal'),
			('password_reset', 'password_change', 'Password Reset Request', 'Hi {{username}}, you requested a password reset. Click the link below to reset your password.', 'Password Reset Request', '["email"]', 'high'),
			('email_verification', 'email_verification', 'Verify Your Email', 'Hi {{username}}, please verify your email address by clicking the link below.', 'Verify Your Email', '["email"]', 'high'),
			('new_post_notification', 'new_post', 'New Post from {{author}}', '{{author}} just published a new post: {{post_title}}', 'New Post Available', '["email", "in_app"]', 'normal'),
			('comment_reply', 'comment_reply', 'Reply to Your Comment', '{{replier}} replied to your comment on {{post_title}}', 'New Reply to Your Comment', '["email", "in_app"]', 'normal'),
			('comment_rejected', 'comment_rejected', 'Comment Rejected', 'Your comment on {{model_type}} has been rejected by a moderator. Please review our community guidelines and try again.', 'Comment Rejected', '["email", "in_app"]', 'normal'),
			('system_maintenance', 'maintenance', 'Scheduled Maintenance', 'We will be performing scheduled maintenance on {{date}} from {{start_time}} to {{end_time}}.', 'Scheduled Maintenance Notice', '["email", "in_app"]', 'high'),
			('security_alert', 'security_alert', 'Security Alert', 'We detected unusual activity on your account. Please review and contact support if needed.', 'Security Alert', '["email", "in_app"]', 'urgent')
			ON CONFLICT (name) DO NOTHING;
		`)
		if err != nil {
			return err
		}

		return nil
	},
	Down: func() error {
		// Drop tables in reverse order (due to foreign key constraints)
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS notification_deliveries;
			DROP TABLE IF EXISTS notification_preferences;
			DROP TABLE IF EXISTS notification_templates;
			DROP TABLE IF EXISTS notifications;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
