package dto

type LoginMeta struct {
	DeviceID  string `json:"deviceId"`
	IPAddress string `json:"ipAddress"`
	UserAgent string `json:"userAgent"`
}
