-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'unread',
    channels JSONB NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_notifications_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT check_notification_type CHECK (type IN (
        'user_registration', 'password_change', 'email_verification', 'profile_update',
        'system_alert', 'maintenance', 'security_alert',
        'new_post', 'comment_reply', 'mention', 'like',
        'invitation', 'role_change', 'permission_update'
    )),
    CONSTRAINT check_notification_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    CONSTRAINT check_notification_status CHECK (status IN ('unread', 'read', 'archived', 'deleted'))
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_priority ON notifications(priority);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status);
CREATE INDEX IF NOT EXISTS idx_notifications_user_type ON notifications(user_id, type);
CREATE INDEX IF NOT EXISTS idx_notifications_expires_at ON notifications(expires_at) WHERE expires_at IS NOT NULL;

-- Create notification_templates table
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    subject VARCHAR(255),
    channels JSONB NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_notification_templates_type CHECK (type IN (
        'user_registration', 'password_change', 'email_verification', 'profile_update',
        'system_alert', 'maintenance', 'security_alert',
        'new_post', 'comment_reply', 'mention', 'like',
        'invitation', 'role_change', 'permission_update'
    )),
    CONSTRAINT fk_notification_templates_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent'))
);

-- Create indexes for notification_templates
CREATE INDEX IF NOT EXISTS idx_notification_templates_name ON notification_templates(name);
CREATE INDEX IF NOT EXISTS idx_notification_templates_type ON notification_templates(type);
CREATE INDEX IF NOT EXISTS idx_notification_templates_active ON notification_templates(is_active);

-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    email BOOLEAN NOT NULL DEFAULT true,
    sms BOOLEAN NOT NULL DEFAULT false,
    push BOOLEAN NOT NULL DEFAULT true,
    in_app BOOLEAN NOT NULL DEFAULT true,
    webhook BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_notification_preferences_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_notification_preferences_type CHECK (type IN (
        'user_registration', 'password_change', 'email_verification', 'profile_update',
        'system_alert', 'maintenance', 'security_alert',
        'new_post', 'comment_reply', 'mention', 'like',
        'invitation', 'role_change', 'permission_update'
    )),
    UNIQUE(user_id, type)
);

-- Create indexes for notification_preferences
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_type ON notification_preferences(type);

-- Create notification_deliveries table
CREATE TABLE IF NOT EXISTS notification_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_notification_deliveries_notification_id FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    CONSTRAINT fk_notification_deliveries_channel CHECK (channel IN ('email', 'sms', 'push', 'in_app', 'webhook')),
    CONSTRAINT fk_notification_deliveries_status CHECK (status IN ('pending', 'sent', 'failed', 'delivered'))
);

-- Create indexes for notification_deliveries
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_notification_id ON notification_deliveries(notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_channel ON notification_deliveries(channel);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_status ON notification_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_attempts ON notification_deliveries(attempts);

-- Insert default notification templates
INSERT INTO notification_templates (name, type, title, message, subject, channels, priority) VALUES
('welcome_user', 'user_registration', 'Welcome to Our Platform!', 'Hi {{username}}, welcome to our platform! We''re excited to have you on board.', 'Welcome to Our Platform', '["email", "in_app"]', 'normal'),
('password_reset', 'password_change', 'Password Reset Request', 'Hi {{username}}, you requested a password reset. Click the link below to reset your password.', 'Password Reset Request', '["email"]', 'high'),
('email_verification', 'email_verification', 'Verify Your Email', 'Hi {{username}}, please verify your email address by clicking the link below.', 'Verify Your Email', '["email"]', 'high'),
('new_post_notification', 'new_post', 'New Post from {{author}}', '{{author}} just published a new post: {{post_title}}', 'New Post Available', '["email", "in_app"]', 'normal'),
('comment_reply', 'comment_reply', 'Reply to Your Comment', '{{replier}} replied to your comment on {{post_title}}', 'New Reply to Your Comment', '["email", "in_app"]', 'normal'),
('system_maintenance', 'maintenance', 'Scheduled Maintenance', 'We will be performing scheduled maintenance on {{date}} from {{start_time}} to {{end_time}}.', 'Scheduled Maintenance Notice', '["email", "in_app"]', 'high'),
('security_alert', 'security_alert', 'Security Alert', 'We detected unusual activity on your account. Please review and contact support if needed.', 'Security Alert', '["email", "in_app"]', 'urgent')
ON CONFLICT (name) DO NOTHING;
