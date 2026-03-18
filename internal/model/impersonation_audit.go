package model

import "time"

type ImpersonationAudit struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ImpersonatorID uint      `json:"impersonator_id" gorm:"not null;index"`
	ImpersonatedUserID uint  `json:"impersonated_user_id" gorm:"not null;index"`
	Reason         string    `json:"reason" gorm:"type:varchar(255);not null"`
	IPAddress      string    `json:"ip_address" gorm:"type:varchar(45);not null"`
	UserAgent      string    `json:"user_agent" gorm:"type:varchar(255);not null"`
	Timestamp      time.Time `json:"timestamp" gorm:"autoCreateTime;index"`
}

