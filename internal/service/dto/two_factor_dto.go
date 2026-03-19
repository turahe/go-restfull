package dto

type SetupResult struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}

