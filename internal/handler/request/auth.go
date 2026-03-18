package request

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=190"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=190"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	DeviceID string `json:"deviceId" binding:"required,min=4,max=64"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required,min=10,max=200"`
	DeviceID     string `json:"deviceId" binding:"required,min=4,max=64"`
}

type ImpersonateRequest struct {
	UserID   uint   `json:"userId" binding:"required,gt=0"`
	Reason   string `json:"reason" binding:"required,min=5,max=255"`
	DeviceID string `json:"deviceId" binding:"required,min=4,max=64"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,min=8,max=72"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=72"`
}

type ChangeEmailRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,min=8,max=72"`
	NewEmail        string `json:"newEmail" binding:"required,email,max=190"`
}

