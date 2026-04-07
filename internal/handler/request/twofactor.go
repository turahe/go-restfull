package request

type TwoFASetupRequest struct {
	// currently empty; reserved for future (e.g. delivery method)
}

type TwoFAEnableRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type TwoFAVerifyRequest struct {
	ChallengeID string `json:"challengeId" binding:"required,len=36"`
	Code        string `json:"code" binding:"required,len=6"`
	DeviceID    string `json:"deviceId" binding:"required,min=4,max=64"`
}
