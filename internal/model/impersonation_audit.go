package model

import "time"

type ImpersonationAudit struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ImpersonatorID uint      `json:"impersonatorId" gorm:"not null;index"`
	ImpersonatedUserID uint  `json:"impersonatedUserId" gorm:"not null;index"`
	Reason         string    `json:"reason" gorm:"type:varchar(255);not null"`
	IPAddress      string    `json:"ipAddress" gorm:"type:varchar(45);not null"`
	UserAgent      string    `json:"userAgent" gorm:"type:varchar(255);not null"`
	Timestamp      time.Time `json:"timestamp" gorm:"autoCreateTime;index"`
}

