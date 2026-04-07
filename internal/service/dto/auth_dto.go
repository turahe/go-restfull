package dto

type TwoFactorSetupResult struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}
