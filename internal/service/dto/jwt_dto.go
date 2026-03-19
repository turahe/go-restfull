package dto

import "github.com/golang-jwt/jwt/v5"

type AccessClaims struct {
	jwt.RegisteredClaims

	UserID      uint     `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	DeviceID    string   `json:"device_id"`

	Impersonation       bool   `json:"impersonation,omitempty"`
	ImpersonatedUserID  *uint  `json:"impersonated_user_id,omitempty"`
	ImpersonatorID      *uint  `json:"impersonator_id,omitempty"`
	ImpersonationReason string `json:"impersonation_reason,omitempty"`
}

