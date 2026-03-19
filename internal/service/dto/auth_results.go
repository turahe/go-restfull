package dto

import "time"

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
