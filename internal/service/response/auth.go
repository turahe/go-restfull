package svcresp

import "time"

type LoginMeta struct {
	DeviceID  string `json:"deviceId"`
	IPAddress string `json:"ipAddress"`
	UserAgent string `json:"userAgent"`
}

type AuthUser struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type LoginResult struct {
	TwoFactorRequired bool      `json:"twoFactorRequired"`
	ChallengeID       string    `json:"challengeId,omitempty"`
	AccessToken       string    `json:"accessToken,omitempty"`
	RefreshToken      string    `json:"refreshToken,omitempty"`
	ExpiresAt         time.Time `json:"expiresAt"`
	SessionID         string    `json:"sessionId"`
	User              AuthUser  `json:"user"`
}

type RefreshResult struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionID    string    `json:"sessionId"`
}

type ImpersonationResult struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

